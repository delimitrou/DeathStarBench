#include <utility>

#ifndef SOCIAL_NETWORK_MICROSERVICES_TRACING_H
#define SOCIAL_NETWORK_MICROSERVICES_TRACING_H

#include <string>
#include <yaml-cpp/yaml.h>
#include <jaegertracing/Tracer.h>

#include <opentracing/propagation.h>
#include <string>
#include <map>
#include "logger.h"

namespace social_network {

using opentracing::expected;
using opentracing::string_view;

class TextMapReader : public opentracing::TextMapReader {
 public:
  explicit TextMapReader(const std::map<std::string, std::string> &text_map)
      : _text_map(text_map) {}

  expected<void> ForeachKey(
      std::function<expected<void>(string_view key, string_view value)> f)
  const override {
    for (const auto& key_value : _text_map) {
      auto result = f(key_value.first, key_value.second);
      if (!result) return result;
    }
    return {};
  }

 private:
  const std::map<std::string, std::string>& _text_map;
};

class TextMapWriter : public opentracing::TextMapWriter {
 public:
  explicit TextMapWriter(std::map<std::string, std::string> &text_map)
    : _text_map(text_map) {}

  expected<void> Set(string_view key, string_view value) const override {
    _text_map[key] = value;
    return {};
  }

 private:
  std::map<std::string, std::string>& _text_map;
};

void SetUpTracer(
    const std::string &config_file_path,
    const std::string &service) {
  auto configYAML = YAML::LoadFile(config_file_path);

  // Enable local Jaeger agent, by prepending the service name to the default
  // Jaeger agent's hostname
  // configYAML["reporter"]["localAgentHostPort"] = service + "-" +
  //     configYAML["reporter"]["localAgentHostPort"].as<std::string>();

  auto config = jaegertracing::Config::parse(configYAML);

  bool r = false;
  while (!r) {
    try
    {
      auto tracer = jaegertracing::Tracer::make(
        service, config, jaegertracing::logging::consoleLogger());
      r = true;
      opentracing::Tracer::InitGlobal(
      std::static_pointer_cast<opentracing::Tracer>(tracer));
    }
    catch(...)
    {
      LOG(error) << "Failed to connect to jaeger, retrying ...";
      sleep(1);
    }
  }


}


} //namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_TRACING_H
