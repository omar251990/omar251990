/**
 * @file http_server.hpp
 * @brief HTTP API Server using cpp-httplib
 */

#pragma once

#include <memory>
#include <string>
#include <functional>
#include <httplib.h>

namespace protei {
    namespace services {
        class RoutingService;
        class CampaignService;
    }
}

namespace protei::api {

class HttpServer {
public:
    HttpServer(const std::string& host, int port);
    ~HttpServer();

    void start();
    void stop();

    // Register service dependencies
    void register_routing_service(std::shared_ptr<services::RoutingService> service);
    void register_campaign_service(std::shared_ptr<services::CampaignService> service);

private:
    void setup_routes();
    void setup_middleware();

    std::string host_;
    int port_;
    std::unique_ptr<httplib::Server> server_;
    bool running_;

    // Service dependencies
    std::shared_ptr<services::RoutingService> routing_service_;
    std::shared_ptr<services::CampaignService> campaign_service_;
};

} // namespace protei::api
