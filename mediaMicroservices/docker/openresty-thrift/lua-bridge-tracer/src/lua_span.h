#pragma once

#include "lua_class_description.h"

#include <opentracing/tracer.h>

#include <memory>

namespace lua_bridge_tracer {
class LuaSpan {
 public:
  explicit LuaSpan(const std::shared_ptr<opentracing::Tracer>& tracer,
                   const std::shared_ptr<opentracing::Span>& span)
      : tracer_{tracer}, span_{span} {}

  static const LuaClassDescription description;

 private:
  std::shared_ptr<opentracing::Tracer> tracer_;
  std::shared_ptr<opentracing::Span> span_;
  std::vector<opentracing::LogRecord> log_records_;

  static int free(lua_State* L) noexcept;

  static int set_operation_name(lua_State* L) noexcept;

  static int tracer(lua_State* L) noexcept;

  static int finish(lua_State* L) noexcept;

  static int context(lua_State* L) noexcept;

  static int set_tag(lua_State* L) noexcept;

  static int log_kv(lua_State* L) noexcept;

  static int set_baggage_item(lua_State* L) noexcept;

  static int get_baggage_item(lua_State* L) noexcept;
};
}  // namespace lua_bridge_tracer
