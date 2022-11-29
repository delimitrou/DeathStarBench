#include <signal.h>
#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TThreadedServer.h>
#include <thrift/transport/TBufferTransports.h>
#include <thrift/transport/TServerSocket.h>

#include <boost/program_options.hpp>

#include "../../gen-cpp/social_network_types.h"
#include "../ClientPool.h"
#include "../logger.h"
#include "../tracing.h"
#include "../utils.h"
#include "../utils_mongodb.h"
#include "../utils_redis.h"
#include "../utils_thrift.h"
#include "UserTimelineHandler.h"

using apache::thrift::protocol::TBinaryProtocolFactory;
using apache::thrift::server::TThreadedServer;
using apache::thrift::transport::TFramedTransportFactory;
using apache::thrift::transport::TServerSocket;
using namespace social_network;

void sigintHandler(int sig) { exit(EXIT_SUCCESS); }

int main(int argc, char *argv[]) {
  signal(SIGINT, sigintHandler);
  init_logger();

  // Command line options
  namespace po = boost::program_options;
  po::options_description desc("Options");
  desc.add_options()("help", "produce help message")(
      "redis-cluster",
      po::value<bool>()->default_value(false)->implicit_value(true),
      "Enable redis cluster mode");

  po::variables_map vm;
  po::store(po::parse_command_line(argc, argv, desc), vm);
  po::notify(vm);

  if (vm.count("help")) {
    std::cout << desc << "\n";
    return 0;
  }

  bool redis_cluster_flag = false;
  if (vm.count("redis-cluster")) {
    if (vm["redis-cluster"].as<bool>()) {
      redis_cluster_flag = true;
    }
  }

  SetUpTracer("config/jaeger-config.yml", "user-timeline-service");

  json config_json;
  if (load_config_file("config/service-config.json", &config_json) != 0) {
    exit(EXIT_FAILURE);
  }

  int port = config_json["user-timeline-service"]["port"];

  int post_storage_port = config_json["post-storage-service"]["port"];
  std::string post_storage_addr = config_json["post-storage-service"]["addr"];
  int post_storage_conns = config_json["post-storage-service"]["connections"];
  int post_storage_timeout = config_json["post-storage-service"]["timeout_ms"];
  int post_storage_keepalive =
      config_json["post-storage-service"]["keepalive_ms"];

  int mongodb_conns = config_json["user-timeline-mongodb"]["connections"];
  int mongodb_timeout = config_json["user-timeline-mongodb"]["timeout_ms"];

  int redis_cluster_config_flag = config_json["user-timeline-redis"]["use_cluster"];
  int redis_replica_config_flag = config_json["user-timeline-redis"]["use_replica"];

  auto mongodb_client_pool =
      init_mongodb_client_pool(config_json, "user-timeline", mongodb_conns);

  if (mongodb_client_pool == nullptr) {
    return EXIT_FAILURE;
  }

  if (redis_replica_config_flag && (redis_cluster_config_flag || redis_cluster_flag)) {
      LOG(error) << "Can't start service when Redis Cluster and Redis Replica are enabled at the same time";
      exit(EXIT_FAILURE);
  }

  ClientPool<ThriftClient<PostStorageServiceClient>> post_storage_client_pool(
      "post-storage-client", post_storage_addr, post_storage_port, 0,
      post_storage_conns, post_storage_timeout, post_storage_keepalive,
      config_json);

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
  std::shared_ptr<TServerSocket> server_socket =
      get_server_socket(config_json, "0.0.0.0", port);

  if (redis_cluster_flag || redis_cluster_config_flag) {
    RedisCluster redis_client_pool =
        init_redis_cluster_client_pool(config_json, "user-timeline");
    TThreadedServer server(std::make_shared<UserTimelineServiceProcessor>(
                               std::make_shared<UserTimelineHandler>(
                                   &redis_client_pool, mongodb_client_pool,
                                   &post_storage_client_pool)),
                           server_socket,
                           std::make_shared<TFramedTransportFactory>(),
                           std::make_shared<TBinaryProtocolFactory>());
    LOG(info) << "Starting the user-timeline-service server with Redis Cluster support...";
    server.serve();
  }
  else if (redis_replica_config_flag) {
      Redis redis_replica_client_pool = init_redis_replica_client_pool(config_json, "redis-replica");
      Redis redis_primary_client_pool = init_redis_replica_client_pool(config_json, "redis-primary");
      TThreadedServer server(std::make_shared<UserTimelineServiceProcessor>(
          std::make_shared<UserTimelineHandler>(
              &redis_replica_client_pool, &redis_primary_client_pool, mongodb_client_pool,
              &post_storage_client_pool)),
          server_socket,
          std::make_shared<TFramedTransportFactory>(),
          std::make_shared<TBinaryProtocolFactory>());
      LOG(info) << "Starting the user-timeline-service server with replicated Redis support...";
      server.serve();

  }
  else {
    Redis redis_client_pool =
        init_redis_client_pool(config_json, "user-timeline");
    TThreadedServer server(std::make_shared<UserTimelineServiceProcessor>(
                               std::make_shared<UserTimelineHandler>(
                                   &redis_client_pool, mongodb_client_pool,
                                   &post_storage_client_pool)),
                           server_socket,
                           std::make_shared<TFramedTransportFactory>(),
                           std::make_shared<TBinaryProtocolFactory>());
    LOG(info) << "Starting the user-timeline-service server...";
    server.serve();
  }
}