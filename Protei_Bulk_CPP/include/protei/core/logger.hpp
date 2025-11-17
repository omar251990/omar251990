/**
 * @file logger.hpp
 * @brief Logging System using spdlog
 */

#pragma once

#include <memory>
#include <string>
#include <spdlog/spdlog.h>
#include <spdlog/sinks/rotating_file_sink.h>
#include <spdlog/sinks/stdout_color_sinks.h>

namespace protei::core {

/**
 * @brief Logger wrapper around spdlog
 */
class Logger {
public:
    explicit Logger(const std::string& name = "protei_bulk");
    ~Logger() = default;

    // Logging methods
    template<typename... Args>
    void trace(const char* fmt, Args&&... args) {
        logger_->trace(fmt, std::forward<Args>(args)...);
    }

    template<typename... Args>
    void debug(const char* fmt, Args&&... args) {
        logger_->debug(fmt, std::forward<Args>(args)...);
    }

    template<typename... Args>
    void info(const char* fmt, Args&&... args) {
        logger_->info(fmt, std::forward<Args>(args)...);
    }

    template<typename... Args>
    void warn(const char* fmt, Args&&... args) {
        logger_->warn(fmt, std::forward<Args>(args)...);
    }

    template<typename... Args>
    void error(const char* fmt, Args&&... args) {
        logger_->error(fmt, std::forward<Args>(args)...);
    }

    template<typename... Args>
    void critical(const char* fmt, Args&&... args) {
        logger_->critical(fmt, std::forward<Args>(args)...);
    }

    // Set log level
    void set_level(spdlog::level::level_enum level);

    // Flush logs
    void flush();

private:
    std::shared_ptr<spdlog::logger> logger_;
};

} // namespace protei::core
