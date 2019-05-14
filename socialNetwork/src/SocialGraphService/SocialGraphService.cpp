#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TThreadedServer.h>
#include <thrift/transport/TServerSocket.h>
#include <thrift/transport/TBufferTransports.h>
#include <signal.h>

#include "../utils.h"
#include "../utils_mongodb.h"
#include "SocialGraphHandler.h"

using json = nlohmann::json;
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

  SetUpTracer("config/jaeger-config.yml", "social-graph-service");

  json config_json;
  if (load_config_file("config/service-config.json", &config_json) != 0) {
    exit(EXIT_FAILURE);
  }

  int port = config_json["social-graph-service"]["port"];
  int redis_port = config_json["social-graph-redis"]["port"];
  std::string redis_addr = config_json["social-graph-redis"]["addr"];
  std::string user_addr = config_json["user-service"]["addr"];
  int user_port = config_json["user-service"]["port"];

  mongoc_client_pool_t *mongodb_client_pool =
      init_mongodb_client_pool(config_json, "social-graph", 128);

  if (mongodb_client_pool == nullptr) {
    return EXIT_FAILURE;
  }
  ClientPool<RedisClient> redis_client_pool("redis", redis_addr, redis_port,
      0, 128, 1000);
  ClientPool<ThriftClient<UserServiceClient>> user_client_pool(
      "social-graph", user_addr, user_port, 0, 128, 1000);

  mongoc_client_t *mongodb_client = mongoc_client_pool_pop(mongodb_client_pool);
  if (!mongodb_client) {
    LOG(fatal) << "Failed to pop mongoc client";
    return EXIT_FAILURE;
  }
  bool r = false;
  while (!r) {
    r = CreateIndex(mongodb_client, "social-graph", "user_id", true);
    if (!r) {
      LOG(error) << "Failed to create mongodb index, try again";
      sleep(1);
    }
  }
  mongoc_client_pool_push(mongodb_client_pool, mongodb_client);

  TThreadedServer server(
      std::make_shared<SocialGraphServiceProcessor>(
          std::make_shared<SocialGraphHandler>(
              mongodb_client_pool,
              &redis_client_pool,
              &user_client_pool)),
      std::make_shared<TServerSocket>("0.0.0.0", port),
      std::make_shared<TFramedTransportFactory>(),
      std::make_shared<TBinaryProtocolFactory>()
  );

  std::cout << "Starting the social-graph-service server ..." << std::endl;
  server.serve();
}

