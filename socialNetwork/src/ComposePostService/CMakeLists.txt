add_executable(
    ComposePostService
    ComposePostService.cpp
    ${THRIFT_GEN_CPP_DIR}/ComposePostService.cpp
    ${THRIFT_GEN_CPP_DIR}/PostStorageService.cpp
    ${THRIFT_GEN_CPP_DIR}/UserTimelineService.cpp
    ${THRIFT_GEN_CPP_DIR}/social_network_types.cpp
)

target_include_directories(
    ComposePostService PRIVATE
    /usr/local/include/jaegertracing
    /usr/local/include/cpp_redis
)

target_link_libraries(
    ComposePostService
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
    /usr/local/lib/libSimpleAmqpClient.so

)

install(TARGETS ComposePostService DESTINATION ./)