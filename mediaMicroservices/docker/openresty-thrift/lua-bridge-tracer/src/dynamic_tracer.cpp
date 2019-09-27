#include "dynamic_tracer.h"

#include <opentracing/dynamic_load.h>

#include <stdexcept>

namespace lua_bridge_tracer {
//------------------------------------------------------------------------------
// DynamicSpan
//------------------------------------------------------------------------------
namespace {
class DynamicSpan : public opentracing::Span {
 public:
  DynamicSpan(std::shared_ptr<opentracing::Tracer> tracer,
              std::unique_ptr<opentracing::Span>&& span) noexcept
      : tracer_{tracer}, span_{std::move(span)} {}

 private:
  std::shared_ptr<opentracing::Tracer> tracer_;
  std::unique_ptr<opentracing::Span> span_;

  void FinishWithOptions(const opentracing::FinishSpanOptions&
                             finish_span_options) noexcept final {
    span_->FinishWithOptions(finish_span_options);
  }

  void SetOperationName(opentracing::string_view name) noexcept final {
    span_->SetOperationName(name);
  }

  void SetTag(opentracing::string_view key,
              const opentracing::Value& value) noexcept final {
    span_->SetTag(key, value);
  }

  void SetBaggageItem(opentracing::string_view restricted_key,
                      opentracing::string_view value) noexcept final {
    span_->SetBaggageItem(restricted_key, value);
  }

  std::string BaggageItem(opentracing::string_view restricted_key) const
      noexcept final {
    return span_->BaggageItem(restricted_key);
  }

  void Log(std::initializer_list<
           std::pair<opentracing::string_view, opentracing::Value>>
               fields) noexcept final {
    span_->Log(fields);
  }

  const opentracing::SpanContext& context() const noexcept final {
    return span_->context();
  }

  const opentracing::Tracer& tracer() const noexcept final { return *tracer_; }
};
}  // namespace

//------------------------------------------------------------------------------
// DynamicTracer
//------------------------------------------------------------------------------
namespace {
class DynamicTracer : public opentracing::Tracer,
                      public std::enable_shared_from_this<DynamicTracer> {
 public:
  DynamicTracer(opentracing::DynamicTracingLibraryHandle&& handle,
                std::shared_ptr<opentracing::Tracer>&& tracer) noexcept
      : handle_{std::move(handle)}, tracer_{std::move(tracer)} {}

 private:
  opentracing::DynamicTracingLibraryHandle handle_;
  std::shared_ptr<opentracing::Tracer> tracer_;

  std::unique_ptr<opentracing::Span> StartSpanWithOptions(
      opentracing::string_view operation_name,
      const opentracing::StartSpanOptions& options) const noexcept final {
    auto span = tracer_->StartSpanWithOptions(operation_name, options);
    if (span == nullptr) return nullptr;
    return std::unique_ptr<opentracing::Span>{
        new (std::nothrow) DynamicSpan(tracer_, std::move(span))};
  }

  opentracing::expected<void> Inject(const opentracing::SpanContext& sc,
                                     std::ostream& writer) const final {
    return tracer_->Inject(sc, writer);
  }

  opentracing::expected<void> Inject(
      const opentracing::SpanContext& sc,
      const opentracing::TextMapWriter& writer) const final {
    return tracer_->Inject(sc, writer);
  }

  opentracing::expected<void> Inject(
      const opentracing::SpanContext& sc,
      const opentracing::HTTPHeadersWriter& writer) const final {
    return tracer_->Inject(sc, writer);
  }

  opentracing::expected<void> Inject(
      const opentracing::SpanContext& sc,
      const opentracing::CustomCarrierWriter& writer) const final {
    return tracer_->Inject(sc, writer);
  }

  opentracing::expected<std::unique_ptr<opentracing::SpanContext>> Extract(
      std::istream& reader) const final {
    return tracer_->Extract(reader);
  }

  opentracing::expected<std::unique_ptr<opentracing::SpanContext>> Extract(
      const opentracing::TextMapReader& reader) const final {
    return tracer_->Extract(reader);
  }

  opentracing::expected<std::unique_ptr<opentracing::SpanContext>> Extract(
      const opentracing::HTTPHeadersReader& reader) const final {
    return tracer_->Extract(reader);
  }

  opentracing::expected<std::unique_ptr<opentracing::SpanContext>> Extract(
      const opentracing::CustomCarrierReader& reader) const final {
    return reader.Extract(*tracer_);
  }

  void Close() noexcept final { tracer_->Close(); }
};
}  // namespace

//------------------------------------------------------------------------------
// make_dynamic_tracer
//------------------------------------------------------------------------------
// Dynamically loads a C++ OpenTracing plugin and constructs a tracer with the
// given configuration.
//
// The opentracing::DynamicTracingLibraryHandle returned can't be freed until
// the opentracing::Tracer is freed. To accomplish this, we build a new
// opentracing::Tracer that wraps the plugin's tracer and owns the
// opentracing::DynamicTracingLibraryHandle.
std::shared_ptr<opentracing::Tracer> load_tracer(const char* library_name,
                                                 const char* config) {
  std::string error_message;
  auto handle_maybe =
      opentracing::DynamicallyLoadTracingLibrary(library_name, error_message);
  if (!handle_maybe) {
    throw std::runtime_error{error_message};
  }
  auto& handle = *handle_maybe;
  auto tracer_maybe = handle.tracer_factory().MakeTracer(config, error_message);
  if (!tracer_maybe) {
    throw std::runtime_error{error_message};
  }
  return std::make_shared<DynamicTracer>(std::move(handle),
                                         std::move(*tracer_maybe));
}
}  // namespace lua_bridge_tracer
