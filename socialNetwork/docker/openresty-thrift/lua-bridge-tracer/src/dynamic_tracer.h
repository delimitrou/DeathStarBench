#pragma once

#include <opentracing/tracer.h>

namespace lua_bridge_tracer {
std::shared_ptr<opentracing::Tracer> load_tracer(const char* tracer_library,
                                                 const char* config);
}  // namespace lua_bridge_tracer
