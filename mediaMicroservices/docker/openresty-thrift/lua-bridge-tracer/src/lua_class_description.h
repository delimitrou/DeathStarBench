#pragma once

#include <initializer_list>

extern "C" {
#include <lauxlib.h>
#include <lua.h>
}  // extern "C"

namespace lua_bridge_tracer {
struct LuaClassDescription {
  const char* metatable;
  int (*free_function)(lua_State*);
  std::initializer_list<const struct luaL_Reg> methods;
};
}  // namespace lua_bridge_tracer
