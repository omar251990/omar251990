#pragma once
#include "protei/core/database.hpp"
#include "protei/core/redis_client.hpp"
namespace protei::services {
class CampaignService {
public:
    CampaignService(core::Database& db, core::RedisClient& redis) : db_(db), redis_(redis) {}
private:
    core::Database& db_;
    core::RedisClient& redis_;
};
}
