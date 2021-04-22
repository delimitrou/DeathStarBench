#ifndef SOCIAL_NETWORK_MICROSERVICES_LOGGER_H
#define SOCIAL_NETWORK_MICROSERVICES_LOGGER_H

#include <boost/log/trivial.hpp>
#include <boost/log/utility/setup/console.hpp>
#include <boost/log/utility/setup/common_attributes.hpp>

#include <string.h>

namespace social_network {
#define __FILENAME__ \
    (strrchr(__FILE__, '/') ? strrchr(__FILE__, '/') + 1 : __FILE__)
#define LOG(severity) \
    BOOST_LOG_TRIVIAL(severity) << "(" << __FILENAME__ << ":" \
    << __LINE__ << ":" << __FUNCTION__ << ") "

void init_logger() {
  boost::log::register_simple_formatter_factory
      <boost::log::trivial::severity_level, char>("Severity");
  boost::log::add_common_attributes();
  boost::log::add_console_log(
      std::cerr, boost::log::keywords::format =
          "[%TimeStamp%] <%Severity%>: %Message%");
  boost::log::core::get()->set_filter (
      boost::log::trivial::severity >= boost::log::trivial::info
  );
}


} //namespace social_network

#endif //SOCIAL_NETWORK_MICROSERVICES_LOGGER_H
