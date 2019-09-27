add_executable(
    PageService
    PageService.cpp
    ${THRIFT_GEN_CPP_DIR}/MovieReviewService.cpp
    ${THRIFT_GEN_CPP_DIR}/PageService.cpp
    ${THRIFT_GEN_CPP_DIR}/CastInfoService.cpp
    ${THRIFT_GEN_CPP_DIR}/PlotService.cpp
    ${THRIFT_GEN_CPP_DIR}/MovieInfoService.cpp
    ${THRIFT_GEN_CPP_DIR}/media_service_types.cpp

)

target_include_directories(
    PageService PRIVATE
    /usr/local/include/jaegertracing
)

target_link_libraries(
    PageService
    nlohmann_json::nlohmann_json
    ${THRIFT_LIB}
    ${CMAKE_THREAD_LIBS_INIT}
    ${Boost_LIBRARIES}
    Boost::log
    Boost::log_setup
    jaegertracing
)

install(TARGETS PageService DESTINATION ./)