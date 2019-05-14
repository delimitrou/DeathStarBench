#include <signal.h>

#include <thrift/server/TThreadedServer.h>
#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/transport/TServerSocket.h>
#include <thrift/transport/TBufferTransports.h>

#include "../utils.h"
#include "TextHandler.h"

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
  SetUpTracer("config/jaeger-config.yml", "text-service");

  json config_json;
  if (load_config_file("config/service-config.json", &config_json) == 0) {

    int port = config_json["text-service"]["port"];
    std::string compose_addr = config_json["compose-post-service"]["addr"];
    int compose_port = config_json["compose-post-service"]["port"];

    std::string url_addr = config_json["url-shorten-service"]["addr"];
    int url_port = config_json["url-shorten-service"]["port"];

    std::string user_mention_addr = config_json["user-mention-service"]["addr"];
    int user_mention_port = config_json["user-mention-service"]["port"];

    ClientPool<ThriftClient<ComposePostServiceClient>> compose_client_pool(
        "compose-post", compose_addr, compose_port, 0, 128, 1000);

    ClientPool<ThriftClient<UrlShortenServiceClient>> url_client_pool(
        "url-shorten-service", url_addr, url_port, 0, 128, 1000);

    ClientPool<ThriftClient<UserMentionServiceClient>> user_mention_pool(
        "user-mention-service", user_mention_addr,
        user_mention_port, 0, 128, 1000);

    TThreadedServer server(
        std::make_shared<TextServiceProcessor>(
            std::make_shared<TextHandler>(
                &compose_client_pool,
                &url_client_pool,
                &user_mention_pool)),
        std::make_shared<TServerSocket>("0.0.0.0", port),
        std::make_shared<TFramedTransportFactory>(),
        std::make_shared<TBinaryProtocolFactory>()
    );

    std::cout << "Starting the text-service server..." << std::endl;
    server.serve();
  } else exit(EXIT_FAILURE);
}


