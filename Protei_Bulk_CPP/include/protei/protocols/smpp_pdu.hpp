/**
 * @file smpp_pdu.hpp
 * @brief SMPP Protocol Data Unit definitions
 */

#pragma once

#include <cstdint>
#include <string>
#include <vector>
#include <array>

namespace protei::protocols {

// SMPP Command IDs
enum class SmppCommand : uint32_t {
    BIND_RECEIVER = 0x00000001,
    BIND_TRANSMITTER = 0x00000002,
    BIND_TRANSCEIVER = 0x00000009,
    BIND_RECEIVER_RESP = 0x80000001,
    BIND_TRANSMITTER_RESP = 0x80000002,
    BIND_TRANSCEIVER_RESP = 0x80000009,
    SUBMIT_SM = 0x00000004,
    SUBMIT_SM_RESP = 0x80000004,
    DELIVER_SM = 0x00000005,
    DELIVER_SM_RESP = 0x80000005,
    UNBIND = 0x00000006,
    UNBIND_RESP = 0x80000006,
    ENQUIRE_LINK = 0x00000015,
    ENQUIRE_LINK_RESP = 0x80000015,
    SUBMIT_MULTI = 0x00000021,
    SUBMIT_MULTI_RESP = 0x80000021,
    QUERY_SM = 0x00000003,
    QUERY_SM_RESP = 0x80000003,
    CANCEL_SM = 0x00000008,
    CANCEL_SM_RESP = 0x80000008
};

// SMPP Status codes
enum class SmppStatus : uint32_t {
    ESME_ROK = 0x00000000,              // No Error
    ESME_RINVMSGLEN = 0x00000001,       // Message Length is invalid
    ESME_RINVCMDLEN = 0x00000002,       // Command Length is invalid
    ESME_RINVCMDID = 0x00000003,        // Invalid Command ID
    ESME_RINVBNDSTS = 0x00000004,       // Incorrect BIND Status
    ESME_RALYBND = 0x00000005,          // ESME Already in Bound State
    ESME_RINVPRTFLG = 0x00000006,       // Invalid Priority Flag
    ESME_RINVREGDLVFLG = 0x00000007,    // Invalid Registered Delivery Flag
    ESME_RSYSERR = 0x00000008,          // System Error
    ESME_RINVSRCADR = 0x0000000A,       // Invalid Source Address
    ESME_RINVDSTADR = 0x0000000B,       // Invalid Destination Address
    ESME_RINVMSGID = 0x0000000C,        // Message ID is invalid
    ESME_RBINDFAIL = 0x0000000D,        // Bind Failed
    ESME_RINVPASWD = 0x0000000E,        // Invalid Password
    ESME_RINVSYSID = 0x0000000F,        // Invalid System ID
    ESME_RSUBMITFAIL = 0x00000045,      // submit_sm or submit_multi failed
    ESME_RTHROTTLED = 0x00000058        // Throttling error
};

// SMPP PDU Header
struct SmppHeader {
    uint32_t command_length;
    uint32_t command_id;
    uint32_t command_status;
    uint32_t sequence_number;

    SmppHeader() : command_length(0), command_id(0), command_status(0), sequence_number(0) {}
};

// Base PDU class
class SmppPdu {
public:
    SmppHeader header;

    SmppPdu() = default;
    virtual ~SmppPdu() = default;

    // Encode PDU to binary
    virtual std::vector<uint8_t> encode() const = 0;

    // Get PDU type
    virtual SmppCommand get_command() const = 0;
};

// Bind PDU
class BindPdu : public SmppPdu {
public:
    std::string system_id;
    std::string password;
    std::string system_type;
    uint8_t interface_version;
    uint8_t addr_ton;
    uint8_t addr_npi;
    std::string address_range;

    BindPdu() : interface_version(0x34), addr_ton(0), addr_npi(0) {}

    std::vector<uint8_t> encode() const override;
    SmppCommand get_command() const override { return SmppCommand::BIND_TRANSCEIVER; }
};

// Bind Response PDU
class BindRespPdu : public SmppPdu {
public:
    std::string system_id;

