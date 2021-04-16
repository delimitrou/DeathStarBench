#ifndef SOCIAL_NETWORK_MICROSERVICES_GENERICCLIENT_H
#define SOCIAL_NETWORK_MICROSERVICES_GENERICCLIENT_H

#include <string>
#include <chrono>

namespace social_network {

class GenericClient{
 public:
  virtual ~GenericClient() = default;
  virtual void Connect() = 0;
  virtual void Disconnect() = 0;
  virtual bool IsConnected() = 0;

  long _connect_timestamp;
  long _keepalive_ms;

 protected:
  std::string _addr;
  int _port;
};

} // namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_GENERICCLIENT_H
