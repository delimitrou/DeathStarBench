/*
 * 64-bit Unique Id Generator
 *
 * ------------------------------------------------------------------------
 * | 12 bit machine ID |       40-bit timestamp          | 12-bit counter |
 * ------------------------------------------------------------------------
 *
 * 12-bit machine Id code by hasing the MAC address
 * 40-bit UNIX timestamp in millisecond precision with custom epoch
 * 12 bit counter which increases monotonically on single process
 *
 */

#include <signal.h>

#include <thrift/server/TThreadedServer.h>
#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/transport/TServerSocket.h>
#include <thrift/transport/TBufferTransports.h>

#include "../utils.h"
#include "UniqueIdHandler.h"

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

  SetUpTracer("config/jaeger-config.yml", "unique-id-service");

  json config_json;
  if (load_config_file("config/service-config.json", &config_json) != 0) {
    exit(EXIT_FAILURE);
  }

//  std::string addr = config_json["UniqueIdService"]["addr"];
  int port = config_json["unique-id-service"]["port"];

  std::string compose_addr = config_json["compose-review-service"]["addr"];
  int compose_port = config_json["compose-review-service"]["port"];

  std::string machine_id;
  if (GetMachineId(&machine_id) != 0) {
    exit(EXIT_FAILURE);
  }

  std::mutex thread_lock;
  ClientPool<ThriftClient<ComposeReviewServiceClient>> compose_client_pool(
      "compose-review-client", compose_addr, compose_port, 0, 128, 1000);

  TThreadedServer server (
      std::make_shared<UniqueIdServiceProcessor>(
          std::make_shared<UniqueIdHandler>(
              &thread_lock, machine_id, &compose_client_pool)),
      std::make_shared<TServerSocket>("0.0.0.0", port),
      std::make_shared<TFramedTransportFactory>(),
      std::make_shared<TBinaryProtocolFactory>()
  );

  std::cout << "Starting the unique-id-service server ..." << std::endl;
  server.serve();
}
