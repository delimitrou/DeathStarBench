include("../cmake/Findlibmemcached.cmake")
include("../cmake/Findthrift.cmake")

find_package(libmongoc-1.0 1.13 REQUIRED)
find_package(nlohmann_json 3.5.0 REQUIRED)
find_package(Threads)
find_package(OpenSSL REQUIRED)

set(Boost_USE_STATIC_LIBS ON)
find_package(Boost 1.54.0 REQUIRED COMPONENTS log log_setup)
if(Boost_FOUND)
  include_directories(${Boost_INCLUDE_DIRS})
  link_directories(${Boost_LIBRARY_DIRS})
endif()

add_library(jaegertracing SHARED IMPORTED)
set_target_properties(jaegertracing PROPERTIES IMPORTED_LOCATION
    /usr/local/lib/libjaegertracing.so)

set(THRIFT_GEN_CPP_DIR ../../gen-cpp)

add_subdirectory(ComposeReviewService)
add_subdirectory(MovieIdService)
add_subdirectory(MovieReviewService)
add_subdirectory(RatingService)
add_subdirectory(ReviewStorageService)
add_subdirectory(TextService)
add_subdirectory(UniqueIdService)
add_subdirectory(UserReviewService)
add_subdirectory(UserService)
add_subdirectory(CastInfoService)
add_subdirectory(PlotService)
add_subdirectory(MovieInfoService)
add_subdirectory(PageService)