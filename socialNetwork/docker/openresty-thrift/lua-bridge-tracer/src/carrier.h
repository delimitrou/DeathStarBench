#pragma once

#include <opentracing/propagation.h>

extern "C" {
#include <lauxlib.h>
#include <lua.h>
}  // extern "C"

namespace lua_bridge_tracer {
class LuaCarrierWriter : public opentracing::HTTPHeadersWriter {
 public:
  explicit LuaCarrierWriter(lua_State* lua_state) noexcept;

  opentracing::expected<void> Set(opentracing::string_view key,
                                  opentracing::string_view value) const final;

 private:
  lua_State* lua_state_;
};

class LuaCarrierReader : public opentracing::HTTPHeadersReader {
 public:
  explicit LuaCarrierReader(lua_State* lua_state) noexcept;

  opentracing::expected<void> ForeachKey(
      std::function<opentracing::expected<void>(opentracing::string_view key,
                                                opentracing::string_view value)>
          f) const final;

 private:
  lua_State* lua_state_;
};
}  // namespace lua_bridge_tracer
