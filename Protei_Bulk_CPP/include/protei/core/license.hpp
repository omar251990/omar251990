/**
 * @file license.hpp
 * @brief Commercial License Management System
 *
 * Enterprise licensing with hardware binding and activation
 */

#pragma once

#include <string>
#include <chrono>
#include <vector>

namespace protei::core {

/**
 * @brief License information
 */
struct LicenseInfo {
    std::string license_key;
    std::string customer_name;
    std::string customer_id;
    std::string product_edition;  // Enterprise, Professional, Standard
    std::chrono::system_clock::time_point issue_date;
    std::chrono::system_clock::time_point expiry_date;

    // Feature limits
    int max_tps;                  // Maximum transactions per second
    int max_concurrent_campaigns;
    int max_users;
    int max_smsc_connections;
    bool unlimited_messages;

    // Enabled features
    bool enable_whatsapp;
    bool enable_email;
    bool enable_viber;
    bool enable_rcs;
    bool enable_voice;
    bool enable_ai_designer;
    bool enable_chatbot;
    bool enable_journey_automation;
    bool enable_multi_tenancy;

    // Hardware binding
    std::string machine_id;
    std::string cpu_id;
    std::string mac_address;

    // Activation
    bool is_activated;
    std::string activation_code;
    std::chrono::system_clock::time_point activation_date;

    // Validity
    bool is_valid;
    std::string validation_message;
};

/**
 * @brief License Manager
 */
class LicenseManager {
public:
    static LicenseManager& instance() {
        static LicenseManager instance;
        return instance;
    }

    LicenseManager(const LicenseManager&) = delete;
    LicenseManager& operator=(const LicenseManager&) = delete;

    /**
     * @brief Initialize license system
     */
    bool initialize(const std::string& license_file = "/opt/protei_bulk/config/license.key");

    /**
     * @brief Validate license
     */
    bool validate();

    /**
     * @brief Activate license
     */
    bool activate(const std::string& activation_code);

    /**
     * @brief Get license information
     */
    const LicenseInfo& get_license_info() const { return license_info_; }

    /**
     * @brief Check if feature is enabled
     */
    bool is_feature_enabled(const std::string& feature) const;

    /**
     * @brief Check TPS limit
     */
    bool check_tps_limit(int current_tps) const;

    /**
     * @brief Get days until expiry
     */
    int get_days_until_expiry() const;

    /**
     * @brief Is license expired
     */
    bool is_expired() const;

    /**
     * @brief Generate machine fingerprint
     */
    std::string get_machine_fingerprint() const;

private:
    LicenseManager() = default;

    bool load_license(const std::string& license_file);
    bool validate_signature(const std::string& license_data);
    std::string decrypt_license(const std::string& encrypted_data);
    std::string get_cpu_id() const;
    std::string get_mac_address() const;
    std::string calculate_machine_id() const;

    LicenseInfo license_info_;
    std::string license_file_path_;
    bool initialized_ = false;
};

/**
 * @brief License Exception
 */
class LicenseException : public std::runtime_error {
public:
    explicit LicenseException(const std::string& message)
        : std::runtime_error(message) {}
};

} // namespace protei::core
