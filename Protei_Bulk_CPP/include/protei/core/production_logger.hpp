/**
 * @file production_logger.hpp
 * @brief Enterprise Production Logging System
 *
 * Comprehensive logging for commercial deployment:
 * - warning.log: All warnings and non-critical issues
 * - alarm.log: Critical errors and system alarms
 * - system.log: Performance metrics and utilization
 * - application.log: General application logs
 * - cdr.log: Complete call detail records
 */

#pragma once

#include <memory>
#include <string>
#include <mutex>
#include <fstream>
#include <chrono>
#include <spdlog/spdlog.h>
#include <spdlog/sinks/rotating_file_sink.h>
#include <spdlog/sinks/daily_file_sink.h>

namespace protei::core {

/**
 * @brief System metrics structure
 */
struct SystemMetrics {
    double cpu_usage_percent;
    double memory_usage_mb;
    double memory_usage_percent;
    long disk_usage_mb;
    long disk_available_mb;
    int active_connections;
    int queue_depth;
    long messages_per_second;
    long total_messages_sent;
    long total_messages_failed;
    std::chrono::system_clock::time_point timestamp;
};

/**
 * @brief CDR (Call Detail Record) structure
 */
struct CDRRecord {
    std::string message_id;
    std::string campaign_id;
    std::string customer_id;
    std::string msisdn;
    std::string sender_id;
    std::string message_text;
    int message_length;
    int message_parts;
    std::string submit_time;
    std::string delivery_time;
    std::string status;
    std::string error_code;
    std::string smsc_id;
    std::string route_id;
    double cost;
    std::string operator_name;
    std::string country_code;
    int retry_count;
    std::string final_status;
    long processing_time_ms;
};

/**
 * @brief Production Logger - Enterprise Grade
 */
class ProductionLogger {
public:
    static ProductionLogger& instance() {
        static ProductionLogger instance;
        return instance;
    }

    ProductionLogger(const ProductionLogger&) = delete;
    ProductionLogger& operator=(const ProductionLogger&) = delete;

    void initialize(const std::string& log_dir = "/opt/protei_bulk/logs");
    void shutdown();

    // Application logging
    template<typename... Args>
    void info(const char* fmt, Args&&... args) {
        if (app_logger_) {
            app_logger_->info(fmt, std::forward<Args>(args)...);
        }
    }

    template<typename... Args>
    void debug(const char* fmt, Args&&... args) {
        if (app_logger_) {
            app_logger_->debug(fmt, std::forward<Args>(args)...);
        }
    }

    // Warning logging
    template<typename... Args>
    void warning(const char* fmt, Args&&... args) {
        if (warning_logger_) {
            warning_logger_->warn(fmt, std::forward<Args>(args)...);
        }
    }

    // Alarm logging (critical errors)
    template<typename... Args>
    void alarm(const char* fmt, Args&&... args) {
        if (alarm_logger_) {
            alarm_logger_->critical(fmt, std::forward<Args>(args)...);
        }
        // Also log to application log
        if (app_logger_) {
            app_logger_->critical(fmt, std::forward<Args>(args)...);
        }
    }

    // System metrics logging
    void log_system_metrics(const SystemMetrics& metrics);

    // CDR logging
    void log_cdr(const CDRRecord& cdr);

    // Campaign statistics
    void log_campaign_stats(const std::string& campaign_id,
                           long total_sent,
                           long successful,
                           long failed,
                           double success_rate);

    // Performance logging
    void log_performance(const std::string& operation,
                        long duration_ms,
                        bool success);

    // Security logging
    void log_security_event(const std::string& event_type,
                           const std::string& user,
                           const std::string& ip_address,
                           const std::string& details);

    // Get current metrics
    SystemMetrics get_current_metrics();

private:
    ProductionLogger() = default;
    ~ProductionLogger();

    void create_loggers(const std::string& log_dir);
    void collect_system_metrics();
    std::string format_cdr_record(const CDRRecord& cdr);

    std::shared_ptr<spdlog::logger> app_logger_;
    std::shared_ptr<spdlog::logger> warning_logger_;
    std::shared_ptr<spdlog::logger> alarm_logger_;
    std::shared_ptr<spdlog::logger> system_logger_;
    std::shared_ptr<spdlog::logger> cdr_logger_;
    std::shared_ptr<spdlog::logger> security_logger_;

    std::mutex metrics_mutex_;
    SystemMetrics current_metrics_;
    bool initialized_ = false;
};

/**
 * @brief System Monitor - Continuous monitoring
 */
class SystemMonitor {
public:
    static SystemMonitor& instance() {
        static SystemMonitor instance;
        return instance;
    }

    void start();
    void stop();

    // Get system statistics
    double get_cpu_usage();
    double get_memory_usage_mb();
    double get_memory_usage_percent();
    long get_disk_usage();

    // Get application statistics
    int get_active_connections();
    int get_queue_depth();
    long get_messages_per_second();

private:
    SystemMonitor() = default;
    void monitoring_loop();

    std::atomic<bool> running_{false};
    std::unique_ptr<std::thread> monitor_thread_;
};

/**
 * @brief CDR Manager - Complete tracking
 */
class CDRManager {
public:
    static CDRManager& instance() {
        static CDRManager instance;
        return instance;
    }

    void initialize(const std::string& cdr_dir);

    // Record message CDR
    void record_message(const CDRRecord& cdr);

    // Update CDR with delivery status
    void update_delivery_status(const std::string& message_id,
                                const std::string& status,
                                const std::string& delivery_time);

    // Get statistics
    struct Statistics {
        long total_messages;
        long successful;
        long failed;
        long pending;
        double success_rate;
        double average_delivery_time_ms;
    };

    Statistics get_statistics(const std::string& campaign_id);
    Statistics get_daily_statistics();

private:
    CDRManager() = default;

    std::string cdr_directory_;
    std::mutex cdr_mutex_;
    std::unordered_map<std::string, CDRRecord> pending_cdrs_;
};

} // namespace protei::core
