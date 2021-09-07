add_executable(
    HomeTimelineService
    HomeTimelineService.cpp
    ${THRIFT_GEN_CPP_DIR}/HomeTimelineService.cpp
    ${THRIFT_GEN_CPP_DIR}/PostStorageService.cpp
    ${THRIFT_GEN_CPP_DIR}/SocialGraphService.cpp
    ${THRIFT_GEN_CPP_DIR}/social_network_types.cpp
)

target_include_directories(
    HomeTimelineService PRIVATE
    /usr/local/include/jaegertracing
    /usr/local/include/hiredis
    /usr/local/include/sw
)

target_link_libraries(
    HomeTimelineService
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    Boost::program_options
    jaegertracing
    /usr/local/lib/libhiredis.a
    /usr/local/lib/libhiredis_ssl.a
    /usr/local/lib/libredis++.a
    OpenSSL::SSL
)

install(TARGETS HomeTimelineService DESTINATION ./)