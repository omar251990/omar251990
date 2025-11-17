/**
 * @file logger.cpp
 * @brief Logger Implementation
 */

#include "protei/core/logger.hpp"
#include <vector>

namespace protei::core {

Logger::Logger(const std::string& name) {
    try {
        // Create sinks
        auto console_sink = std::make_shared<spdlog::sinks::stdout_color_sink_mt>();
        console_sink->set_level(spdlog::level::info);
        console_sink->set_pattern("[%Y-%m-%d %H:%M:%S.%e] [%^%l%$] [%n] %v");

        auto file_sink = std::make_shared<spdlog::sinks::rotating_file_sink_mt>(
            "logs/protei_bulk.log",
            1024 * 1024 * 10,  // 10 MB
            5                   // 5 rotating files
        );
        file_sink->set_level(spdlog::level::trace);
        file_sink->set_pattern("[%Y-%m-%d %H:%M:%S.%e] [%l] [%n] [%t] %v");

        // Create logger with multiple sinks
        std::vector<spdlog::sink_ptr> sinks{console_sink, file_sink};
        logger_ = std::make_shared<spdlog::logger>(name, sinks.begin(), sinks.end());
        logger_->set_level(spdlog::level::trace);
        logger_->flush_on(spdlog::level::err);

        // Register it
        spdlog::register_logger(logger_);

    } catch (const spdlog::spdlog_ex& ex) {
        std::cerr << "Log initialization failed: " << ex.what() << std::endl;
    }
}

void Logger::set_level(spdlog::level::level_enum level) {
    logger_->set_level(level);
}

void Logger::flush() {
    logger_->flush();
}

} // namespace protei::core
