add_executable(
    MediaService
    MediaService.cpp
    ${THRIFT_GEN_CPP_DIR}/MediaService.cpp
    ${THRIFT_GEN_CPP_DIR}/social_network_types.cpp
)

target_include_directories(
    MediaService PRIVATE
    /usr/local/include/jaegertracing
)

target_link_libraries(
    MediaService
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    jaegertracing
)

install(TARGETS MediaService DESTINATION ./)