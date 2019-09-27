#pragma once

#include "lua_class_description.h"

#include <opentracing/span.h>

#include <memory>

namespace lua_bridge_tracer {
class LuaSpanContext {
 public:
  // OpenTracing C++ doesn't yet support copying the opentracing::SpanContext
  // from an opentracing::Span (See
  // https://github.com/opentracing/opentracing-cpp/pull/56).
  //
  // So when the opentracing::SpanContext is referenced we need to hold an
  // std::shared_ptr to the opentracing::Span to ensure that it isn't freed.
  explicit LuaSpanContext(const std::shared_ptr<const opentracing::Span>& span)
      : span_{span} {}

  explicit LuaSpanContext(
      std::unique_ptr<const opentracing::SpanContext>&& span_context)
      : span_context_{std::move(span_context)} {}

  static const LuaClassDescription description;

  const opentracing::SpanContext& span_context() const noexcept {
    if (span_ != nullptr) return span_->context();
    return *span_context_;
  }

 private:
  std::shared_ptr<const opentracing::Span> span_;
  std::unique_ptr<const opentracing::SpanContext> span_context_;

  static int free(lua_State* L) noexcept;
};
}  // namespace lua_bridge_tracer
