#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TThreadedServer.h>
#include <thrift/transport/TServerSocket.h>
#include <thrift/transport/TBufferTransports.h>
#include <signal.h>

#include "UserReviewHandler.h"
#include "../utils.h"
#include "../utils_mongodb.h"

using apache::thrift::server::TThreadedServer;
using apache::thrift::transport::TServerSocket;
using apache::thrift::transport::TFramedTransportFactory;
using apache::thrift::protocol::TBinaryProtocolFactory;
using media_service::UserReviewHandler;
using namespace media_service;

void sigintHandler(int sig) {
  exit(EXIT_SUCCESS);
}

int main(int argc, char *argv[]) {
  signal(SIGINT, sigintHandler);
  init_logger();

  SetUpTracer("config/jaeger-config.yml", "user-review-service");

  json config_json;
  if (load_config_file("config/service-config.json", &config_json) != 0) {
    LOG(fatal) << "Cannot open the config file.";
    exit(EXIT_FAILURE);
  }

  int port = config_json["user-review-service"]["port"];
  std::string redis_addr =
      config_json["user-review-redis"]["addr"];
  int redis_port = config_json["user-review-redis"]["port"];
  int review_storage_port = config_json["review-storage-service"]["port"];
  std::string review_storage_addr = config_json["review-storage-service"]["addr"];

  mongoc_client_pool_t *mongodb_client_pool =
      init_mongodb_client_pool(config_json, "user-review", 128);
  ClientPool<RedisClient> redis_client_pool("user-review-redis",
                                            redis_addr, redis_port, 0, 128, 1000);
  ClientPool<ThriftClient<ReviewStorageServiceClient>>
      review_storage_client_pool("review-storage-client", review_storage_addr,
                                 review_storage_port, 0, 128, 1000);

  if (mongodb_client_pool == nullptr) {
    return EXIT_FAILURE;
  }

  mongoc_client_t *mongodb_client = mongoc_client_pool_pop(mongodb_client_pool);
  if (!mongodb_client) {
    LOG(fatal) << "Failed to pop mongoc client";
    return EXIT_FAILURE;
  }
  bool r = false;
  while (!r) {
    r = CreateIndex(mongodb_client, "user-review", "user_id", true);
    if (!r) {
      LOG(error) << "Failed to create mongodb index, try again";
      sleep(1);
    }
  }
  mongoc_client_pool_push(mongodb_client_pool, mongodb_client);

  TThreadedServer server(
      std::make_shared<UserReviewServiceProcessor>(
          std::make_shared<UserReviewHandler>(
              &redis_client_pool,
              mongodb_client_pool,
              &review_storage_client_pool)),
      std::make_shared<TServerSocket>("0.0.0.0", port),
      std::make_shared<TFramedTransportFactory>(),
      std::make_shared<TBinaryProtocolFactory>()
  );
  std::cout << "Starting the user-review-service server ..." << std::endl;
  server.serve();

}