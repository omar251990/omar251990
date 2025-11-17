cmake_minimum_required(VERSION 3.20)
project(Protei_Bulk_CPP VERSION 1.0.0 LANGUAGES CXX)

# C++ Standard
set(CMAKE_CXX_STANDARD 20)
set(CMAKE_CXX_STANDARD_REQUIRED ON)
set(CMAKE_CXX_EXTENSIONS OFF)

# Build type
if(NOT CMAKE_BUILD_TYPE)
    set(CMAKE_BUILD_TYPE Release)
endif()

# Compiler flags
set(CMAKE_CXX_FLAGS "-Wall -Wextra -Wpedantic")
set(CMAKE_CXX_FLAGS_DEBUG "-g -O0 -DDEBUG")
set(CMAKE_CXX_FLAGS_RELEASE "-O3 -DNDEBUG -march=native")

# Output directories
set(CMAKE_RUNTIME_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/bin)
set(CMAKE_LIBRARY_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)
set(CMAKE_ARCHIVE_OUTPUT_DIRECTORY ${CMAKE_BINARY_DIR}/lib)

# Include directories
include_directories(${PROJECT_SOURCE_DIR}/include)
include_directories(${PROJECT_SOURCE_DIR}/libs)

# Find packages
find_package(Threads REQUIRED)
find_package(OpenSSL REQUIRED)
find_package(ZLIB REQUIRED)

# Third-party libraries
include(FetchContent)

# Boost - for ASIO, date_time, etc.
find_package(Boost 1.75 REQUIRED COMPONENTS system thread date_time program_options)
include_directories(${Boost_INCLUDE_DIRS})

# nlohmann/json - JSON parsing
FetchContent_Declare(
    json
    GIT_REPOSITORY https://github.com/nlohmann/json.git
    GIT_TAG v3.11.3
)
FetchContent_MakeAvailable(json)

# spdlog - Logging
FetchContent_Declare(
    spdlog
    GIT_REPOSITORY https://github.com/gabime/spdlog.git
    GIT_TAG v1.12.0
)
FetchContent_MakeAvailable(spdlog)

# cpp-httplib - HTTP server
FetchContent_Declare(
    httplib
    GIT_REPOSITORY https://github.com/yhirose/cpp-httplib.git
    GIT_TAG v0.14.3
)
FetchContent_MakeAvailable(httplib)

# libpqxx - PostgreSQL C++ client
find_package(libpqxx REQUIRED)
if(NOT libpqxx_FOUND)
    message(STATUS "libpqxx not found via find_package, trying pkg-config")
    find_package(PkgConfig REQUIRED)
    pkg_check_modules(LIBPQXX REQUIRED libpqxx)
    include_directories(${LIBPQXX_INCLUDE_DIRS})
    link_directories(${LIBPQXX_LIBRARY_DIRS})
endif()

# redis-plus-plus - Redis C++ client
find_package(PkgConfig REQUIRED)
pkg_check_modules(REDIS++ REQUIRED redis++)
include_directories(${REDIS++_INCLUDE_DIRS})
link_directories(${REDIS++_LIBRARY_DIRS})

# Google Test - for testing
FetchContent_Declare(
    googletest
    GIT_REPOSITORY https://github.com/google/googletest.git
    GIT_TAG v1.14.0
)
set(gtest_force_shared_crt ON CACHE BOOL "" FORCE)
FetchContent_MakeAvailable(googletest)

# Source files
set(PROTEI_SOURCES
    # Core
    src/core/config.cpp
    src/core/database.cpp
    src/core/redis_client.cpp
    src/core/logger.cpp

    # API
    src/api/http_server.cpp
    src/api/api_router.cpp
    src/api/middleware.cpp

    # Services
    src/services/routing_service.cpp
    src/services/campaign_service.cpp
    src/services/profiling_service.cpp
    src/services/segmentation_service.cpp
    src/services/dcdl_service.cpp
    src/services/message_service.cpp
    src/services/analytics_service.cpp

    # Protocols
    src/protocols/smpp_server.cpp
    src/protocols/smpp_client.cpp
    src/protocols/smpp_pdu.cpp
    src/protocols/whatsapp_client.cpp
    src/protocols/email_client.cpp

    # Models
    src/models/message.cpp
    src/models/campaign.cpp
    src/models/user.cpp
    src/models/route.cpp

    # Utils
    src/utils/crypto.cpp
    src/utils/validators.cpp
    src/utils/string_utils.cpp
    src/utils/date_utils.cpp
)

# Main executable
add_executable(protei_bulk src/main.cpp ${PROTEI_SOURCES})

# Link libraries
target_link_libraries(protei_bulk
    PRIVATE
    Threads::Threads
    OpenSSL::SSL
    OpenSSL::Crypto
    ZLIB::ZLIB
    ${Boost_LIBRARIES}
    nlohmann_json::nlohmann_json
    spdlog::spdlog
    httplib::httplib
    pqxx
    redis++
    hiredis
)

# Install targets
install(TARGETS protei_bulk
    RUNTIME DESTINATION bin
    LIBRARY DESTINATION lib
    ARCHIVE DESTINATION lib
)

# Tests
enable_testing()
add_subdirectory(tests)

# Documentation
find_package(Doxygen)
if(DOXYGEN_FOUND)
    set(DOXYGEN_IN ${CMAKE_CURRENT_SOURCE_DIR}/docs/Doxyfile.in)
    set(DOXYGEN_OUT ${CMAKE_CURRENT_BINARY_DIR}/Doxyfile)

    configure_file(${DOXYGEN_IN} ${DOXYGEN_OUT} @ONLY)

    add_custom_target(docs
        COMMAND ${DOXYGEN_EXECUTABLE} ${DOXYGEN_OUT}
        WORKING_DIRECTORY ${CMAKE_CURRENT_BINARY_DIR}
        COMMENT "Generating API documentation with Doxygen"
        VERBATIM
    )
endif()

# Print build configuration
message(STATUS "")
message(STATUS "Protei_Bulk C++ Build Configuration:")
message(STATUS "  Version: ${PROJECT_VERSION}")
message(STATUS "  Build Type: ${CMAKE_BUILD_TYPE}")
message(STATUS "  C++ Standard: C++${CMAKE_CXX_STANDARD}")
message(STATUS "  Compiler: ${CMAKE_CXX_COMPILER_ID} ${CMAKE_CXX_COMPILER_VERSION}")
message(STATUS "  Install Prefix: ${CMAKE_INSTALL_PREFIX}")
message(STATUS "")
