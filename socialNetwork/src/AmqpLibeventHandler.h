#ifndef SOCIAL_NETWORK_MICROSERVICES_SRC_AMQPLIBEVENTHANDLER_H_
#define SOCIAL_NETWORK_MICROSERVICES_SRC_AMQPLIBEVENTHANDLER_H_

#include <functional>
#include <unistd.h>
#include <amqpcpp.h>
#include <event2/event.h>
#include <amqpcpp/libevent.h>

namespace social_network {

class LibEventHandler : public AMQP::LibEventHandler {
 public:
  LibEventHandler(struct event_base* evbase) : AMQP::LibEventHandler(evbase),
                                               evbase_(evbase) {}
  void onError(AMQP::TcpConnection *connection, const char *message) override
  {
    std::cout << "Error: " << message << std::endl;
    event_base_loopbreak(evbase_);
  }
 private:
  struct event_base* evbase_ {nullptr};
};

class AmqpLibeventHandler {
 public:
  using EventBasePtrT = std::unique_ptr<struct event_base,
                                        std::function<void(struct event_base*)> >;
  using EventPtrT = std::unique_ptr<struct event,
                                    std::function<void(struct event*)> >;

  AmqpLibeventHandler()
      : evbase_(event_base_new(), event_base_free)
      , evhandler_(evbase_.get())
  {
    is_running_ = false;
  }

  void Start()
  {
    is_running_ = true;
    event_base_dispatch(evbase_.get());
  }
  void Stop()
  {
    is_running_ = false;
    event_base_loopbreak(evbase_.get());
  }

  operator AMQP::TcpHandler* ()
  {
    return &evhandler_;
  }

  bool GetIsRunning() {
    return is_running_;
  }

 private:
  EventBasePtrT evbase_;
  LibEventHandler evhandler_;
  bool is_running_;

};


} // namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_SRC_AMQPLIBEVENTHANDLER_H_
