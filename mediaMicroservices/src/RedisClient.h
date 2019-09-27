#ifndef MEDIA_MICROSERVICES_REDISCLIENT_H
#define MEDIA_MICROSERVICES_REDISCLIENT_H

#include <string>
#include <cpp_redis/cpp_redis>

#include "logger.h"
#include "GenericClient.h"

namespace media_service {

class RedisClient : public GenericClient {
 public:
  RedisClient(const std::string &addr, int port);
  RedisClient(const RedisClient &) = delete;
  RedisClient & operator=(const RedisClient &) = delete;
  RedisClient(RedisClient &&) = default;
  RedisClient & operator=(RedisClient &&) = default;

  ~RedisClient() override ;

  cpp_redis::client *GetClient() const;

  void Connect() override ;
  void Disconnect() override ;
  void KeepAlive() override ;
  void KeepAlive(int timeout_ms) override ;
  bool IsConnected() override ;

 private:
  cpp_redis::client * _client;
};

RedisClient::RedisClient(const std::string &addr, int port) {
  _addr = addr;
  _port = port;
  _client = new cpp_redis::client();
}

RedisClient::~RedisClient() {
  Disconnect();
  delete _client;
}

cpp_redis::client* RedisClient::GetClient() const {
  return _client;
}

void RedisClient::Connect() {
  if (!IsConnected()) {
    _client->connect(_addr, _port, [](const std::string& host, std::size_t port,
        cpp_redis::client::connect_state status) {
      if (status == cpp_redis::client::connect_state::dropped) {
        LOG(error) << "Failed to connect " << host << ":" << port;
        throw status;
      }
    });
  }
}

void RedisClient::Disconnect() {
  if (IsConnected()) {
    _client->disconnect();
  }
}

bool RedisClient::IsConnected() {
  return _client->is_connected();
}

void RedisClient::KeepAlive() {

}

void RedisClient::KeepAlive(int timeout_ms) {

}

} // mediua_service

#endif //MEDIA_MICROSERVICES_REDISCLIENT_H
