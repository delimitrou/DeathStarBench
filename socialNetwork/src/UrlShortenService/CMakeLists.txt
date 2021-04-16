add_executable(
    UrlShortenService
    UrlShortenService.cpp
    ${THRIFT_GEN_CPP_DIR}/UrlShortenService.cpp
    ${THRIFT_GEN_CPP_DIR}/social_network_types.cpp
)

target_include_directories(
    UrlShortenService PRIVATE
    ${LIBMEMCACHED_INCLUDE_DIR}
    ${MONGOC_INCLUDE_DIRS}
    /usr/local/include/jaegertracing
)

target_link_libraries(
    UrlShortenService
    nlohmann_json::nlohmann_json
    ${MONGOC_LIBRARIES}
    ${LIBMEMCACHED_LIBRARIES}
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    jaegertracing
)

install(TARGETS UrlShortenService DESTINATION ./)