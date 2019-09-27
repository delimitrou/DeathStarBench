#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TThreadedServer.h>
#include <thrift/transport/TServerSocket.h>
#include <thrift/transport/TBufferTransports.h>
#include <signal.h>

#include "ComposeReviewHandler.h"
#include "../utils.h"
#include "../utils_memcached.h"

using json = nlohmann::json;
using apache::thrift::server::TThreadedServer;
using apache::thrift::transport::TServerSocket;
using apache::thrift::transport::TFramedTransportFactory;
using apache::thrift::protocol::TBinaryProtocolFactory;
using namespace media_service;

void sigintHandler(int sig) {
  exit(EXIT_SUCCESS);
}

int main(int argc, char *argv[]) {
  signal(SIGINT, sigintHandler);
  init_logger();

  SetUpTracer("config/jaeger-config.yml", "compose-review-service");

  json config_json;
  if (load_config_file("config/service-config.json", &config_json) != 0) {
    exit(EXIT_FAILURE);
  }

  int port = config_json["compose-review-service"]["port"];
  std::string review_storage_addr =
      config_json["review-storage-service"]["addr"];
  int review_storage_port = config_json["review-storage-service"]["port"];

  std::string user_review_addr = config_json["user-review-service"]["addr"];
  int user_review_port = config_json["user-review-service"]["port"];

  std::string movie_review_addr = config_json["movie-review-service"]["addr"];
  int movie_review_port = config_json["movie-review-service"]["port"];

  ClientPool<ThriftClient<ReviewStorageServiceClient>> compose_client_pool(
      "compose-review-service", review_storage_addr, review_storage_port, 0, 128, 1000);
  ClientPool<ThriftClient<UserReviewServiceClient>> user_client_pool(
      "user-review-service", user_review_addr, user_review_port, 0, 128, 1000);
  ClientPool<ThriftClient<MovieReviewServiceClient>> movie_client_pool(
      "movie-review-service", movie_review_addr, movie_review_port, 0, 128, 1000);


  std::string mmc_addr = config_json["compose-review-memcached"]["addr"];
  int mmc_port = config_json["compose-review-memcached"]["port"];
  std::string config_str = "--SERVER=" + mmc_addr + ":" + std::to_string(mmc_port);
  auto memcached_client = memcached(config_str.c_str(), config_str.length());
  memcached_behavior_set(memcached_client, MEMCACHED_BEHAVIOR_NO_BLOCK, 1);
  memcached_behavior_set(memcached_client, MEMCACHED_BEHAVIOR_TCP_NODELAY, 1);
  memcached_behavior_set(
      memcached_client, MEMCACHED_BEHAVIOR_BINARY_PROTOCOL, 1);
  auto memcached_client_pool = memcached_pool_create(
      memcached_client, MEMCACHED_POOL_MIN_SIZE, MEMCACHED_POOL_MAX_SIZE);

  TThreadedServer server(
      std::make_shared<ComposeReviewServiceProcessor>(
          std::make_shared<ComposeReviewHandler>(
              memcached_client_pool,
              &compose_client_pool,
              &user_client_pool,
              &movie_client_pool)),
      std::make_shared<TServerSocket>("0.0.0.0", port),
      std::make_shared<TFramedTransportFactory>(),
      std::make_shared<TBinaryProtocolFactory>()
  );
  std::cout << "Starting the compose-review-service server ..." << std::endl;
  server.serve();
}




