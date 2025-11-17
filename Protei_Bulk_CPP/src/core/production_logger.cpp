/**
 * @file production_logger.cpp
 * @brief Production Logger Implementation
 */

#include "protei/core/production_logger.hpp"
#include <fstream>
#include <iomanip>
#include <sstream>
#include <sys/statvfs.h>
#include <unistd.h>
#include <thread>

namespace protei::core {

ProductionLogger::~ProductionLogger() {
    shutdown();
}

void ProductionLogger::initialize(const std::string& log_dir) {
    if (initialized_) return;

    create_loggers(log_dir);
    initialized_ = true;

    info("Production Logger initialized - Enterprise Edition");
    log_system_metrics(get_current_metrics());
}

void ProductionLogger::create_loggers(const std::string& log_dir) {
    try {
        // Application logger - rotating 50MB files, keep 10
        app_logger_ = spdlog::rotating_logger_mt(
            "application",
            log_dir + "/application.log",
            1024 * 1024 * 50,  // 50MB
            10                  // Keep 10 files
        );
        app_logger_->set_pattern("[%Y-%m-%d %H:%M:%S.%e] [%l] [%t] %v");
        app_logger_->set_level(spdlog::level::info);

        // Warning logger - rotating 10MB files, keep 5
        warning_logger_ = spdlog::rotating_logger_mt(
            "warning",
            log_dir + "/warning.log",
            1024 * 1024 * 10,  // 10MB
            5
        );
        warning_logger_->set_pattern("[%Y-%m-%d %H:%M:%S.%e] [WARNING] %v");
        warning_logger_->set_level(spdlog::level::warn);

        // Alarm logger - rotating 10MB files, keep 10 (important!)
        alarm_logger_ = spdlog::rotating_logger_mt(
            "alarm",
            log_dir + "/alarm.log",
            1024 * 1024 * 10,  // 10MB
            10
        );
        alarm_logger_->set_pattern("[%Y-%m-%d %H:%M:%S.%e] [ALARM] [CRITICAL] %v");
        alarm_logger_->set_level(spdlog::level::critical);

        // System logger - daily rotation
        system_logger_ = spdlog::daily_logger_mt(
            "system",
            log_dir + "/system.log",
            0,   // Rotate at midnight
            0
        );
        system_logger_->set_pattern("[%Y-%m-%d %H:%M:%S.%e] [SYSTEM] %v");

        // CDR logger - daily rotation, keep 90 days
        cdr_logger_ = spdlog::daily_logger_mt(
            "cdr",
            log_dir + "/cdr.log",
            0,
            0
        );
        cdr_logger_->set_pattern("%v");  // Raw CDR data

        // Security logger - rotating 20MB, keep 20
        security_logger_ = spdlog::rotating_logger_mt(
            "security",
            log_dir + "/security.log",
            1024 * 1024 * 20,
            20
        );
        security_logger_->set_pattern("[%Y-%m-%d %H:%M:%S.%e] [SECURITY] %v");

        // All loggers flush on error
        app_logger_->flush_on(spdlog::level::err);
        warning_logger_->flush_on(spdlog::level::warn);
        alarm_logger_->flush_on(spdlog::level::critical);
        system_logger_->flush_on(spdlog::level::info);
        security_logger_->flush_on(spdlog::level::warn);

    } catch (const spdlog::spdlog_ex& ex) {
        std::cerr << "Logger initialization failed: " << ex.what() << std::endl;
    }
}

void ProductionLogger::shutdown() {
    if (!initialized_) return;

    info("Production Logger shutting down");

    spdlog::shutdown();
    initialized_ = false;
}

SystemMetrics ProductionLogger::get_current_metrics() {
    std::lock_guard<std::mutex> lock(metrics_mutex_);

    SystemMetrics metrics;
    metrics.timestamp = std::chrono::system_clock::now();

    // Get CPU usage
    metrics.cpu_usage_percent = SystemMonitor::instance().get_cpu_usage();

    // Get memory usage
    metrics.memory_usage_mb = SystemMonitor::instance().get_memory_usage_mb();
    metrics.memory_usage_percent = SystemMonitor::instance().get_memory_usage_percent();

    // Get disk usage
    struct statvfs stat;
    if (statvfs("/opt/protei_bulk", &stat) == 0) {
        unsigned long total = stat.f_blocks * stat.f_frsize;
        unsigned long available = stat.f_bavail * stat.f_frsize;
        metrics.disk_usage_mb = (total - available) / (1024 * 1024);
        metrics.disk_available_mb = available / (1024 * 1024);
    }

    // Get application metrics
    metrics.active_connections = SystemMonitor::instance().get_active_connections();
    metrics.queue_depth = SystemMonitor::instance().get_queue_depth();
    metrics.messages_per_second = SystemMonitor::instance().get_messages_per_second();

    current_metrics_ = metrics;
    return metrics;
}

void ProductionLogger::log_system_metrics(const SystemMetrics& metrics) {
    if (!system_logger_) return;

    std::ostringstream oss;
    oss << std::fixed << std::setprecision(2);
    oss << "CPU:" << metrics.cpu_usage_percent << "% ";
    oss << "| Memory:" << metrics.memory_usage_mb << "MB ("
        << metrics.memory_usage_percent << "%) ";
    oss << "| Disk:" << metrics.disk_usage_mb << "MB used, "
        << metrics.disk_available_mb << "MB available ";
    oss << "| Connections:" << metrics.active_connections << " ";
    oss << "| Queue:" << metrics.queue_depth << " ";
    oss << "| TPS:" << metrics.messages_per_second;

    system_logger_->info(oss.str());
}

std::string ProductionLogger::format_cdr_record(const CDRRecord& cdr) {
    std::ostringstream oss;

    // CSV format for easy parsing
    oss << cdr.message_id << ","
        << cdr.campaign_id << ","
        << cdr.customer_id << ","
        << cdr.msisdn << ","
        << cdr.sender_id << ","
        << "\"" << cdr.message_text << "\","
        << cdr.message_length << ","
        << cdr.message_parts << ","
        << cdr.submit_time << ","
        << cdr.delivery_time << ","
        << cdr.status << ","
        << cdr.error_code << ","
        << cdr.smsc_id << ","
        << cdr.route_id << ","
        << std::fixed << std::setprecision(4) << cdr.cost << ","
        << cdr.operator_name << ","
        << cdr.country_code << ","
        << cdr.retry_count << ","
        << cdr.final_status << ","
        << cdr.processing_time_ms;

    return oss.str();
}

void ProductionLogger::log_cdr(const CDRRecord& cdr) {
    if (!cdr_logger_) return;

    cdr_logger_->info(format_cdr_record(cdr));

    // Also log to CDR manager for statistics
    CDRManager::instance().record_message(cdr);
}

void ProductionLogger::log_campaign_stats(const std::string& campaign_id,
                                         long total_sent,
                                         long successful,
                                         long failed,
                                         double success_rate) {
    info("Campaign {} Statistics: Total={}, Success={}, Failed={}, SuccessRate={:.2f}%",
         campaign_id, total_sent, successful, failed, success_rate);
}

void ProductionLogger::log_performance(const std::string& operation,
                                      long duration_ms,
                                      bool success) {
    if (duration_ms > 1000) {
        warning("Slow operation: {} took {}ms", operation, duration_ms);
    } else {
        debug("Operation: {} completed in {}ms ({})",
              operation, duration_ms, success ? "success" : "failed");
    }
}

void ProductionLogger::log_security_event(const std::string& event_type,
                                         const std::string& user,
                                         const std::string& ip_address,
                                         const std::string& details) {
    if (!security_logger_) return;

    std::ostringstream oss;
    oss << event_type << " | User:" << user
        << " | IP:" << ip_address
        << " | Details:" << details;

    security_logger_->warn(oss.str());

    // Log critical security events as alarms
    if (event_type == "UNAUTHORIZED_ACCESS" ||
        event_type == "BRUTE_FORCE" ||
        event_type == "INJECTION_ATTEMPT") {
        alarm("Security Alert: {}", oss.str());
    }
}

// System Monitor Implementation
void SystemMonitor::start() {
    if (running_) return;

    running_ = true;
    monitor_thread_ = std::make_unique<std::thread>(&SystemMonitor::monitoring_loop, this);

    ProductionLogger::instance().info("System Monitor started");
}

void SystemMonitor::stop() {
    if (!running_) return;

    running_ = false;
    if (monitor_thread_ && monitor_thread_->joinable()) {
        monitor_thread_->join();
    }

    ProductionLogger::instance().info("System Monitor stopped");
}

void SystemMonitor::monitoring_loop() {
    while (running_) {
        // Collect and log metrics every 60 seconds
        auto metrics = ProductionLogger::instance().get_current_metrics();
        ProductionLogger::instance().log_system_metrics(metrics);

        // Check thresholds and create alarms
        if (metrics.cpu_usage_percent > 90.0) {
            ProductionLogger::instance().alarm(
                "High CPU usage: {:.2f}% (threshold: 90%)",
                metrics.cpu_usage_percent);
        }

        if (metrics.memory_usage_percent > 85.0) {
            ProductionLogger::instance().alarm(
                "High memory usage: {:.2f}% (threshold: 85%)",
                metrics.memory_usage_percent);
        }

        if (metrics.disk_available_mb < 1024) {  // Less than 1GB
            ProductionLogger::instance().alarm(
                "Low disk space: {}MB available (threshold: 1024MB)",
                metrics.disk_available_mb);
        }

        if (metrics.queue_depth > 10000) {
            ProductionLogger::instance().warning(
                "High queue depth: {} messages (threshold: 10000)",
                metrics.queue_depth);
        }

        std::this_thread::sleep_for(std::chrono::seconds(60));
    }
}

double SystemMonitor::get_cpu_usage() {
    // Read /proc/stat for CPU usage
    std::ifstream file("/proc/stat");
    std::string line;
    std::getline(file, line);

    // Simple CPU usage calculation
    // In production, use more sophisticated method
    static long long prev_idle = 0, prev_total = 0;

    long long user, nice, system, idle, iowait, irq, softirq;
    sscanf(line.c_str(), "cpu %lld %lld %lld %lld %lld %lld %lld",
           &user, &nice, &system, &idle, &iowait, &irq, &softirq);

    long long total = user + nice + system + idle + iowait + irq + softirq;
    long long total_diff = total - prev_total;
    long long idle_diff = idle - prev_idle;

    double cpu_percent = 0.0;
    if (total_diff > 0) {
        cpu_percent = 100.0 * (1.0 - (double)idle_diff / total_diff);
    }

    prev_idle = idle;
    prev_total = total;

    return cpu_percent;
}

double SystemMonitor::get_memory_usage_mb() {
    // Read /proc/meminfo
    std::ifstream file("/proc/meminfo");
    std::string line;
    long mem_total = 0, mem_free = 0, mem_available = 0;

    while (std::getline(file, line)) {
        if (line.find("MemTotal:") == 0) {
            sscanf(line.c_str(), "MemTotal: %ld kB", &mem_total);
        } else if (line.find("MemFree:") == 0) {
            sscanf(line.c_str(), "MemFree: %ld kB", &mem_free);
        } else if (line.find("MemAvailable:") == 0) {
            sscanf(line.c_str(), "MemAvailable: %ld kB", &mem_available);
        }
    }

    long mem_used = mem_total - mem_available;
    return mem_used / 1024.0;  // Convert to MB
}

double SystemMonitor::get_memory_usage_percent() {
    std::ifstream file("/proc/meminfo");
    std::string line;
    long mem_total = 0, mem_available = 0;

    while (std::getline(file, line)) {
        if (line.find("MemTotal:") == 0) {
            sscanf(line.c_str(), "MemTotal: %ld kB", &mem_total);
        } else if (line.find("MemAvailable:") == 0) {
            sscanf(line.c_str(), "MemAvailable: %ld kB", &mem_available);
        }
    }

    if (mem_total > 0) {
        long mem_used = mem_total - mem_available;
        return 100.0 * mem_used / mem_total;
    }

    return 0.0;
}

long SystemMonitor::get_disk_usage() {
    struct statvfs stat;
    if (statvfs("/opt/protei_bulk", &stat) == 0) {
        unsigned long total = stat.f_blocks * stat.f_frsize;
        unsigned long available = stat.f_bavail * stat.f_frsize;
        return (total - available) / (1024 * 1024);  // MB
    }
    return 0;
}

int SystemMonitor::get_active_connections() {
    // This would be implemented based on actual connection tracking
    // Placeholder for now
    return 0;
}

int SystemMonitor::get_queue_depth() {
    // This would be implemented based on actual queue monitoring
    // Placeholder for now
    return 0;
}

long SystemMonitor::get_messages_per_second() {
    // This would be implemented based on actual message throughput tracking
    // Placeholder for now
    return 0;
}

// CDR Manager Implementation
void CDRManager::initialize(const std::string& cdr_dir) {
    cdr_directory_ = cdr_dir;
    ProductionLogger::instance().info("CDR Manager initialized: {}", cdr_dir);
}

void CDRManager::record_message(const CDRRecord& cdr) {
    std::lock_guard<std::mutex> lock(cdr_mutex_);
    pending_cdrs_[cdr.message_id] = cdr;
}

void CDRManager::update_delivery_status(const std::string& message_id,
                                        const std::string& status,
                                        const std::string& delivery_time) {
    std::lock_guard<std::mutex> lock(cdr_mutex_);

    auto it = pending_cdrs_.find(message_id);
    if (it != pending_cdrs_.end()) {
        it->second.final_status = status;
        it->second.delivery_time = delivery_time;

        // Log updated CDR
        ProductionLogger::instance().log_cdr(it->second);

        // Remove from pending
        pending_cdrs_.erase(it);
    }
}

CDRManager::Statistics CDRManager::get_statistics(const std::string& campaign_id) {
    std::lock_guard<std::mutex> lock(cdr_mutex_);

    Statistics stats{};
    // Implementation would query database or aggregate from CDR files
    // Placeholder for now
    return stats;
}

CDRManager::Statistics CDRManager::get_daily_statistics() {
    std::lock_guard<std::mutex> lock(cdr_mutex_);

    Statistics stats{};
    // Implementation would query database for today's statistics
    // Placeholder for now
    return stats;
}

} // namespace protei::core
