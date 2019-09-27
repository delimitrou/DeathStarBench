add_executable(
    MovieInfoService
    MovieInfoService.cpp
    ${THRIFT_GEN_CPP_DIR}/MovieInfoService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
)

target_include_directories(
    MovieInfoService PRIVATE
    ${LIBMEMCACHED_INCLUDE_DIR}
    ${MONGOC_INCLUDE_DIRS}
    /usr/local/include/jaegertracing
)

target_link_libraries(
    MovieInfoService
    ${MONGOC_LIBRARIES}
    ${LIBMEMCACHED_LIBRARIES}
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    jaegertracing
)

target_compile_definitions (
    MovieInfoService PRIVATE
    "${MONGOC_DEFINITIONS}"
)

install(TARGETS MovieInfoService DESTINATION ./)