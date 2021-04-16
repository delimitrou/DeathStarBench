#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TThreadedServer.h>
#include <thrift/transport/TServerSocket.h>
#include <thrift/transport/TBufferTransports.h>
#include <nlohmann/json.hpp>
#include <signal.h>

#include "../ClientPool.h"
#include "../RedisClient.h"
#include "../logger.h"
#include "../tracing.h"
#include "../utils.h"
#include "../utils_mongodb.h"
#include "../../gen-cpp/social_network_types.h"
#include "UserTimelineHandler.h"

using apache::thrift::server::TThreadedServer;
using apache::thrift::transport::TServerSocket;
using apache::thrift::transport::TFramedTransportFactory;
using apache::thrift::protocol::TBinaryProtocolFactory;
using namespace social_network;

void sigintHandler(int sig) {
  exit(EXIT_SUCCESS);
}

int main(int argc, char *argv[]) {
  signal(SIGINT, sigintHandler);
  init_logger();
  SetUpTracer("config/jaeger-config.yml", "user-timeline-service");

  json config_json;
  if (load_config_file("config/service-config.json", &config_json) != 0) {
    exit(EXIT_FAILURE);
  }

  int port = config_json["user-timeline-service"]["port"];
  std::string redis_addr =
      config_json["user-timeline-redis"]["addr"];
  int redis_port = config_json["user-timeline-redis"]["port"];
  int redis_conns = config_json["user-timeline-redis"]["connections"];
  int redis_timeout = config_json["user-timeline-redis"]["timeout_ms"];
  int redis_keepalive = config_json["user-timeline-redis"]["keepalive_ms"];

  int post_storage_port = config_json["post-storage-service"]["port"];
  std::string post_storage_addr = config_json["post-storage-service"]["addr"];
  int post_storage_conns = config_json["post-storage-service"]["connections"];
  int post_storage_timeout = config_json["post-storage-service"]["timeout_ms"];
  int post_storage_keepalive = config_json["post-storage-service"]["keepalive_ms"];

  int mongodb_conns = config_json["user-timeline-mongodb"]["connections"];
  int mongodb_timeout = config_json["user-timeline-mongodb"]["timeout_ms"];

  auto mongodb_client_pool = init_mongodb_client_pool(
      config_json, "user-timeline", mongodb_conns);

  ClientPool<RedisClient> redis_client_pool("user-timeline-redis", 
      redis_addr, redis_port, 0, redis_conns, redis_timeout, redis_keepalive);

  if (mongodb_client_pool == nullptr) {
    return EXIT_FAILURE;
  }

  ClientPool<ThriftClient<PostStorageServiceClient>>
      post_storage_client_pool("post-storage-client", post_storage_addr,
                               post_storage_port, 0, post_storage_conns, post_storage_timeout, post_storage_keepalive);

  mongoc_client_t *mongodb_client = mongoc_client_pool_pop(mongodb_client_pool);
  if (!mongodb_client) {
    LOG(fatal) << "Failed to pop mongoc client";
    return EXIT_FAILURE;
  }
  bool r = false;
  while (!r) {
    r = CreateIndex(mongodb_client, "user-timeline", "user_id", true);
    if (!r) {
      LOG(error) << "Failed to create mongodb index, try again";
      sleep(1);
    }
  }
  mongoc_client_pool_push(mongodb_client_pool, mongodb_client);

  TThreadedServer server (
      std::make_shared<UserTimelineServiceProcessor>(
          std::make_shared<UserTimelineHandler>(
              &redis_client_pool, mongodb_client_pool,
              &post_storage_client_pool)),
      std::make_shared<TServerSocket>("0.0.0.0", port),
      std::make_shared<TFramedTransportFactory>(),
      std::make_shared<TBinaryProtocolFactory>()
  );

  LOG(info) << "Starting the user-timeline-service server...";
  server.serve();


}