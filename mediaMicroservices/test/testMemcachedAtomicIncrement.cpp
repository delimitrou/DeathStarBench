#include <thread>
#include <iostream>
#include <libmemcached/memcached.h>
#include <libmemcached/util.h>

using namespace std;

void worker (memcached_pool_st *client_pool) {
  memcached_return_t memcached_rc;
  auto client = memcached_pool_pop(client_pool, true, &memcached_rc);
  uint64_t value;
  memcached_add(client, "key", 3, "0", 1, 0, 0);
  for (int i = 0; i < 1000; i++) {
    memcached_increment (
        client,
        "key",
        3,
        1,
        &value
    );
  }
  memcached_pool_push(client_pool, client);
}

int main() {
  for (int j = 0; j < 100; j++){
    string addr = "ath-8.ece.cornell.edu";
    int port = 20000;
    string config_str = "--SERVER=" + addr + ":" + std::to_string(port);
    auto memcached_client = memcached(config_str.c_str(), config_str.length());
    memcached_behavior_set(memcached_client, MEMCACHED_BEHAVIOR_NO_BLOCK, 1);
    memcached_behavior_set(memcached_client, MEMCACHED_BEHAVIOR_TCP_NODELAY, 1);
    memcached_behavior_set(
        memcached_client, MEMCACHED_BEHAVIOR_BINARY_PROTOCOL, 1);
    auto client_pool = memcached_pool_create(memcached_client, 8, 1024);
    thread *threads[10];
    for (auto &i : threads) {
      i = new thread(worker, client_pool);
    }
    for (auto &i : threads) {
      i->join();
    }
    for (auto &i : threads) {
      delete i;
    }

    memcached_return_t memcached_rc;
    auto client = memcached_pool_pop(client_pool, true, &memcached_rc);
    size_t size;
    uint32_t flags;
    string res = memcached_get(client, "key", 3, &size, &flags, nullptr);
    cout << res << endl;
    memcached_delete (client, "key", 3, 0);
    memcached_pool_push(client_pool, client);
    memcached_pool_destroy(client_pool);
  }



  return 0;
}
