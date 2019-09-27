add_executable(
    ComposeReviewService
    ComposeReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/ComposeReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp
    ${THRIFT_GEN_CPP_DIR}/ReviewStorageService.cpp
    ${THRIFT_GEN_CPP_DIR}/UserReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/MovieReviewService.cpp
)

target_include_directories(
    ComposeReviewService PRIVATE
    ${LIBMEMCACHED_INCLUDE_DIR}
    /usr/local/include/jaegertracing
)

target_link_libraries(
    ComposeReviewService
    ${LIBMEMCACHED_LIBRARIES}
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    jaegertracing
)

install(TARGETS ComposeReviewService DESTINATION ./)