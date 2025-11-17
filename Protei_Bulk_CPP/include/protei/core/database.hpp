/**
 * @file database.hpp
 * @brief PostgreSQL Database Connection Pool
 */

#pragma once

#include <memory>
#include <queue>
#include <mutex>
#include <condition_variable>
#include <pqxx/pqxx>
#include "config.hpp"

namespace protei::core {

/**
 * @brief PostgreSQL Connection Wrapper
 */
class Connection {
public:
    explicit Connection(const std::string& conn_str);
    ~Connection() = default;

    pqxx::connection& get() { return conn_; }
    bool is_open() const { return conn_.is_open(); }
    void reset();

private:
    pqxx::connection conn_;
};

/**
 * @brief Connection Pool (Singleton)
 */
class Database {
public:
    static Database& instance() {
        static Database instance;
        return instance;
    }

    Database(const Database&) = delete;
    Database& operator=(const Database&) = delete;

    void initialize(const DatabaseConfig& config);
    void shutdown();

    /**
     * @brief Get a connection from the pool
     */
    std::shared_ptr<Connection> get_connection();

    /**
     * @brief Return a connection to the pool
     */
    void return_connection(std::shared_ptr<Connection> conn);

    /**
     * @brief Execute a query and return results
     */
    template<typename Func>
    auto execute(Func&& func) -> decltype(func(std::declval<pqxx::connection&>())) {
        auto conn = get_connection();
        try {
            auto result = func(conn->get());
            return_connection(conn);
            return result;
        } catch (...) {
            return_connection(conn);
            throw;
        }
    }

    /**
     * @brief Execute a transaction
     */
    template<typename Func>
    auto transaction(Func&& func) -> decltype(func(std::declval<pqxx::work&>())) {
        auto conn = get_connection();
        try {
            pqxx::work txn(conn->get());
            auto result = func(txn);
            txn.commit();
            return_connection(conn);
            return result;
        } catch (...) {
            return_connection(conn);
            throw;
        }
    }

    size_t pool_size() const { return pool_size_; }
    size_t available_connections() const {
        std::lock_guard<std::mutex> lock(mutex_);
        return available_.size();
    }

private:
    Database() = default;
    ~Database();

    void create_pool();

    std::string connection_string_;
    size_t pool_size_ = 20;
    bool initialized_ = false;

    std::queue<std::shared_ptr<Connection>> available_;
    std::vector<std::shared_ptr<Connection>> all_connections_;
    mutable std::mutex mutex_;
    std::condition_variable cv_;
};

} // namespace protei::core
