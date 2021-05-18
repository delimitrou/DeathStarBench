#include <signal.h>
#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TThreadedServer.h>
#include <thrift/transport/TBufferTransports.h>
#include <thrift/transport/TServerSocket.h>

#include "../utils.h"
#include "../utils_thrift.h"
#include "MediaHandler.h"

using apache::thrift::protocol::TBinaryProtocolFactory;
using apache::thrift::server::TThreadedServer;
using apache::thrift::transport::TFramedTransportFactory;
using apache::thrift::transport::TServerSocket;
using namespace social_network;

void sigintHandler(int sig) { exit(EXIT_SUCCESS); }

int main(int argc, char *argv[]) {
  signal(SIGINT, sigintHandler);
  init_logger();
  SetUpTracer("config/jaeger-config.yml", "media-service");
  json config_json;
  if (load_config_file("config/service-config.json", &config_json) != 0) {
    exit(EXIT_FAILURE);
  }

  int port = config_json["media-service"]["port"];
  std::shared_ptr<TServerSocket> server_socket = get_server_socket(config_json, "0.0.0.0", port);

  TThreadedServer server(
      std::make_shared<MediaServiceProcessor>(std::make_shared<MediaHandler>()),
      server_socket,
      std::make_shared<TFramedTransportFactory>(),
      std::make_shared<TBinaryProtocolFactory>());

  LOG(info) << "Starting the media-service server...";
  server.serve();
}
