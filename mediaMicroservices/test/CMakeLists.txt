include("../cmake/Findthrift.cmake")
include("../cmake/Findlibmemcached.cmake")

find_package(Threads)

set(Boost_USE_STATIC_LIBS ON)
find_package(Boost 1.54.0 REQUIRED COMPONENTS log log_setup)
include_directories(${Boost_INCLUDE_DIRS})
link_directories(${Boost_LIBRARY_DIRS})

#include_directories(
#    ${TEST_SOURCE_DIR}/src
#)
#
#link_libraries(
#    "${THRIFT_LIB}"
#    "${CMAKE_THREAD_LIBS_INIT}"
#    ${Boost_LIBRARIES}
#    Boost::log
#    Boost::log_setup
#)

#add_executable(
#    testClientPool
#    testClientPool.cpp
#    ../gen-cpp/ComposeReviewService.cpp
#    ../gen-cpp/media_service_types.cpp
#
#)

#add_executable(
#    testMemcachedAtomicIncrement
#    testMemcachedAtomicIncrement.cpp
#)
#
#target_include_directories(
#    testMemcachedAtomicIncrement PRIVATE
#    ${LIBMEMCACHED_INCLUDE_DIR}
#)
#
#target_link_libraries(
#    testMemcachedAtomicIncrement
#    ${LIBMEMCACHED_LIBRARIES}
#    ${CMAKE_THREAD_LIBS_INIT}
#)