#ifndef SOCIAL_NETWORK_MICROSERVICES_SRC_UTILS_REDIS_H_
#define SOCIAL_NETWORK_MICROSERVICES_SRC_UTILS_REDIS_H_

#include <sw/redis++/redis++.h>
#include <chrono>

using namespace sw::redis;
namespace social_network {

Redis init_redis_client_pool(
    const json &config_json,
    const std::string &service_name
) {
  ConnectionOptions connection_options;
  connection_options.host = config_json[service_name + "-redis"]["addr"];
  connection_options.port = config_json[service_name + "-redis"]["port"];

  if (config_json["ssl"]["enabled"]) {
    std::string ca_file = config_json["ssl"]["caPath"];

    connection_options.tls.enabled = true;
    connection_options.tls.cacert = ca_file.c_str();
  }

  ConnectionPoolOptions pool_options;
  pool_options.size = config_json[service_name + "-redis"]["connections"];
  pool_options.wait_timeout = std::chrono::milliseconds(config_json[service_name + "-redis"]["timeout_ms"]);
  pool_options.connection_lifetime = std::chrono::milliseconds(config_json[service_name + "-redis"]["keepalive_ms"]);

  return Redis(connection_options, pool_options);
}

RedisCluster init_redis_cluster_client_pool(
    const json &config_json,
    const std::string &service_name
) {
  ConnectionOptions connection_options;
  connection_options.host = config_json[service_name + "-redis"]["addr"];
  connection_options.port = config_json[service_name + "-redis"]["port"];

  if (config_json["ssl"]["enabled"]) {
    std::string ca_file = config_json["ssl"]["caPath"];

    connection_options.tls.enabled = true;
    connection_options.tls.cacert = ca_file.c_str();
  }

  ConnectionPoolOptions pool_options;
  pool_options.size = config_json[service_name + "-redis"]["connections"];
  pool_options.wait_timeout = std::chrono::milliseconds(config_json[service_name + "-redis"]["timeout_ms"]);
  pool_options.connection_lifetime = std::chrono::milliseconds(config_json[service_name + "-redis"]["keepalive_ms"]);

  return RedisCluster(connection_options, pool_options);
}

Redis init_redis_replica_client_pool(
    const json& config_json,
    const std::string& service_name
) {
    ConnectionOptions connection_options;
    connection_options.host = config_json[service_name]["addr"];
    connection_options.port = config_json[service_name]["port"];

    if (config_json["ssl"]["enabled"]) {
        std::string ca_file = config_json["ssl"]["caPath"];

        connection_options.tls.enabled = true;
        connection_options.tls.cacert = ca_file.c_str();
    }

    ConnectionPoolOptions pool_options;
    pool_options.size = config_json[service_name]["connections"];
    pool_options.wait_timeout = std::chrono::milliseconds(config_json[service_name]["timeout_ms"]);
    pool_options.connection_lifetime = std::chrono::milliseconds(config_json[service_name]["keepalive_ms"]);

    return Redis(connection_options, pool_options);
}


} // namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_SRC_UTILS_REDIS_H_
