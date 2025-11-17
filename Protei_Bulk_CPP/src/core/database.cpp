/**
 * @file database.cpp
 * @brief Database Connection Pool Implementation
 */

#include "protei/core/database.hpp"
#include <stdexcept>
#include <chrono>

namespace protei::core {

Connection::Connection(const std::string& conn_str)
    : conn_(conn_str) {
    if (!conn_.is_open()) {
        throw std::runtime_error("Failed to open database connection");
    }
}

void Connection::reset() {
    // Could add connection reset logic here if needed
    if (!conn_.is_open()) {
        throw std::runtime_error("Connection is closed");
    }
}

Database::~Database() {
    shutdown();
}

void Database::initialize(const DatabaseConfig& config) {
    std::lock_guard<std::mutex> lock(mutex_);

    if (initialized_) {
        return;
    }

    connection_string_ = config.get_connection_string();
    pool_size_ = config.pool_size;

    create_pool();
    initialized_ = true;
}

void Database::create_pool() {
    all_connections_.clear();
    while (!available_.empty()) {
        available_.pop();
    }

    for (size_t i = 0; i < pool_size_; ++i) {
        try {
            auto conn = std::make_shared<Connection>(connection_string_);
            all_connections_.push_back(conn);
            available_.push(conn);
        } catch (const std::exception& e) {
            throw std::runtime_error("Failed to create connection pool: " + std::string(e.what()));
        }
    }
}

void Database::shutdown() {
    std::lock_guard<std::mutex> lock(mutex_);

    if (!initialized_) {
        return;
    }

    // Clear the pool
    while (!available_.empty()) {
        available_.pop();
    }

    all_connections_.clear();
    initialized_ = false;
}

std::shared_ptr<Connection> Database::get_connection() {
    std::unique_lock<std::mutex> lock(mutex_);

    // Wait for available connection with timeout
    if (available_.empty()) {
        bool got_connection = cv_.wait_for(lock, std::chrono::seconds(30),
            [this] { return !available_.empty(); });

        if (!got_connection) {
            throw std::runtime_error("Connection pool timeout - no connections available");
        }
    }

    auto conn = available_.front();
    available_.pop();

    // Verify connection is still open
    if (!conn->is_open()) {
        // Recreate connection
        conn = std::make_shared<Connection>(connection_string_);
    }

    return conn;
}

void Database::return_connection(std::shared_ptr<Connection> conn) {
    if (conn && conn->is_open()) {
        std::lock_guard<std::mutex> lock(mutex_);
        available_.push(conn);
        cv_.notify_one();
    }
}

} // namespace protei::core