    std::vector<uint8_t> encode() const override;
    SmppCommand get_command() const override { return SmppCommand::BIND_TRANSCEIVER_RESP; }
};

// Submit SM PDU
class SubmitSmPdu : public SmppPdu {
public:
    std::string service_type;
    uint8_t source_addr_ton;
    uint8_t source_addr_npi;
    std::string source_addr;
    uint8_t dest_addr_ton;
    uint8_t dest_addr_npi;
    std::string destination_addr;
    uint8_t esm_class;
    uint8_t protocol_id;
    uint8_t priority_flag;
    std::string schedule_delivery_time;
    std::string validity_period;
    uint8_t registered_delivery;
    uint8_t replace_if_present_flag;
    uint8_t data_coding;
    uint8_t sm_default_msg_id;
    uint8_t sm_length;
    std::vector<uint8_t> short_message;

    SubmitSmPdu() : source_addr_ton(0), source_addr_npi(0),
                    dest_addr_ton(1), dest_addr_npi(1),
                    esm_class(0), protocol_id(0), priority_flag(0),
                    registered_delivery(1), replace_if_present_flag(0),
                    data_coding(0), sm_default_msg_id(0), sm_length(0) {}

    std::vector<uint8_t> encode() const override;
    SmppCommand get_command() const override { return SmppCommand::SUBMIT_SM; }
};

// Submit SM Response PDU
class SubmitSmRespPdu : public SmppPdu {
public:
    std::string message_id;

    std::vector<uint8_t> encode() const override;
    SmppCommand get_command() const override { return SmppCommand::SUBMIT_SM_RESP; }
};

// Deliver SM PDU
class DeliverSmPdu : public SmppPdu {
public:
    std::string service_type;
    uint8_t source_addr_ton;
    uint8_t source_addr_npi;
    std::string source_addr;
    uint8_t dest_addr_ton;
    uint8_t dest_addr_npi;
    std::string destination_addr;
    uint8_t esm_class;
    uint8_t protocol_id;
    uint8_t priority_flag;
    std::string schedule_delivery_time;
    std::string validity_period;
    uint8_t registered_delivery;
    uint8_t replace_if_present_flag;
    uint8_t data_coding;
    uint8_t sm_default_msg_id;
    uint8_t sm_length;
    std::vector<uint8_t> short_message;

    DeliverSmPdu() : source_addr_ton(1), source_addr_npi(1),
                     dest_addr_ton(0), dest_addr_npi(0),
                     esm_class(0), protocol_id(0), priority_flag(0),
                     registered_delivery(0), replace_if_present_flag(0),
                     data_coding(0), sm_default_msg_id(0), sm_length(0) {}

    std::vector<uint8_t> encode() const override;
    SmppCommand get_command() const override { return SmppCommand::DELIVER_SM; }
};

// Enquire Link PDU
class EnquireLinkPdu : public SmppPdu {
public:
    std::vector<uint8_t> encode() const override;
    SmppCommand get_command() const override { return SmppCommand::ENQUIRE_LINK; }
};

// Enquire Link Response PDU
class EnquireLinkRespPdu : public SmppPdu {
public:
    std::vector<uint8_t> encode() const override;
    SmppCommand get_command() const override { return SmppCommand::ENQUIRE_LINK_RESP; }
};

// Unbind PDU
class UnbindPdu : public SmppPdu {
public:
    std::vector<uint8_t> encode() const override;
    SmppCommand get_command() const override { return SmppCommand::UNBIND; }
};

// Unbind Response PDU
class UnbindRespPdu : public SmppPdu {
public:
    std::vector<uint8_t> encode() const override;
    SmppCommand get_command() const override { return SmppCommand::UNBIND_RESP; }
};

// PDU Parser
class SmppPduParser {
public:
    // Decode PDU from binary data
    static std::unique_ptr<SmppPdu> decode(const std::vector<uint8_t>& data);

    // Read header from binary data
    static SmppHeader read_header(const std::vector<uint8_t>& data);

private:
    static std::string read_c_string(const std::vector<uint8_t>& data, size_t& offset);
    static uint8_t read_uint8(const std::vector<uint8_t>& data, size_t& offset);
    static uint32_t read_uint32(const std::vector<uint8_t>& data, size_t& offset);
};

// Utility functions
std::vector<uint8_t> encode_c_string(const std::string& str);
std::vector<uint8_t> encode_uint8(uint8_t value);
std::vector<uint8_t> encode_uint32(uint32_t value);

} // namespace protei::protocols
