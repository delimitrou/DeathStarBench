cmake_minimum_required(VERSION 3.5)
project(media_microservices)

set(CMAKE_CXX_STANDARD 14)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
#set(CMAKE_CXX_FLAGS "${CMAKE_CXX_FLAGS} -W -Wall -Wextra -O2")
SET(CMAKE_INSTALL_PREFIX /usr/local/bin)

add_subdirectory(src)

#add_subdirectory(test)
#enable_testing()
#add_test(
#    testClientPool
#    testClientPool
#)
