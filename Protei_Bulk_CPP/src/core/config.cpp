/**
 * @file config.cpp
 * @brief Configuration Management Implementation
 */

#include "protei/core/config.hpp"
#include <fstream>
#include <cstdlib>
#include <random>
#include <sstream>
#include <iomanip>

namespace protei::core {

Config::Config() {
    // Load from environment variables first
    load_from_env();

    // Generate secret key if not set
    if (security_config_.secret_key.empty()) {
        generate_secret_key();
    }
}

void Config::load_from_file(const std::string& config_file) {
    try {
        boost::property_tree::ptree pt;
        boost::property_tree::ini_parser::read_ini(config_file, pt);

        load_app_config(pt);
        load_db_config(pt);
        load_protocol_config(pt);
        load_security_config(pt);

    } catch (const std::exception& e) {
        throw std::runtime_error("Failed to load config file: " + std::string(e.what()));
    }
}

void Config::load_from_env() {
    // Database configuration
    if (const char* db_host = std::getenv("DB_HOST")) {
        db_config_.host = db_host;
    }
    if (const char* db_port = std::getenv("DB_PORT")) {
        db_config_.port = std::atoi(db_port);
    }
    if (const char* db_name = std::getenv("DB_NAME")) {
        db_config_.database = db_name;
    }
    if (const char* db_user = std::getenv("DB_USER")) {
        db_config_.username = db_user;
    }
    if (const char* db_password = std::getenv("DB_PASSWORD")) {
        db_config_.password = db_password;
    }

    // Redis configuration
    if (const char* redis_host = std::getenv("REDIS_HOST")) {
        redis_config_.host = redis_host;
    }
    if (const char* redis_port = std::getenv("REDIS_PORT")) {
        redis_config_.port = std::atoi(redis_port);
    }
    if (const char* redis_password = std::getenv("REDIS_PASSWORD")) {
        redis_config_.password = redis_password;
    }
    if (const char* redis_db = std::getenv("REDIS_DB")) {
        redis_config_.database = std::atoi(redis_db);
    }

    // Application configuration
    if (const char* app_env = std::getenv("APP_ENV")) {
        app_config_.environment = app_env;
    }
    if (const char* log_level = std::getenv("LOG_LEVEL")) {
        // Store for logger initialization
    }
}

void Config::load_app_config(const boost::property_tree::ptree& pt) {
    if (auto app = pt.get_child_optional("Application")) {
        app_config_.app_name = app->get<std::string>("app_name", app_config_.app_name);
        app_config_.version = app->get<std::string>("version", app_config_.version);
        app_config_.build = app->get<std::string>("build", app_config_.build);
        app_config_.environment = app->get<std::string>("environment", app_config_.environment);
    }

    if (auto runtime = pt.get_child_optional("Runtime")) {
        app_config_.max_workers = runtime->get<int>("max_workers", app_config_.max_workers);
        app_config_.queue_size = runtime->get<int>("queue_size", app_config_.queue_size);
    }

    if (auto perf = pt.get_child_optional("Performance")) {
        app_config_.enable_monitoring = perf->get<bool>("enable_monitoring", app_config_.enable_monitoring);
    }
}

void Config::load_db_config(const boost::property_tree::ptree& pt) {
    // Environment variables take precedence
    bool has_env_db = std::getenv("DB_HOST") != nullptr;

    if (auto db = pt.get_child_optional("PostgreSQL")) {
        if (!has_env_db) {
            db_config_.host = db->get<std::string>("host", db_config_.host);
            db_config_.port = db->get<int>("port", db_config_.port);
            db_config_.database = db->get<std::string>("database", db_config_.database);
            db_config_.username = db->get<std::string>("username", db_config_.username);
            db_config_.password = db->get<std::string>("password", db_config_.password);
        }
        db_config_.pool_size = db->get<int>("pool_size", db_config_.pool_size);
        db_config_.max_connections = db->get<int>("max_connections", db_config_.max_connections);
    }

    bool has_env_redis = std::getenv("REDIS_HOST") != nullptr;

    if (auto redis = pt.get_child_optional("Redis")) {
        redis_config_.enabled = redis->get<bool>("enabled", redis_config_.enabled);

        if (!has_env_redis) {
            redis_config_.host = redis->get<std::string>("host", redis_config_.host);
            redis_config_.port = redis->get<int>("port", redis_config_.port);
            redis_config_.password = redis->get<std::string>("password", redis_config_.password);
            redis_config_.database = redis->get<int>("database", redis_config_.database);
        }
        redis_config_.pool_size = redis->get<int>("pool_size", redis_config_.pool_size);
    }
}

void Config::load_protocol_config(const boost::property_tree::ptree& pt) {
    if (auto smpp = pt.get_child_optional("SMPP")) {
        smpp_config_.enabled = smpp->get<bool>("enabled", smpp_config_.enabled);
        smpp_config_.bind_address = smpp->get<std::string>("bind_address", smpp_config_.bind_address);
        smpp_config_.bind_port = smpp->get<int>("bind_port", smpp_config_.bind_port);
        smpp_config_.system_id = smpp->get<std::string>("system_id", smpp_config_.system_id);
        smpp_config_.max_connections = smpp->get<int>("max_connections", smpp_config_.max_connections);
        smpp_config_.enquire_link_interval = smpp->get<int>("enquire_link_interval", smpp_config_.enquire_link_interval);
    }

    if (auto http = pt.get_child_optional("HTTP")) {
        api_config_.enabled = http->get<bool>("enabled", api_config_.enabled);
        api_config_.bind_address = http->get<std::string>("bind_address", api_config_.bind_address);
        api_config_.bind_port = http->get<int>("bind_port", api_config_.bind_port);
        api_config_.enable_https = http->get<bool>("enable_https", api_config_.enable_https);
        api_config_.ssl_cert_file = http->get<std::string>("ssl_cert_file", api_config_.ssl_cert_file);
        api_config_.ssl_key_file = http->get<std::string>("ssl_key_file", api_config_.ssl_key_file);
    }
}

void Config::load_security_config(const boost::property_tree::ptree& pt) {
    if (auto auth = pt.get_child_optional("Authentication")) {
        security_config_.access_token_expire_minutes = auth->get<int>("session_timeout",
            security_config_.access_token_expire_minutes);
    }

    if (auto pwd = pt.get_child_optional("Password_Policy")) {
        security_config_.password_min_length = pwd->get<int>("min_length",
            security_config_.password_min_length);
        security_config_.password_expiry_days = pwd->get<int>("password_expiry_days",
            security_config_.password_expiry_days);
    }
}

void Config::generate_secret_key() {
    std::random_device rd;
    std::mt19937 gen(rd());
    std::uniform_int_distribution<> dis(0, 255);

    std::ostringstream oss;
    for (int i = 0; i < 32; ++i) {
        oss << std::hex << std::setw(2) << std::setfill('0') << dis(gen);
    }

    security_config_.secret_key = oss.str();
}

template<typename T>
T Config::get(const std::string& key, const T& default_value) const {
    // Simple key-value retrieval
    // Could be enhanced with property_tree for nested access
    return default_value;
}

// Template instantiations
template std::string Config::get<std::string>(const std::string&, const std::string&) const;
template int Config::get<int>(const std::string&, const int&) const;
template bool Config::get<bool>(const std::string&, const bool&) const;

} // namespace protei::core
