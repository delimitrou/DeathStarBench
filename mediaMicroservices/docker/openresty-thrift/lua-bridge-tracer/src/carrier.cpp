#include "carrier.h"

namespace lua_bridge_tracer {
//------------------------------------------------------------------------------
// constructor
//------------------------------------------------------------------------------
LuaCarrierWriter::LuaCarrierWriter(lua_State* lua_state) noexcept
    : lua_state_{lua_state} {}

//------------------------------------------------------------------------------
// Set
//------------------------------------------------------------------------------
opentracing::expected<void> LuaCarrierWriter::Set(
    opentracing::string_view key, opentracing::string_view value) const {
  lua_pushlstring(lua_state_, key.data(), key.size());
  lua_pushlstring(lua_state_, value.data(), value.size());
  lua_settable(lua_state_, -3);
  return {};
}

//------------------------------------------------------------------------------
// constructor
//------------------------------------------------------------------------------
LuaCarrierReader::LuaCarrierReader(lua_State* lua_state) noexcept
    : lua_state_{lua_state} {}

//------------------------------------------------------------------------------
// ForeachKey
//------------------------------------------------------------------------------
opentracing::expected<void> LuaCarrierReader::ForeachKey(
    std::function<opentracing::expected<void>(opentracing::string_view key,
                                              opentracing::string_view value)>
        f) const {
  auto top = lua_gettop(lua_state_);
  lua_pushnil(lua_state_);
  while (lua_next(lua_state_, -2)) {
    // ignore if the key or value isn't a string
    if (!lua_isstring(lua_state_, -1) || !lua_isstring(lua_state_, -2)) {
      lua_pop(lua_state_, 1);
      continue;
    }
    lua_pushvalue(lua_state_, -2);
    size_t key_len, value_len;
    auto key = lua_tolstring(lua_state_, -1, &key_len);
    auto value = lua_tolstring(lua_state_, -2, &value_len);
    auto was_successful = f({key, key_len}, {value, value_len});
    if (!was_successful) {
      lua_settop(lua_state_, top);
      return was_successful;
    }
    lua_pop(lua_state_, 2);
  }
  return {};
}
}  // namespace lua_bridge_tracer
