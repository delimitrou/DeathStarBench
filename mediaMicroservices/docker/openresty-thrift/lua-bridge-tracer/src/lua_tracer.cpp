#include "lua_tracer.h"

#include "carrier.h"
#include "dynamic_tracer.h"
#include "lua_span.h"
#include "lua_span_context.h"
#include "utility.h"

#include <opentracing/dynamic_load.h>

#include <cstdint>
#include <sstream>
#include <stdexcept>

#define METATABLE "lua_opentracing_bridge.tracer"

namespace lua_bridge_tracer {
//------------------------------------------------------------------------------
// check_lua_tracer
//------------------------------------------------------------------------------
static LuaTracer* check_lua_tracer(lua_State* L) noexcept {
  void* user_data = luaL_checkudata(L, 1, METATABLE);
  luaL_argcheck(L, user_data != NULL, 1, "`" METATABLE "' expected");
  return *static_cast<LuaTracer**>(user_data);
}

//------------------------------------------------------------------------------
// get_reference_type
//------------------------------------------------------------------------------
static opentracing::SpanReferenceType get_reference_type(lua_State* L) {
  if (lua_type(L, -1) != LUA_TSTRING) {
    throw std::runtime_error{"reference_type must be string"};
  }
  auto reference_type = opentracing::string_view{lua_tostring(L, -1)};
  if (reference_type == "child_of" || reference_type == "CHILD_OF") {
    return opentracing::SpanReferenceType::ChildOfRef;
  }
  if (reference_type == "follows_from" || reference_type == "FOLLOWS_FROM") {
    return opentracing::SpanReferenceType::FollowsFromRef;
  }
  throw std::runtime_error{"invalid reference type: " +
                           std::string{reference_type}};
}

//------------------------------------------------------------------------------
// get_span_context
//------------------------------------------------------------------------------
static const opentracing::SpanContext& get_span_context(lua_State* L,
                                                        int index) {
  void* user_data =
      luaL_checkudata(L, index, LuaSpanContext::description.metatable);
  if (user_data == nullptr) {
    throw std::runtime_error{
        "span_context must be of type " +
        std::string{LuaSpanContext::description.metatable}};
  }

  auto span_context = *static_cast<LuaSpanContext**>(user_data);
  return span_context->span_context();
}

//------------------------------------------------------------------------------
// get_reference
//------------------------------------------------------------------------------
static std::pair<opentracing::SpanReferenceType,
                 const opentracing::SpanContext*>
get_reference(lua_State* L) {
  switch (lua_type(L, -1)) {
    case LUA_TTABLE:
      break;
    default:
      throw std::runtime_error{"reference must be a table"};
  }

  auto table_len = get_table_len(L, -1);

  // Could be a nil reference so ignore
  if (table_len == 1) {
    return {opentracing::SpanReferenceType::ChildOfRef, nullptr};
  }

  if (table_len != 2) {
    throw std::runtime_error{"reference must contain 2 elements"};
  }

  lua_pushinteger(L, 1);
  lua_gettable(L, -2);
  auto reference_type = get_reference_type(L);
  lua_pop(L, 1);

  lua_pushinteger(L, 2);
  lua_gettable(L, -2);

  auto& span_context = get_span_context(L, -1);
  lua_pop(L, 1);

  return {reference_type, &span_context};
}

//------------------------------------------------------------------------------
// get_references
//------------------------------------------------------------------------------
static std::vector<
    std::pair<opentracing::SpanReferenceType, const opentracing::SpanContext*>>
get_references(lua_State* L) {
  switch (lua_type(L, -1)) {
    case LUA_TTABLE:
      break;
    case LUA_TNIL:
    case LUA_TNONE:
      return {};
    default:
      throw std::runtime_error{"references must be a table"};
  }
  std::vector<std::pair<opentracing::SpanReferenceType,
                        const opentracing::SpanContext*>>
      result;

  auto num_references = get_table_len(L, -1);
  result.reserve(num_references);
  for (int i = 1; i < num_references + 1; ++i) {
    lua_pushinteger(L, i);
    lua_gettable(L, -2);
    result.push_back(get_reference(L));
    lua_pop(L, 1);
  }

  return result;
}

//------------------------------------------------------------------------------
// get_tags
//------------------------------------------------------------------------------
static std::vector<std::pair<std::string, opentracing::Value>> get_tags(
    lua_State* L) {
  switch (lua_type(L, -1)) {
    case LUA_TTABLE:
      break;
    case LUA_TNIL:
    case LUA_TNONE:
      return {};
    default:
      throw std::runtime_error{"tags must be a table"};
  }
  return to_key_values(L, -1);
}

//------------------------------------------------------------------------------
// get_start_span_options
//------------------------------------------------------------------------------
static opentracing::StartSpanOptions get_start_span_options(lua_State* L,
                                                            int index) {
  opentracing::StartSpanOptions result;

  lua_getfield(L, index, "start_time");
  result.start_system_timestamp = convert_timestamp(L, -1);
  lua_pop(L, 1);

  lua_getfield(L, index, "references");
  result.references = get_references(L);
  lua_pop(L, 1);

  lua_getfield(L, index, "tags");
  result.tags = get_tags(L);
  lua_pop(L, 1);

  return result;
}

//------------------------------------------------------------------------------
// new_lua_tracer
//------------------------------------------------------------------------------
int LuaTracer::new_lua_tracer(lua_State* L) noexcept {
  auto library_name = luaL_checkstring(L, -2);
  auto config = luaL_checkstring(L, -1);
  auto userdata =
      static_cast<LuaTracer**>(lua_newuserdata(L, sizeof(LuaTracer*)));

  try {
    auto tracer = std::unique_ptr<LuaTracer>{
        new LuaTracer{load_tracer(library_name, config)}};
    *userdata = tracer.release();

    // tag the metatable
    luaL_getmetatable(L, description.metatable);
    lua_setmetatable(L, -2);

    return 1;
  } catch (const std::exception& e) {
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// new_lua_tracer_from_global
//------------------------------------------------------------------------------
int LuaTracer::new_lua_tracer_from_global(lua_State* L) noexcept {
  auto userdata =
      static_cast<LuaTracer**>(lua_newuserdata(L, sizeof(LuaTracer*)));
  try {
    auto ot_tracer = opentracing::Tracer::Global();
    if (ot_tracer == nullptr) {
      throw std::runtime_error{"opentracing::Global not initialized"};
    }
    auto tracer = std::unique_ptr<LuaTracer>{new LuaTracer{ot_tracer}};
    *userdata = tracer.release();

    // tag the metatable
    luaL_getmetatable(L, description.metatable);
    lua_setmetatable(L, -2);

    return 1;
  } catch (const std::exception& e) {
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// free
//------------------------------------------------------------------------------
int LuaTracer::free(lua_State* L) noexcept {
  auto tracer = check_lua_tracer(L);
  delete tracer;
  return 0;
}

//------------------------------------------------------------------------------
// start_span
//------------------------------------------------------------------------------
int LuaTracer::start_span(lua_State* L) noexcept {
  auto top = lua_gettop(L);
  auto tracer = check_lua_tracer(L);
  auto operation_name = luaL_checkstring(L, 2);
  auto num_arguments = lua_gettop(L);
  if (num_arguments >= 3) {
    luaL_checktype(L, 3, LUA_TTABLE);
  }
  auto userdata = static_cast<LuaSpan**>(lua_newuserdata(L, sizeof(LuaSpan*)));

  try {
    opentracing::StartSpanOptions start_span_options;
    if (num_arguments >= 3) {
      start_span_options = get_start_span_options(L, -2);
    }
    auto span = tracer->tracer_->StartSpanWithOptions(operation_name,
                                                      start_span_options);
    if (span == nullptr) {
      throw std::runtime_error{"unable to create span"};
    }
    auto lua_span = std::unique_ptr<LuaSpan>{new LuaSpan{
        tracer->tracer_, std::shared_ptr<opentracing::Span>{span.release()}}};
    *userdata = lua_span.release();

    luaL_getmetatable(L, LuaSpan::description.metatable);
    lua_setmetatable(L, -2);

    return 1;
  } catch (const std::exception& e) {
    lua_settop(L, top);
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// close
//------------------------------------------------------------------------------
int LuaTracer::close(lua_State* L) noexcept {
  auto tracer = check_lua_tracer(L);
  tracer->tracer_->Close();
  return 0;
}

//------------------------------------------------------------------------------
// inject
//------------------------------------------------------------------------------
template <class Carrier>
int LuaTracer::inject(lua_State* L) noexcept {
  auto top = lua_gettop(L);
  auto tracer = check_lua_tracer(L);
  luaL_checktype(L, -1, LUA_TTABLE);
  try {
    auto& span_context = get_span_context(L, -2);
    LuaCarrierWriter writer{L};
    auto was_successful = tracer->tracer_->Inject(
        span_context, static_cast<const Carrier&>(writer));
    if (!was_successful) {
      throw std::runtime_error{"failed to inject span context: " +
                               was_successful.error().message()};
    }
    return 0;
  } catch (const std::exception& e) {
    lua_settop(L, top);
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
};

//------------------------------------------------------------------------------
// binary_inject
//------------------------------------------------------------------------------
int LuaTracer::binary_inject(lua_State* L) noexcept {
  auto tracer = check_lua_tracer(L);
  try {
    auto& span_context = get_span_context(L, -1);
    std::ostringstream oss;
    auto was_successful = tracer->tracer_->Inject(span_context, oss);
    if (!was_successful) {
      throw std::runtime_error{"failed to inject span context: " +
                               was_successful.error().message()};
    }
    lua_pushstring(L, oss.str().c_str());
    return 1;
  } catch (const std::exception& e) {
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// extract
//------------------------------------------------------------------------------
template <class Carrier>
int LuaTracer::extract(lua_State* L) noexcept {
  auto top = lua_gettop(L);
  auto tracer = check_lua_tracer(L);
  luaL_checktype(L, -1, LUA_TTABLE);
  auto userdata = static_cast<LuaSpanContext**>(
      lua_newuserdata(L, sizeof(LuaSpanContext*)));
  try {
    lua_pushvalue(L, -2);
    LuaCarrierReader reader{L};
    auto span_context_maybe =
        tracer->tracer_->Extract(static_cast<const Carrier&>(reader));
    lua_pop(L, 1);
    if (!span_context_maybe) {
      throw std::runtime_error{"failed to inject span context: " +
                               span_context_maybe.error().message()};
    }
    auto span_context = std::move(*span_context_maybe);
    if (span_context == nullptr) {
      lua_pushnil(L);
      return 1;
    }

    *userdata = new LuaSpanContext{std::move(span_context)};
    luaL_getmetatable(L, LuaSpanContext::description.metatable);
    lua_setmetatable(L, -2);

    return 1;
  } catch (const std::exception& e) {
    lua_settop(L, top);
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

int LuaTracer::binary_extract(lua_State* L) noexcept {
  auto tracer = check_lua_tracer(L);
  size_t context_len;
  auto context_data = luaL_checklstring(L, -1, &context_len);
  auto userdata = static_cast<LuaSpanContext**>(
      lua_newuserdata(L, sizeof(LuaSpanContext*)));
  try {
    std::istringstream iss{std::string{context_data, context_len}};
    auto span_context_maybe = tracer->tracer_->Extract(iss);
    if (!span_context_maybe) {
      throw std::runtime_error{"failed to inject span context: " +
                               span_context_maybe.error().message()};
    }
    auto span_context = std::move(*span_context_maybe);
    if (span_context == nullptr) {
      lua_pushnil(L);
      return 1;
    }

    *userdata = new LuaSpanContext{std::move(span_context)};
    luaL_getmetatable(L, LuaSpanContext::description.metatable);
    lua_setmetatable(L, -2);

    return 1;
  } catch (const std::exception& e) {
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// description
//------------------------------------------------------------------------------
const LuaClassDescription LuaTracer::description = {
    METATABLE,
    LuaTracer::free,
    {{"start_span", LuaTracer::start_span},
     {"text_map_inject", LuaTracer::inject<opentracing::TextMapWriter>},
     {"http_headers_inject", LuaTracer::inject<opentracing::HTTPHeadersWriter>},
     {"binary_inject", LuaTracer::binary_inject},
     {"text_map_extract", LuaTracer::extract<opentracing::TextMapReader>},
     {"http_headers_extract",
      LuaTracer::extract<opentracing::HTTPHeadersReader>},
     {"binary_extract", LuaTracer::binary_extract},
     {"close", LuaTracer::close},
     {nullptr, nullptr}}};
}  // namespace lua_bridge_tracer
