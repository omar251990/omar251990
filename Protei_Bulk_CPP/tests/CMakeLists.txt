# Google Test
enable_testing()

# Test sources
set(TEST_SOURCES
    test_main.cpp
    test_config.cpp
    test_database.cpp
    test_redis.cpp
)

# Create test executable
add_executable(protei_tests ${TEST_SOURCES})

target_link_libraries(protei_tests
    PRIVATE
    gtest
    gtest_main
    Threads::Threads
    nlohmann_json::nlohmann_json
    spdlog::spdlog
)

# Register tests
include(GoogleTest)
gtest_discover_tests(protei_tests)
