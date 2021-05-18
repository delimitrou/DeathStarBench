#ifndef SOCIAL_NETWORK_MICROSERVICES_THRIFTCLIENT_H
#define SOCIAL_NETWORK_MICROSERVICES_THRIFTCLIENT_H

#include <string>
#include <thread>
#include <iostream>
#include <chrono>
#include <boost/log/trivial.hpp>

#include <thrift/protocol/TBinaryProtocol.h>
#include <thrift/transport/TSocket.h>
#include <thrift/transport/TSSLSocket.h>
#include <thrift/transport/TTransportUtils.h>
#include <thrift/stdcxx.h>
#include <nlohmann/json.hpp>
#include "logger.h"
#include "GenericClient.h"


namespace social_network {

using apache::thrift::protocol::TProtocol;
using apache::thrift::protocol::TBinaryProtocol;
using apache::thrift::transport::TFramedTransport;
using apache::thrift::transport::TSocket;
using apache::thrift::transport::TSSLSocketFactory;
using apache::thrift::transport::TTransport;
using apache::thrift::TException;
using json = nlohmann::json;

template<class TThriftClient>
class ThriftClient : public GenericClient {
 public:
  ThriftClient(const std::string &addr, int port);
  ThriftClient(const std::string &addr, int port, int keepalive_ms, const json &config_json);

  ThriftClient(const ThriftClient &) = delete;
  ThriftClient &operator=(const ThriftClient &) = delete;
  ThriftClient(ThriftClient<TThriftClient> &&) = default;
  ThriftClient &operator=(ThriftClient &&) = default;

  ~ThriftClient() override;

  TThriftClient *GetClient() const;

  void Connect() override;
  void Disconnect() override;
  bool IsConnected() override;

 private:
  TThriftClient *_client;

  std::shared_ptr<TSocket> _socket;
  std::shared_ptr<TTransport> _transport;
  std::shared_ptr<TProtocol> _protocol;
};

template<class TThriftClient>
ThriftClient<TThriftClient>::ThriftClient(
    const std::string &addr, int port) {
  _addr = addr;
  _port = port;
  _socket = std::shared_ptr<TSocket>(new TSocket(addr, port));
  _socket->setKeepAlive(true);
  _transport = std::shared_ptr<TTransport>(new TFramedTransport(_socket));
  _protocol = std::shared_ptr<TProtocol>(new TBinaryProtocol(_transport));
  _client = new TThriftClient(_protocol);
  _connect_timestamp = 0;
  _keepalive_ms = 0;
}

template <class TThriftClient>
ThriftClient<TThriftClient>::ThriftClient(
    const std::string &addr, int port, int keepalive_ms, const json &config_json) {
  _addr = addr;
  _port = port;
  bool ssl_enabled = config_json["ssl"]["enabled"];

  if (ssl_enabled) {
    std::string ca_path = config_json["ssl"]["caPath"];
    std::string ciphers = config_json["ssl"]["ciphers"];

    std::shared_ptr<TSSLSocketFactory> factory;
    factory = std::make_shared<TSSLSocketFactory>();
    factory->ciphers(ciphers);
    factory->loadTrustedCertificates(ca_path.c_str());

    // if (config_json["ssl"]["verifyClient"]) {
    //   std::string cert_path = config_json["ssl"]["clientCertPath"];
    //   std::string key_path = config_json["ssl"]["clientKeyPath"];
    //   factory->loadCertificate(cert_path.c_str());
    //   factory->loadPrivateKey(key_path.c_str());
    // }
    // Need verify server
    factory->authenticate(true);
    _socket = factory->createSocket(addr, port);
  } else {
    _socket = std::shared_ptr<TSocket>(new TSocket(addr, port));
  }
  _socket->setKeepAlive(true);
  _transport = std::shared_ptr<TTransport>(new TFramedTransport(_socket));
  _protocol = std::shared_ptr<TProtocol>(new TBinaryProtocol(_transport));
  _client = new TThriftClient(_protocol);
  _connect_timestamp = std::chrono::duration_cast<std::chrono::milliseconds>(
                           std::chrono::system_clock::now().time_since_epoch())
                           .count();
  _keepalive_ms = keepalive_ms;
}

template<class TThriftClient>
ThriftClient<TThriftClient>::~ThriftClient() {
  Disconnect();
  delete _client;
}

template<class TThriftClient>
TThriftClient *ThriftClient<TThriftClient>::GetClient() const {
  return _client;
}

template<class TThriftClient>
bool ThriftClient<TThriftClient>::IsConnected() {
  return _transport->isOpen();
}

template<class TThriftClient>
void ThriftClient<TThriftClient>::Connect() {
  if (!IsConnected()) {
    try {
      _transport->open();
    } catch (TException &tx) {
      throw tx;
    }
  }
}

template<class TThriftClient>
void ThriftClient<TThriftClient>::Disconnect() {
  if (IsConnected()) {
    try {
      _transport->close();
    } catch (TException &tx) {
      throw tx;
    }
  }
}

} // namespace social_network


#endif //SOCIAL_NETWORK_MICROSERVICES_THRIFTCLIENT_H
