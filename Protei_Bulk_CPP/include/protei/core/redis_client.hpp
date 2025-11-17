/**
 * @file redis_client.hpp
 * @brief Redis Client Wrapper
 */

#pragma once

#include <memory>
#include <string>
#include <vector>
#include <optional>
#include <chrono>
#include <sw/redis++/redis++.h>
#include "config.hpp"

namespace protei::core {

/**
 * @brief Redis Client Singleton
 */
class RedisClient {
public:
    static RedisClient& instance() {
        static RedisClient instance;
        return instance;
    }

    RedisClient(const RedisClient&) = delete;
    RedisClient& operator=(const RedisClient&) = delete;

    void initialize(const RedisConfig& config);
    void shutdown();

    // String operations
    bool set(const std::string& key, const std::string& value);
    bool set(const std::string& key, const std::string& value, std::chrono::seconds ttl);
    std::optional<std::string> get(const std::string& key);
    bool del(const std::string& key);
    bool exists(const std::string& key);

    // Hash operations
    bool hset(const std::string& key, const std::string& field, const std::string& value);
    std::optional<std::string> hget(const std::string& key, const std::string& field);
    std::unordered_map<std::string, std::string> hgetall(const std::string& key);
    bool hdel(const std::string& key, const std::string& field);

    // List operations
    long long lpush(const std::string& key, const std::string& value);
    long long rpush(const std::string& key, const std::string& value);
    std::optional<std::string> lpop(const std::string& key);
    std::optional<std::string> rpop(const std::string& key);
    long long llen(const std::string& key);

    // Set operations
    bool sadd(const std::string& key, const std::string& member);
    bool sismember(const std::string& key, const std::string& member);
    std::unordered_set<std::string> smembers(const std::string& key);

    // Sorted set operations
    bool zadd(const std::string& key, double score, const std::string& member);
    std::vector<std::string> zrange(const std::string& key, long long start, long long stop);

    // Utility
    bool ping();
    long long incr(const std::string& key);
    long long decr(const std::string& key);
    bool expire(const std::string& key, std::chrono::seconds ttl);

    // Pub/Sub
    void publish(const std::string& channel, const std::string& message);

private:
    RedisClient() = default;
    ~RedisClient() = default;

    std::unique_ptr<sw::redis::Redis> redis_;
    bool initialized_ = false;
};

} // namespace protei::core
