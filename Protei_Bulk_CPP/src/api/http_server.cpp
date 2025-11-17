#include "protei/api/http_server.hpp"
#include "protei/services/routing_service.hpp"
#include "protei/services/campaign_service.hpp"
#include <nlohmann/json.hpp>

using json = nlohmann::json;

namespace protei::api {

HttpServer::HttpServer(const std::string& host, int port)
    : host_(host), port_(port), server_(std::make_unique<httplib::Server>()), running_(false) {
    setup_middleware();
    setup_routes();
}

HttpServer::~HttpServer() {
    stop();
}

void HttpServer::setup_middleware() {
    // CORS
    server_->set_pre_routing_handler([](const httplib::Request& req, httplib::Response& res) {
        res.set_header("Access-Control-Allow-Origin", "*");
        res.set_header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS");
        res.set_header("Access-Control-Allow-Headers", "Content-Type, Authorization");
        if (req.method == "OPTIONS") {
            res.status = 200;
            return httplib::Server::HandlerResponse::Handled;
        }
        return httplib::Server::HandlerResponse::Unhandled;
    });
}

void HttpServer::setup_routes() {
    // Health check
    server_->Get("/api/v1/health", [](const httplib::Request&, httplib::Response& res) {
        json response = {
            {"status", "healthy"},
            {"version", "1.0.0"},
            {"timestamp", std::time(nullptr)}
        };
        res.set_content(response.dump(), "application/json");
    });

    // Root
    server_->Get("/", [](const httplib::Request&, httplib::Response& res) {
        json response = {
            {"message", "Protei_Bulk C++ API"},
            {"version", "1.0.0"},
            {"docs", "/api/docs"}
        };
        res.set_content(response.dump(), "application/json");
    });

    // Authentication endpoints (stubs)
    server_->Post("/api/v1/auth/login", [](const httplib::Request& req, httplib::Response& res) {
        json response = {
            {"access_token", "stub_token"},
            {"token_type", "bearer"},
            {"expires_in", 3600}
        };
        res.set_content(response.dump(), "application/json");
    });

    // Message endpoints (stubs)
    server_->Post("/api/v1/messages/send", [](const httplib::Request& req, httplib::Response& res) {
        json response = {
            {"message_id", "msg_" + std::to_string(std::time(nullptr))},
            {"status", "queued"}
        };
        res.set_content(response.dump(), "application/json");
    });

    // Campaign endpoints (stubs)
    server_->Get("/api/v1/campaigns", [](const httplib::Request&, httplib::Response& res) {
        json response = {
            {"campaigns", json::array()},
            {"total", 0}
        };
        res.set_content(response.dump(), "application/json");
    });
}

void HttpServer::start() {
    if (running_) return;
    running_ = true;
    server_->listen(host_, port_);
}

void HttpServer::stop() {
    if (!running_) return;
    server_->stop();
    running_ = false;
}

void HttpServer::register_routing_service(std::shared_ptr<services::RoutingService> service) {
    routing_service_ = service;
}

void HttpServer::register_campaign_service(std::shared_ptr<services::CampaignService> service) {
    campaign_service_ = service;
}

} // namespace protei::api
