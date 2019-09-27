#pragma once

#include <opentracing/value.h>

#include <chrono>

extern "C" {
#include <lauxlib.h>
#include <lua.h>
}  // extern "C"

namespace lua_bridge_tracer {
size_t get_table_len(lua_State* L, int index);

std::chrono::system_clock::time_point convert_timestamp(lua_State* L,
                                                        int index);

opentracing::Value to_value(lua_State* L, int index);

std::vector<std::pair<std::string, opentracing::Value>> to_key_values(
    lua_State* L, int index);
}  // namespace lua_bridge_tracer
