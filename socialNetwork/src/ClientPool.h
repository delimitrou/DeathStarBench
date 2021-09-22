#ifndef SOCIAL_NETWORK_MICROSERVICES_CLIENTPOOL_H
#define SOCIAL_NETWORK_MICROSERVICES_CLIENTPOOL_H

#include <vector>
#include <mutex>
#include <condition_variable>
#include <deque>
#include <chrono>
#include <string>
#include <nlohmann/json.hpp>

#include "logger.h"

namespace social_network {
using json = nlohmann::json;

template<class TClient>
class ClientPool {
 public:
  ClientPool(const std::string &client_type, const std::string &addr,
      int port, int min_size, int max_size, int timeout_ms, int keepalive_ms,
      const json &config_json);
  ~ClientPool();

  ClientPool(const ClientPool&) = delete;
  ClientPool& operator=(const ClientPool&) = delete;
  ClientPool(ClientPool&&) = default;
  ClientPool& operator=(ClientPool&&) = default;

  TClient * Pop();
  void Push(TClient *);
  void Keepalive(TClient *);
  void Remove(TClient *);

 private:
  std::deque<TClient *> _pool;
  std::string _addr;
  std::string _client_type;
  int _port;
  int _min_pool_size{};
  int _max_pool_size{};
  int _curr_pool_size{};
  int _timeout_ms;
  int _keepalive_ms;
  std::mutex _mtx;
  std::condition_variable _cv;
  const json *_config_json;

};

template<class TClient>
ClientPool<TClient>::ClientPool(const std::string &client_type,
    const std::string &addr, int port, int min_pool_size,
    int max_pool_size, int timeout_ms, int keepalive_ms,
    const json &config_json) {
  _addr = addr;
  _port = port;
  _min_pool_size = min_pool_size;
  _max_pool_size = max_pool_size;
  _timeout_ms = timeout_ms;
  _client_type = client_type;
  _keepalive_ms = keepalive_ms;
  _config_json = &config_json;

  for (int i = 0; i < min_pool_size; ++i) {
    TClient *client = new TClient(addr, port, keepalive_ms, config_json);
    _pool.emplace_back(client);
  }
  _curr_pool_size = min_pool_size;
}

template<class TClient>
ClientPool<TClient>::~ClientPool() {
  while (!_pool.empty()) {
    delete _pool.front();
    _pool.pop_front();
  }
}

template<class TClient>
TClient * ClientPool<TClient>::Pop() {
  TClient * client = nullptr;
  {
    std::unique_lock<std::mutex> cv_lock(_mtx);
    while (_pool.size() == 0 && _curr_pool_size == _max_pool_size) {
      // Create a new a client if current pool size is less than
      // the max pool size.
      auto wait_time = std::chrono::system_clock::now() +
          std::chrono::milliseconds(_timeout_ms);
      bool wait_success = _cv.wait_until(cv_lock, wait_time,
            [this] { return _pool.size() > 0 || _curr_pool_size < _max_pool_size; });
      if (!wait_success) {
        LOG(warning) << "ClientPool pop timeout";
        LOG(info) << _pool.size() << " " << _curr_pool_size;
        cv_lock.unlock();
        return nullptr;
      }
    }
    if (_pool.size() > 0) {
      client = _pool.front();
      _pool.pop_front();
    } else {
      client = new TClient(_addr, _port, _keepalive_ms, *_config_json);
      _curr_pool_size++;
    }
  cv_lock.unlock();
  } // cv_lock(_mtx)


  if (client) {
    try {
      client->Connect();
    } catch (...) {
      LOG(error) << "Failed to connect " + _client_type;
      Remove(client);
      throw;
    }
  }
  return client;
}

template<class TClient>
void ClientPool<TClient>::Push(TClient *client) {
  std::unique_lock<std::mutex> cv_lock(_mtx);
  _pool.push_back(client);
  cv_lock.unlock();
  _cv.notify_one();
}

template<class TClient>
void ClientPool<TClient>::Remove(TClient *client) {
  // No need to delete it from _pool because the *client has been poped out
  delete client;
  std::unique_lock<std::mutex> cv_lock(_mtx);
  _curr_pool_size--;
  cv_lock.unlock();
  _cv.notify_one();
}

template<class TClient>
void ClientPool<TClient>::Keepalive(TClient *client) {
  long curr_timestamp = std::chrono::duration_cast<std::chrono::milliseconds>(
          std::chrono::system_clock::now().time_since_epoch()).count();
  if (curr_timestamp - client->_connect_timestamp > client->_keepalive_ms) {
    Remove(client);
  } else {
    Push(client);
  }
}

} // namespace social_network


#endif //SOCIAL_NETWORK_MICROSERVICES_CLIENTPOOL_H