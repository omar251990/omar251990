#pragma once
#include <string>
#include <memory>
#include "protei/services/routing_service.hpp"
namespace protei::protocols {
class SmppServer {
public:
    SmppServer(const std::string& host, int port) : host_(host), port_(port) {}
    void start() {}
    void stop() {}
    void set_routing_service(std::shared_ptr<services::RoutingService> service) { routing_service_ = service; }
private:
    std::string host_;
    int port_;
    std::shared_ptr<services::RoutingService> routing_service_;
};
}
