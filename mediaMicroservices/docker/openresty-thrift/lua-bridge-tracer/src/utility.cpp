#include "utility.h"

#include <stdexcept>

namespace lua_bridge_tracer {
//------------------------------------------------------------------------------
// get_table_length
//------------------------------------------------------------------------------
size_t get_table_len(lua_State* L, int index) {
#if LUA_VERSION_NUM > 501
  return lua_rawlen(L, index);
#else
  return lua_objlen(L, index);
#endif
}

//------------------------------------------------------------------------------
// convert_timestamp
//------------------------------------------------------------------------------
std::chrono::system_clock::time_point convert_timestamp(lua_State* L,
                                                        int index) {
  using SystemClock = std::chrono::system_clock;
  switch (lua_type(L, index)) {
    case LUA_TNUMBER:
      break;
    case LUA_TNIL:
    case LUA_TNONE:
      return {};
    default:
      throw std::runtime_error{"timestamp must be a number"};
  }
  auto time_since_epoch =
      std::chrono::microseconds{static_cast<uint64_t>(lua_tonumber(L, index))};
  return SystemClock::from_time_t(std::time_t(0)) +
         std::chrono::duration_cast<SystemClock::duration>(time_since_epoch);
}

//------------------------------------------------------------------------------
// to_value
//------------------------------------------------------------------------------
static opentracing::Value to_dictionary_value(lua_State* L, int index);

opentracing::Value to_value(lua_State* L, int index) {
  switch (lua_type(L, index)) {
    case LUA_TNUMBER: {
      return static_cast<double>(lua_tonumber(L, index));
    }
    case LUA_TSTRING: {
      size_t value_len;
      auto value_data = lua_tolstring(L, index, &value_len);
      std::string value{value_data, value_len};
      return std::move(value);
    }
    case LUA_TBOOLEAN: {
      return static_cast<bool>(lua_toboolean(L, index));
    }
    case LUA_TTABLE: {
      return to_dictionary_value(L, index);
    }
    case LUA_TNIL:
    case LUA_TNONE: {
      return nullptr;
    }
    default:
      throw std::runtime_error{"invalid value type"};
  }
}

//------------------------------------------------------------------------------
// to_dictionary_value
//------------------------------------------------------------------------------
opentracing::Value to_dictionary_value(lua_State* L, int index) {
  auto key_values = to_key_values(L, index);
  opentracing::Dictionary result{std::begin(key_values), std::end(key_values)};
  return result;
}

//------------------------------------------------------------------------------
// to_key_values
//------------------------------------------------------------------------------
std::vector<std::pair<std::string, opentracing::Value>> to_key_values(
    lua_State* L, int index) {
  lua_pushvalue(L, index);
  std::vector<std::pair<std::string, opentracing::Value>> result;
  result.reserve(get_table_len(L, index));
  lua_pushnil(L);
  while (lua_next(L, -2)) {
    // ignore if the key isn't a string
    if (!lua_isstring(L, -2)) {
      lua_pop(L, 1);
      continue;
    }
    lua_pushvalue(L, -2);
    size_t key_len, value_len;
    auto key = lua_tolstring(L, -1, &key_len);
    auto value = to_value(L, -2);
    result.emplace_back(key, value);
    lua_pop(L, 2);
  }
  lua_pop(L, 1);
  return result;
}
}  // namespace lua_bridge_tracer
