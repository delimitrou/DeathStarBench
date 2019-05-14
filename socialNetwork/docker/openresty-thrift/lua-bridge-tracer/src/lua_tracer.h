#pragma once

#include "lua_class_description.h"

#include <opentracing/tracer.h>

#include <memory>

extern "C" {
#include <lauxlib.h>
#include <lua.h>
}  // extern "C"

namespace lua_bridge_tracer {
class LuaTracer {
 public:
  explicit LuaTracer(const std::shared_ptr<opentracing::Tracer>& tracer)
      : tracer_{tracer} {}

  static const LuaClassDescription description;

  static int new_lua_tracer(lua_State* L) noexcept;

  static int new_lua_tracer_from_global(lua_State* L) noexcept;

 private:
  std::shared_ptr<opentracing::Tracer> tracer_;

  static int free(lua_State* L) noexcept;

  static int start_span(lua_State* L) noexcept;

  template <class Carrier>
  static int inject(lua_State* L) noexcept;

  static int binary_inject(lua_State* L) noexcept;

  template <class Carrier>
  static int extract(lua_State* L) noexcept;

  static int binary_extract(lua_State* L) noexcept;

  static int close(lua_State* L) noexcept;
};
}  // namespace lua_bridge_tracer
