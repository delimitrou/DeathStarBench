#include "lua_span.h"

#include "lua_span_context.h"
#include "lua_tracer.h"
#include "utility.h"

#define METATABLE "lua_opentracing_bridge.span"

namespace lua_bridge_tracer {
//------------------------------------------------------------------------------
// check_lua_span
//------------------------------------------------------------------------------
static LuaSpan* check_lua_span(lua_State* L) noexcept {
  void* user_data = luaL_checkudata(L, 1, METATABLE);
  luaL_argcheck(L, user_data != NULL, 1, "`" METATABLE "' expected");
  return *static_cast<LuaSpan**>(user_data);
}

//------------------------------------------------------------------------------
// get_finish_span_options
//------------------------------------------------------------------------------
static opentracing::FinishSpanOptions get_finish_span_options(lua_State* L,
                                                              int index) {
  opentracing::FinishSpanOptions result;

  result.finish_steady_timestamp =
      opentracing::convert_time_point<opentracing::SteadyClock>(
          convert_timestamp(L, index));

  return result;
}

//------------------------------------------------------------------------------
// free
//------------------------------------------------------------------------------
int LuaSpan::free(lua_State* L) noexcept {
  auto span = check_lua_span(L);
  delete span;
  return 0;
}

//------------------------------------------------------------------------------
// set_operation_name
//------------------------------------------------------------------------------
int LuaSpan::set_operation_name(lua_State* L) noexcept {
  auto span = check_lua_span(L);
  size_t operation_name_len;
  auto operation_name_data = luaL_checklstring(L, -1, &operation_name_len);
  span->span_->SetOperationName({operation_name_data, operation_name_len});
  return 0;
}

//------------------------------------------------------------------------------
// tracer
//------------------------------------------------------------------------------
int LuaSpan::tracer(lua_State* L) noexcept {
  auto span = check_lua_span(L);
  auto userdata =
      static_cast<LuaTracer**>(lua_newuserdata(L, sizeof(LuaTracer*)));

  try {
    auto tracer = std::unique_ptr<LuaTracer>{new LuaTracer{span->tracer_}};
    *userdata = tracer.release();

    // tag the metatable
    luaL_getmetatable(L, LuaTracer::description.metatable);
    lua_setmetatable(L, -2);

    return 1;
  } catch (const std::exception& e) {
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// finish
//------------------------------------------------------------------------------
int LuaSpan::finish(lua_State* L) noexcept {
  auto span = check_lua_span(L);
  auto num_arguments = lua_gettop(L);
  if (num_arguments >= 2) {
    luaL_checknumber(L, 2);
  }
  try {
    opentracing::FinishSpanOptions finish_span_options;
    if (num_arguments >= 2) {
      finish_span_options = get_finish_span_options(L, 2);
    }
    finish_span_options.log_records = std::move(span->log_records_);
    span->span_->FinishWithOptions(finish_span_options);
    return 0;
  } catch (const std::exception& e) {
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// context
//------------------------------------------------------------------------------
int LuaSpan::context(lua_State* L) noexcept {
  auto span = check_lua_span(L);
  auto userdata = static_cast<LuaSpanContext**>(
      lua_newuserdata(L, sizeof(LuaSpanContext*)));
  try {
    auto lua_span_context =
        std::unique_ptr<LuaSpanContext>{new LuaSpanContext{span->span_}};
    *userdata = lua_span_context.release();

    luaL_getmetatable(L, LuaSpanContext::description.metatable);
    lua_setmetatable(L, -2);

    return 1;
  } catch (const std::exception& e) {
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// set_tag
//------------------------------------------------------------------------------
int LuaSpan::set_tag(lua_State* L) noexcept {
  auto span = check_lua_span(L);
  size_t key_len;
  auto key_data = luaL_checklstring(L, -2, &key_len);
  try {
    opentracing::string_view key{key_data, key_len};
    auto value = to_value(L, -1);
    span->span_->SetTag(key, std::move(value));
    return 0;
  } catch (const std::exception& e) {
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// log_kv
//------------------------------------------------------------------------------
int LuaSpan::log_kv(lua_State* L) noexcept {
  auto span = check_lua_span(L);
  luaL_checktype(L, -1, LUA_TTABLE);
  try {
    opentracing::LogRecord log_record;
    log_record.timestamp = std::chrono::system_clock::now();
    log_record.fields = to_key_values(L, -1);
    span->log_records_.push_back(log_record);
    return 0;
  } catch (const std::exception& e) {
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// set_baggage_item
//------------------------------------------------------------------------------
int LuaSpan::set_baggage_item(lua_State* L) noexcept {
  auto span = check_lua_span(L);
  size_t key_len;
  auto key_data = luaL_checklstring(L, 2, &key_len);
  size_t value_len;
  auto value_data = luaL_checklstring(L, 3, &value_len);
  try {
    span->span_->SetBaggageItem({key_data, key_len}, {value_data, value_len});
    return 0;
  } catch (const std::exception& e) {
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// get_baggage_item
//------------------------------------------------------------------------------
int LuaSpan::get_baggage_item(lua_State* L) noexcept {
  auto span = check_lua_span(L);
  size_t key_len;
  auto key_data = luaL_checklstring(L, 2, &key_len);
  try {
    auto baggage_item = span->span_->BaggageItem({key_data, key_len});
    lua_pushstring(L, baggage_item.c_str());
    return 1;
  } catch (const std::exception& e) {
    lua_pushstring(L, e.what());
  }
  return lua_error(L);
}

//------------------------------------------------------------------------------
// description
//------------------------------------------------------------------------------
const LuaClassDescription LuaSpan::description = {
    METATABLE,
    LuaSpan::free,
    {{"context", LuaSpan::context},
     {"tracer", LuaSpan::tracer},
     {"set_operation_name", LuaSpan::set_operation_name},
     {"finish", LuaSpan::finish},
     {"set_tag", LuaSpan::set_tag},
     {"log_kv", LuaSpan::log_kv},
     {"set_baggage_item", LuaSpan::set_baggage_item},
     {"get_baggage_item", LuaSpan::get_baggage_item},
     {nullptr, nullptr}}};
}  // namespace lua_bridge_tracer
