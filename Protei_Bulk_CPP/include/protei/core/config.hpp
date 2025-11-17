/**
 * @file config.hpp
 * @brief Configuration Management System
 */

#pragma once

#include <string>
#include <map>
#include <memory>
#include <stdexcept>
#include <boost/property_tree/ptree.hpp>
#include <boost/property_tree/ini_parser.hpp>

namespace protei::core {

/**
 * @brief Database configuration
 */
struct DatabaseConfig {
    std::string host = "localhost";
    int port = 5432;
    std::string database = "protei_bulk";
    std::string username = "protei";
    std::string password = "elephant";
    int pool_size = 20;
    int max_connections = 50;
    int timeout_seconds = 30;

    std::string get_connection_string() const {
        return "host=" + host +
               " port=" + std::to_string(port) +
               " dbname=" + database +
               " user=" + username +
               " password=" + password +
               " connect_timeout=" + std::to_string(timeout_seconds);
    }
};

/**
 * @brief Redis configuration
 */
struct RedisConfig {
    bool enabled = true;
    std::string host = "localhost";
    int port = 6379;
    std::string password = "";
    int database = 0;
    int pool_size = 10;
    int timeout_ms = 1000;
};

/**
 * @brief SMPP configuration
 */
struct SmppConfig {
    bool enabled = true;
    std::string bind_address = "0.0.0.0";
    int bind_port = 2775;
    std::string system_id = "PROTEI_BULK";
    int max_connections = 100;
    int enquire_link_interval = 30;
    int window_size = 10;
};

/**
 * @brief API configuration
 */
struct ApiConfig {
    bool enabled = true;
    std::string bind_address = "0.0.0.0";
    int bind_port = 8080;
    bool enable_https = false;
    std::string ssl_cert_file;
    std::string ssl_key_file;
    bool enable_cors = true;
    int max_body_size_mb = 100;
    int thread_pool_size = 8;
};

/**
 * @brief Application configuration
 */
struct AppConfig {
    std::string app_name = "Protei_Bulk";
    std::string version = "1.0.0";
    std::string build = "001";
    std::string environment = "production";
    std::string base_dir = "/opt/protei_bulk";
    int max_workers = 10;
    int queue_size = 10000;
    bool enable_monitoring = true;
};

/**
 * @brief Security configuration
 */
struct SecurityConfig {
    std::string secret_key;
    std::string jwt_algorithm = "HS256";
    int access_token_expire_minutes = 60;
    int refresh_token_expire_days = 7;
    int password_min_length = 12;
    int password_expiry_days = 90;
    int max_failed_attempts = 5;
    int lockout_duration_minutes = 30;
    bool enable_2fa = true;
};

/**
 * @brief Main configuration class (Singleton)
 */
class Config {
public:
    /**
     * @brief Get singleton instance
     */
    static Config& instance() {
        static Config instance;
        return instance;
    }

    // Delete copy/move constructors
    Config(const Config&) = delete;
    Config& operator=(const Config&) = delete;
    Config(Config&&) = delete;
    Config& operator=(Config&&) = delete;

    /**
     * @brief Load configuration from file
     */
    void load_from_file(const std::string& config_file);

    /**
     * @brief Load configuration from environment variables
     */
    void load_from_env();

    /**
     * @brief Get configuration value by key
     */
    template<typename T>
    T get(const std::string& key, const T& default_value) const;

    // Getters
    const AppConfig& get_app_config() const { return app_config_; }
    const DatabaseConfig& get_database_config() const { return db_config_; }
    const RedisConfig& get_redis_config() const { return redis_config_; }
    const SmppConfig& get_smpp_config() const { return smpp_config_; }
    const ApiConfig& get_api_config() const { return api_config_; }
    const SecurityConfig& get_security_config() const { return security_config_; }

    // Convenience methods
    std::string get_app_environment() const { return app_config_.environment; }
    bool is_http_enabled() const { return api_config_.enabled; }
    bool is_smpp_enabled() const { return smpp_config_.enabled; }
    std::string get_api_bind_address() const { return api_config_.bind_address; }
    int get_api_bind_port() const { return api_config_.bind_port; }
    std::string get_smpp_bind_address() const { return smpp_config_.bind_address; }
    int get_smpp_bind_port() const { return smpp_config_.bind_port; }

private:
    Config();
    ~Config() = default;

    void load_app_config(const boost::property_tree::ptree& pt);
    void load_db_config(const boost::property_tree::ptree& pt);
    void load_protocol_config(const boost::property_tree::ptree& pt);
    void load_security_config(const boost::property_tree::ptree& pt);
    void generate_secret_key();

    AppConfig app_config_;
    DatabaseConfig db_config_;
    RedisConfig redis_config_;
    SmppConfig smpp_config_;
    ApiConfig api_config_;
    SecurityConfig security_config_;
};

} // namespace protei::core
