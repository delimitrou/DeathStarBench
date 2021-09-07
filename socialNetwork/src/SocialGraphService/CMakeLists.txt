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
    /usr/local/include/hiredis
    /usr/local/include/sw
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
    Boost::program_options
    /usr/local/lib/libjaegertracing.so
    /usr/local/lib/libhiredis.a
    /usr/local/lib/libhiredis_ssl.a
    /usr/local/lib/libredis++.a
    OpenSSL::SSL
)

install(TARGETS SocialGraphService DESTINATION ./)