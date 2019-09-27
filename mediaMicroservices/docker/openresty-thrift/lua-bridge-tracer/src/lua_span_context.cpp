#include "lua_span_context.h"

#define METATABLE "lua_opentracing_bridge.span_context"

namespace lua_bridge_tracer {
//------------------------------------------------------------------------------
// check_lua_tracer
//------------------------------------------------------------------------------
static LuaSpanContext* check_lua_span_context(lua_State* L) noexcept {
  void* user_data = luaL_checkudata(L, 1, METATABLE);
  luaL_argcheck(L, user_data != NULL, 1, "`" METATABLE "' expected");
  return *static_cast<LuaSpanContext**>(user_data);
}

//------------------------------------------------------------------------------
// free
//------------------------------------------------------------------------------
int LuaSpanContext::free(lua_State* L) noexcept {
  auto span_context = check_lua_span_context(L);
  delete span_context;
  return 0;
}
//------------------------------------------------------------------------------
// description
//------------------------------------------------------------------------------
const LuaClassDescription LuaSpanContext::description = {
    METATABLE, LuaSpanContext::free, {{nullptr, nullptr}}};
}  // namespace lua_bridge_tracer
