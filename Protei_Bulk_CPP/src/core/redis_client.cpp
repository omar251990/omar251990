/**
 * @file redis_client.cpp
 * @brief Redis Client Implementation
 */

#include "protei/core/redis_client.hpp"
#include <stdexcept>

namespace protei::core {

void RedisClient::initialize(const RedisConfig& config) {
    if (initialized_) {
        return;
    }

    if (!config.enabled) {
        return;
    }

    try {
        sw::redis::ConnectionOptions conn_opts;
        conn_opts.host = config.host;
        conn_opts.port = config.port;
        conn_opts.db = config.database;

        if (!config.password.empty()) {
            conn_opts.password = config.password;
        }

        conn_opts.socket_timeout = std::chrono::milliseconds(config.timeout_ms);

        sw::redis::ConnectionPoolOptions pool_opts;
        pool_opts.size = config.pool_size;

        redis_ = std::make_unique<sw::redis::Redis>(conn_opts, pool_opts);

        // Test connection
        redis_->ping();

        initialized_ = true;

    } catch (const std::exception& e) {
        throw std::runtime_error("Failed to initialize Redis: " + std::string(e.what()));
    }
}

void RedisClient::shutdown() {
    redis_.reset();
    initialized_ = false;
}

bool RedisClient::set(const std::string& key, const std::string& value) {
    if (!initialized_) return false;
    try {
        redis_->set(key, value);
        return true;
    } catch (...) {
        return false;
    }
}

bool RedisClient::set(const std::string& key, const std::string& value, std::chrono::seconds ttl) {
    if (!initialized_) return false;
    try {
        redis_->set(key, value, ttl);
        return true;
    } catch (...) {
        return false;
    }
}

std::optional<std::string> RedisClient::get(const std::string& key) {
    if (!initialized_) return std::nullopt;
    try {
        auto val = redis_->get(key);
        if (val) {
            return *val;
        }
        return std::nullopt;
    } catch (...) {
        return std::nullopt;
    }
}

bool RedisClient::del(const std::string& key) {
    if (!initialized_) return false;
    try {
        return redis_->del(key) > 0;
    } catch (...) {
        return false;
    }
}

bool RedisClient::exists(const std::string& key) {
    if (!initialized_) return false;
    try {
        return redis_->exists(key) > 0;
    } catch (...) {
        return false;
    }
}

bool RedisClient::hset(const std::string& key, const std::string& field, const std::string& value) {
    if (!initialized_) return false;
    try {
        return redis_->hset(key, field, value);
    } catch (...) {
        return false;
    }
}

std::optional<std::string> RedisClient::hget(const std::string& key, const std::string& field) {
    if (!initialized_) return std::nullopt;
    try {
        auto val = redis_->hget(key, field);
        if (val) {
            return *val;
        }
        return std::nullopt;
    } catch (...) {
        return std::nullopt;
    }
}

std::unordered_map<std::string, std::string> RedisClient::hgetall(const std::string& key) {
    if (!initialized_) return {};
    try {
        std::unordered_map<std::string, std::string> result;
        redis_->hgetall(key, std::inserter(result, result.begin()));
        return result;
    } catch (...) {
        return {};
    }
}

bool RedisClient::hdel(const std::string& key, const std::string& field) {
    if (!initialized_) return false;
    try {
        return redis_->hdel(key, field) > 0;
    } catch (...) {
        return false;
    }
}

long long RedisClient::lpush(const std::string& key, const std::string& value) {
    if (!initialized_) return 0;
    try {
        return redis_->lpush(key, value);
    } catch (...) {
        return 0;
    }
}

long long RedisClient::rpush(const std::string& key, const std::string& value) {
    if (!initialized_) return 0;
    try {
        return redis_->rpush(key, value);
    } catch (...) {
        return 0;
    }
}

std::optional<std::string> RedisClient::lpop(const std::string& key) {
    if (!initialized_) return std::nullopt;
    try {
        auto val = redis_->lpop(key);
        if (val) {
            return *val;
        }
        return std::nullopt;
    } catch (...) {
        return std::nullopt;
    }
}

std::optional<std::string> RedisClient::rpop(const std::string& key) {
    if (!initialized_) return std::nullopt;
    try {
        auto val = redis_->rpop(key);
        if (val) {
            return *val;
        }
        return std::nullopt;
    } catch (...) {
        return std::nullopt;
    }
}

long long RedisClient::llen(const std::string& key) {
    if (!initialized_) return 0;
    try {
        return redis_->llen(key);
    } catch (...) {
        return 0;
    }
}

bool RedisClient::sadd(const std::string& key, const std::string& member) {
    if (!initialized_) return false;
    try {
        return redis_->sadd(key, member) > 0;
    } catch (...) {
        return false;
    }
}

bool RedisClient::sismember(const std::string& key, const std::string& member) {
    if (!initialized_) return false;
    try {
        return redis_->sismember(key, member);
    } catch (...) {
        return false;
    }
}

std::unordered_set<std::string> RedisClient::smembers(const std::string& key) {
    if (!initialized_) return {};
    try {
        std::unordered_set<std::string> result;
        redis_->smembers(key, std::inserter(result, result.begin()));
        return result;
    } catch (...) {
        return {};
    }
}

bool RedisClient::zadd(const std::string& key, double score, const std::string& member) {
    if (!initialized_) return false;
    try {
        return redis_->zadd(key, member, score) > 0;
    } catch (...) {
        return false;
    }
}

std::vector<std::string> RedisClient::zrange(const std::string& key, long long start, long long stop) {
    if (!initialized_) return {};
    try {
        std::vector<std::string> result;
        redis_->zrange(key, start, stop, std::back_inserter(result));
        return result;
    } catch (...) {
        return {};
    }
}

bool RedisClient::ping() {
    if (!initialized_) return false;
    try {
        redis_->ping();
        return true;
    } catch (...) {
        return false;
    }
}

long long RedisClient::incr(const std::string& key) {
    if (!initialized_) return 0;
    try {
        return redis_->incr(key);
    } catch (...) {
        return 0;
    }
}

long long RedisClient::decr(const std::string& key) {
    if (!initialized_) return 0;
    try {
        return redis_->decr(key);
    } catch (...) {
        return 0;
    }
}

bool RedisClient::expire(const std::string& key, std::chrono::seconds ttl) {
    if (!initialized_) return false;
    try {
        return redis_->expire(key, ttl);
    } catch (...) {
        return false;
    }
}

void RedisClient::publish(const std::string& channel, const std::string& message) {
    if (!initialized_) return;
    try {
        redis_->publish(channel, message);
    } catch (...) {
        // Ignore
    }
}

} // namespace protei::core
