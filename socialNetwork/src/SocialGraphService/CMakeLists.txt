add_executable(
    SocialGraphService
    SocialGraphService.cpp
    ${THRIFT_GEN_CPP_DIR}/SocialGraphService.cpp
    ${THRIFT_GEN_CPP_DIR}/UserService.cpp
    ${THRIFT_GEN_CPP_DIR}/social_network_types.cpp
)

target_include_directories(
    SocialGraphService PRIVATE
    ${MONGOC_INCLUDE_DIRS}
    /usr/local/include/jaegertracing
    /usr/local/include/cpp_redis
)

target_link_libraries(
    SocialGraphService
    ${MONGOC_LIBRARIES}
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    nlohmann_json::nlohmann_json
    Boost::log
    Boost::log_setup
    OpenSSL::SSL
    /usr/local/lib/libjaegertracing.so
    /usr/local/lib/libcpp_redis.a
    /usr/local/lib/libtacopie.a
)

install(TARGETS SocialGraphService DESTINATION ./)