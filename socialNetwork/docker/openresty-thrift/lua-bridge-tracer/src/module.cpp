#include "lua_span.h"
#include "lua_span_context.h"
#include "lua_tracer.h"

#include <opentracing/dynamic_load.h>
#include <iostream>
#include <iterator>

extern "C" {
#include <lauxlib.h>
#include <lua.h>
}  // extern "C"

// Copied from Lua 5.3 so that we can use it with Lua 5.1.
static void setfuncs(lua_State* L, const luaL_Reg* l, int nup) {
  luaL_checkstack(L, nup + 1, "too many upvalues");
  for (; l->name != NULL; l++) { /* fill the table with given functions */
    int i;
    lua_pushstring(L, l->name);
    for (i = 0; i < nup; i++) /* copy upvalues to the top */
      lua_pushvalue(L, -(nup + 1));
    lua_pushcclosure(L, l->func, nup); /* closure with those upvalues */
    lua_settable(L, -(nup + 3));
  }
  lua_pop(L, nup); /* remove upvalues */
}

static void make_lua_class(
    lua_State* L, const lua_bridge_tracer::LuaClassDescription& description) {
  luaL_newmetatable(L, description.metatable);

  if (description.free_function != nullptr) {
    lua_pushstring(L, "__gc");
    lua_pushcfunction(L, description.free_function);
    lua_settable(L, -3);
  }

  setfuncs(L, &*std::begin(description.methods), 0);

  lua_pushstring(L, "__index");
  lua_pushvalue(L, -2); /* pushes the metatable */
  lua_settable(L, -3);  /* metatable.__index = metatable */
}

extern "C" int luaopen_opentracing_bridge_tracer(lua_State* L) {
  make_lua_class(L, lua_bridge_tracer::LuaTracer::description);
  make_lua_class(L, lua_bridge_tracer::LuaSpan::description);
  make_lua_class(L, lua_bridge_tracer::LuaSpanContext::description);

  lua_newtable(L);
  const struct luaL_Reg functions[] = {
      {"new", lua_bridge_tracer::LuaTracer::new_lua_tracer},
      {"new_from_global",
       lua_bridge_tracer::LuaTracer::new_lua_tracer_from_global},
      {nullptr, nullptr}};
  setfuncs(L, functions, 0);

  return 1;
}
