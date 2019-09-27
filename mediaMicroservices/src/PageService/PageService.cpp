#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/server/TThreadedServer.h>
#include <thrift/transport/TServerSocket.h>
#include <thrift/transport/TBufferTransports.h>
#include <signal.h>

#include "../utils.h"
#include "PageHandler.h"

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

  SetUpTracer("config/jaeger-config.yml", "cast-info-service");

  json config_json;
  if (load_config_file("config/service-config.json", &config_json) != 0) {
    exit(EXIT_FAILURE);
  }

  int port = config_json["page-service"]["port"];
  std::string cast_info_addr = config_json["cast-info-service"]["addr"];
  int cast_info_port = config_json["cast-info-service"]["port"];
  std::string movie_review_addr = config_json["movie-review-service"]["addr"];
  int movie_review_port = config_json["movie-review-service"]["port"];
  std::string movie_info_addr = config_json["movie-info-service"]["addr"];
  int movie_info_port = config_json["movie-info-service"]["port"];
  std::string plot_addr = config_json["plot-service"]["addr"];
  int plot_port = config_json["plot-service"]["port"];

  ClientPool<ThriftClient<MovieInfoServiceClient>>
      movie_info_client_pool("movie-info-client", movie_info_addr,
                             movie_info_port, 0, 128, 1000);
  ClientPool<ThriftClient<CastInfoServiceClient>>
      cast_info_client_pool("cast-info-client", cast_info_addr,
                            cast_info_port, 0, 128, 1000);
  ClientPool<ThriftClient<MovieReviewServiceClient>>
      movie_review_client_pool("movie-review-client", movie_review_addr,
                               movie_review_port, 0, 128, 1000);
  ClientPool<ThriftClient<PlotServiceClient>>
      plot_client_pool("plot-client", plot_addr, plot_port, 0, 128, 1000);

  TThreadedServer server(
      std::make_shared<PageServiceProcessor>(
          std::make_shared<PageHandler>(
              &movie_review_client_pool,
              &movie_info_client_pool,
              &cast_info_client_pool,
              &plot_client_pool)),
      std::make_shared<TServerSocket>("0.0.0.0", port),
      std::make_shared<TFramedTransportFactory>(),
      std::make_shared<TBinaryProtocolFactory>()
  );
  std::cout << "Starting the page-service server ..." << std::endl;
  server.serve();
}