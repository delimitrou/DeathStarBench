cmake_minimum_required(VERSION 3.1)

project(lua-bridge-tracer)


find_package(OpenTracing 1.5.1)

set(CMAKE_CXX_STANDARD 11)

if("${CMAKE_CXX_COMPILER_ID}" MATCHES "Clang")
  set(CMAKE_SHARED_LIBRARY_CREATE_CXX_FLAGS "${CMAKE_SHARED_LIBRARY_CREATE_CXX_FLAGS} -Wl,-undefined,dynamic_lookup")
elseif ("${CMAKE_CXX_COMPILER_ID}" STREQUAL "GNU")
  set(CMAKE_SHARED_LIBRARY_CREATE_CXX_FLAGS "${CMAKE_SHARED_LIBRARY_CREATE_CXX_FLAGS} -Wl,--unresolved-symbols=ignore-in-object-files")
endif()

include_directories(SYSTEM /usr/local/openresty/luajit/include/luajit-2.1)

add_library(opentracing_bridge_tracer SHARED src/module.cpp
                                             src/utility.cpp
                                             src/dynamic_tracer.cpp
                                             src/lua_tracer.cpp
                                             src/carrier.cpp
                                             src/lua_span_context.cpp
                                             src/lua_span.cpp)

target_link_libraries(opentracing_bridge_tracer OpenTracing::opentracing)
set_target_properties(opentracing_bridge_tracer PROPERTIES PREFIX "")
set_target_properties(opentracing_bridge_tracer PROPERTIES SUFFIX ".so")

set(LUA_MODULE_DIR /usr/local/openresty/lualib/opentracing_bridge_tracer)

install(TARGETS opentracing_bridge_tracer LIBRARY DESTINATION ${LUA_MODULE_DIR})
