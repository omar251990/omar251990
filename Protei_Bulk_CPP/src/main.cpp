/**
 * @file main.cpp
 * @brief Protei_Bulk C++ - Main Entry Point
 * @version 1.0.0
 *
 * Enterprise Bulk Messaging Platform
 * High-performance C++ implementation with multi-channel support
 */

#include <iostream>
#include <memory>
#include <csignal>
#include <thread>
#include <vector>

#include "protei/core/config.hpp"
#include "protei/core/logger.hpp"
#include "protei/core/database.hpp"
#include "protei/core/redis_client.hpp"
#include "protei/api/http_server.hpp"
#include "protei/protocols/smpp_server.hpp"
#include "protei/services/routing_service.hpp"
#include "protei/services/campaign_service.hpp"

using namespace protei;

// Global instances
std::unique_ptr<core::Logger> g_logger;
std::unique_ptr<api::HttpServer> g_http_server;
std::unique_ptr<protocols::SmppServer> g_smpp_server;

// Signal handler
volatile std::sig_atomic_t g_shutdown_requested = 0;

void signal_handler(int signal) {
    if (signal == SIGINT || signal == SIGTERM) {
        g_shutdown_requested = 1;
        std::cout << "\nShutdown signal received..." << std::endl;
    }
}

void print_banner() {
    std::cout << R"(
╔═══════════════════════════════════════════════════════════╗
║                                                           ║
║     ██████╗ ██████╗  ██████╗ ████████╗███████╗██╗        ║
║     ██╔══██╗██╔══██╗██╔═══██╗╚══██╔══╝██╔════╝██║        ║
║     ██████╔╝██████╔╝██║   ██║   ██║   █████╗  ██║        ║
║     ██╔═══╝ ██╔══██╗██║   ██║   ██║   ██╔══╝  ██║        ║
║     ██║     ██║  ██║╚██████╔╝   ██║   ███████╗██║        ║
║     ╚═╝     ╚═╝  ╚═╝ ╚═════╝    ╚═╝   ╚══════╝╚═╝        ║
║                                                           ║
║     Enterprise Bulk Messaging Platform - C++ Edition     ║
║     Version 1.0.0 | Build: 001                           ║
║     High-Performance Multi-Channel Messaging             ║
║                                                           ║
╚═══════════════════════════════════════════════════════════╝
)" << std::endl;
}

void print_system_info() {
    std::cout << "System Information:" << std::endl;
    std::cout << "  CPU Cores: " << std::thread::hardware_concurrency() << std::endl;
    std::cout << "  C++ Standard: C++" << __cplusplus / 100 % 100 << std::endl;

    #ifdef __OPTIMIZE__
    std::cout << "  Build Mode: Release (Optimized)" << std::endl;
    #else
    std::cout << "  Build Mode: Debug" << std::endl;
    #endif

    std::cout << std::endl;
}

int main(int argc, char* argv[]) {
    try {
        // Print banner
        print_banner();
        print_system_info();

        // Register signal handlers
        std::signal(SIGINT, signal_handler);
        std::signal(SIGTERM, signal_handler);

        // Initialize logger
        g_logger = std::make_unique<core::Logger>("protei_bulk");
        g_logger->info("Starting Protei_Bulk C++ Edition...");

        // Load configuration
        g_logger->info("Loading configuration...");
        auto& config = core::Config::instance();

        if (argc > 1) {
            config.load_from_file(argv[1]);
        } else {
            config.load_from_file("config/app.conf");
        }

        g_logger->info("Configuration loaded successfully");
        g_logger->info("Environment: {}", config.get_app_environment());

        // Initialize database
        g_logger->info("Initializing database connection pool...");
        auto& db = core::Database::instance();
        db.initialize(config.get_database_config());
        g_logger->info("Database pool initialized: {} connections",
                      config.get_database_config().pool_size);

        // Initialize Redis
        g_logger->info("Connecting to Redis...");
        auto& redis = core::RedisClient::instance();
        redis.initialize(config.get_redis_config());
        g_logger->info("Redis connected: {}:{}",
                      config.get_redis_config().host,
                      config.get_redis_config().port);

        // Initialize services
        g_logger->info("Initializing business services...");

        auto routing_service = std::make_shared<services::RoutingService>(db, redis);
        auto campaign_service = std::make_shared<services::CampaignService>(db, redis);

        g_logger->info("Services initialized");

        // Start HTTP API Server
        if (config.is_http_enabled()) {
            g_logger->info("Starting HTTP API server...");
            g_http_server = std::make_unique<api::HttpServer>(
                config.get_api_bind_address(),
                config.get_api_bind_port()
            );

            // Register service dependencies
            g_http_server->register_routing_service(routing_service);
            g_http_server->register_campaign_service(campaign_service);

            g_http_server->start();
            g_logger->info("HTTP API listening on {}:{}",
                          config.get_api_bind_address(),
                          config.get_api_bind_port());
        }

        // Start SMPP Server
        if (config.is_smpp_enabled()) {
            g_logger->info("Starting SMPP server...");
            g_smpp_server = std::make_unique<protocols::SmppServer>(
                config.get_smpp_bind_address(),
                config.get_smpp_bind_port()
            );

            g_smpp_server->set_routing_service(routing_service);
            g_smpp_server->start();
            g_logger->info("SMPP server listening on {}:{}",
                          config.get_smpp_bind_address(),
                          config.get_smpp_bind_port());
        }

        // Print startup summary
        std::cout << "\n╔═══════════════════════════════════════════════════════════╗" << std::endl;
        std::cout << "║  ✓ Protei_Bulk is now running                            ║" << std::endl;
        std::cout << "╠═══════════════════════════════════════════════════════════╣" << std::endl;

        if (config.is_http_enabled()) {
            std::cout << "║  API:  http://" << config.get_api_bind_address() << ":"
                     << config.get_api_bind_port() << "/api/v1                   ║" << std::endl;
            std::cout << "║  Docs: http://" << config.get_api_bind_address() << ":"
                     << config.get_api_bind_port() << "/api/docs                ║" << std::endl;
        }

        if (config.is_smpp_enabled()) {
            std::cout << "║  SMPP: " << config.get_smpp_bind_address() << ":"
                     << config.get_smpp_bind_port() << "                                   ║" << std::endl;
        }

        std::cout << "╠═══════════════════════════════════════════════════════════╣" << std::endl;
        std::cout << "║  Press Ctrl+C to stop                                    ║" << std::endl;
        std::cout << "╚═══════════════════════════════════════════════════════════╝" << std::endl;
        std::cout << std::endl;

        g_logger->info("Startup complete - All systems operational");

        // Main loop - wait for shutdown signal
        while (!g_shutdown_requested) {
            std::this_thread::sleep_for(std::chrono::seconds(1));

            // Health monitoring could go here
        }

        // Graceful shutdown
        g_logger->info("Initiating graceful shutdown...");

        if (g_smpp_server) {
            g_logger->info("Stopping SMPP server...");
            g_smpp_server->stop();
        }

        if (g_http_server) {
            g_logger->info("Stopping HTTP server...");
            g_http_server->stop();
        }

        g_logger->info("Closing Redis connection...");
        redis.shutdown();

        g_logger->info("Closing database connections...");
        db.shutdown();

        g_logger->info("Shutdown complete. Goodbye!");

        return 0;

    } catch (const std::exception& e) {
        std::cerr << "Fatal error: " << e.what() << std::endl;
        if (g_logger) {
            g_logger->error("Fatal error: {}", e.what());
        }
        return 1;
    }
}
