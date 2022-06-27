#ifndef SOCIAL_NETWORK_MICROSERVICES_SRC_UTILS_MEMCACHED_H_
#define SOCIAL_NETWORK_MICROSERVICES_SRC_UTILS_MEMCACHED_H_

#include <libmemcached/memcached.h>
#include <libmemcached/util.h>

namespace social_network {

memcached_pool_st *init_memcached_client_pool(
    const json &config_json,
    const std::string &service_name,
    uint32_t min_size,
    uint32_t max_size
) {
  std::string addr = config_json[service_name + "-memcached"]["addr"];
  int port = config_json[service_name + "-memcached"]["port"];
  int use_binary_protocol = config_json[service_name + "-memcached"]["binary_protocol"];
  std::string config_str = "--SERVER=" + addr + ":" + std::to_string(port);
  auto memcached_client = memcached(config_str.c_str(), config_str.length());
  memcached_behavior_set(memcached_client, MEMCACHED_BEHAVIOR_NO_BLOCK, 1);
  memcached_behavior_set(memcached_client, MEMCACHED_BEHAVIOR_TCP_NODELAY, 1);
  if (use_binary_protocol == 1) {
    memcached_behavior_set(memcached_client, MEMCACHED_BEHAVIOR_BINARY_PROTOCOL, 1);
  }

  auto memcached_client_pool =
      memcached_pool_create(memcached_client, min_size, max_size);
  return memcached_client_pool;
}

} // namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_SRC_UTILS_MEMCACHED_H_
