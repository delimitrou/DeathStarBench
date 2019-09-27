#include "../src/ClientPool.h"
#include "../src/ThriftClient.h"
#include "../src/ComposeReviewHandler.h"
#include "../src/logger.h"

#include <chrono>
#include <thread>
#include <vector>


using namespace media_service;

void getClient(
    ClientPool<ThriftClient<ComposeReviewServiceClient>> *client_pool,
    int sleep_time) {
  while (true) {
    auto client = client_pool->Pop();
    LOG(debug) << "Pop client " << client->getId();
    client_pool->Push(client->getId());
    LOG(debug) << "Push client " << client->getId();
  }
}

int main(int argc, char *argv[]) {
  init_logger();
  ClientPool<ThriftClient<ComposeReviewServiceClient>> client_pool(
      0, 16, "compose-review-service", 9090);
  bool done = false;
  std::vector<std::thread> threads;
  for (int i = 0; i < 10; i++) {
    threads.emplace_back(std::thread(getClient, &client_pool, 20));
  }
  for (int i = 0; i < 10; i++) {
    threads[i].join();
  }
}

