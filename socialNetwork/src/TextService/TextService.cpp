#include <signal.h>
#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TThreadedServer.h>
#include <thrift/transport/TBufferTransports.h>
#include <thrift/transport/TServerSocket.h>

#include "../utils.h"
#include "../utils_thrift.h"
#include "TextHandler.h"

using apache::thrift::protocol::TBinaryProtocolFactory;
using apache::thrift::server::TThreadedServer;
using apache::thrift::transport::TFramedTransportFactory;
using apache::thrift::transport::TServerSocket;
using namespace social_network;

void sigintHandler(int sig) { exit(EXIT_SUCCESS); }

int main(int argc, char *argv[]) {
  signal(SIGINT, sigintHandler);
  init_logger();
  SetUpTracer("config/jaeger-config.yml", "text-service");

  json config_json;
  if (load_config_file("config/service-config.json", &config_json) == 0) {
    int port = config_json["text-service"]["port"];

    std::string url_addr = config_json["url-shorten-service"]["addr"];
    int url_port = config_json["url-shorten-service"]["port"];
    int url_conns = config_json["url-shorten-service"]["connections"];
    int url_timeout = config_json["url-shorten-service"]["timeout_ms"];
    int url_keepalive = config_json["url-shorten-service"]["keepalive_ms"];

    std::string user_mention_addr = config_json["user-mention-service"]["addr"];
    int user_mention_port = config_json["user-mention-service"]["port"];
    int user_mention_conns = config_json["user-mention-service"]["connections"];
    int user_mention_timeout =
        config_json["user-mention-service"]["timeout_ms"];
    int user_mention_keepalive =
        config_json["user-mention-service"]["keepalive_ms"];

    ClientPool<ThriftClient<UrlShortenServiceClient>> url_client_pool(
        "url-shorten-service", url_addr, url_port, 0, url_conns, url_timeout,
        url_keepalive, config_json);

    ClientPool<ThriftClient<UserMentionServiceClient>> user_mention_pool(
        "user-mention-service", user_mention_addr, user_mention_port, 0,
        user_mention_conns, user_mention_timeout, user_mention_keepalive, config_json);

    std::shared_ptr<TServerSocket> server_socket = get_server_socket(config_json, "0.0.0.0", port);
    TThreadedServer server(
        std::make_shared<TextServiceProcessor>(std::make_shared<TextHandler>(
            &url_client_pool, &user_mention_pool)),
        server_socket,
        std::make_shared<TFramedTransportFactory>(),
        std::make_shared<TBinaryProtocolFactory>());

    LOG(info) << "Starting the text-service server...";
    server.serve();
  } else
    exit(EXIT_FAILURE);
}
