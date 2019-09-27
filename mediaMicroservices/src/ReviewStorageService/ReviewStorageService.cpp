#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TThreadedServer.h>
#include <thrift/transport/TServerSocket.h>
#include <thrift/transport/TBufferTransports.h>
#include "nlohmann/json.hpp"
#include <signal.h>

#include "../utils.h"
#include "../utils_mongodb.h"
#include "../utils_memcached.h"
#include "ReviewStorageHandler.h"

using apache::thrift::server::TThreadedServer;
using apache::thrift::transport::TServerSocket;
using apache::thrift::transport::TFramedTransportFactory;
using apache::thrift::protocol::TBinaryProtocolFactory;
using namespace media_service;

static memcached_pool_st* memcached_client_pool;
static mongoc_client_pool_t* mongodb_client_pool;

void sigintHandler(int sig) {
  if (memcached_client_pool != nullptr) {
    memcached_pool_destroy(memcached_client_pool);
  }
  if (mongodb_client_pool != nullptr) {
    mongoc_client_pool_destroy(mongodb_client_pool);
  }
  exit(EXIT_SUCCESS);
}

int main(int argc, char *argv[]) {
  signal(SIGINT, sigintHandler);

  init_logger();

  SetUpTracer("config/jaeger-config.yml", "review-storage-service");

  json config_json;
  if (load_config_file("config/service-config.json", &config_json) != 0) {
    exit(EXIT_FAILURE);
  }

  int port = config_json["review-storage-service"]["port"];

  memcached_client_pool =
      init_memcached_client_pool(config_json, "review-storage",
          MEMCACHED_POOL_MIN_SIZE, MEMCACHED_POOL_MAX_SIZE);
  mongodb_client_pool = init_mongodb_client_pool(config_json, "review-storage",
      MONGODB_POOL_MAX_SIZE);

  if (memcached_client_pool == nullptr || mongodb_client_pool == nullptr) {
    return EXIT_FAILURE;
  }

  TThreadedServer server (
      std::make_shared<ReviewStorageServiceProcessor>(
          std::make_shared<ReviewStorageHandler>(
              memcached_client_pool, mongodb_client_pool)),
      std::make_shared<TServerSocket>("0.0.0.0", port),
      std::make_shared<TFramedTransportFactory>(),
      std::make_shared<TBinaryProtocolFactory>()
  );

  std::cout << "Starting the review-storage-service server..." << std::endl;
  server.serve();
}