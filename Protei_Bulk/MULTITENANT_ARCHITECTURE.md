# Protei_Bulk Multi-Tenant Architecture

## Overview

Protei_Bulk implements a comprehensive multi-tenant permission and control system where:
- Each Customer (Tenant) is completely isolated
- Each Customer contains its own hierarchy: Admin → Resellers → End-Users → Approvers
- Global System Admin (Protei HQ) can manage all customers and their users
- Fine-grained permission control with inheritance logic

## Architecture Components

### 1. Tenant Structure

```
Global System Admin (Protei HQ)
└── Customer (Tenant)
    ├── Customer Admin
    │   ├── Resellers
    │   │   ├── End Users
    │   │   └── Approvers
    │   └── End Users
    └── Auditors
```

#### User Hierarchy Levels:
- **Level 0**: Global System Admin
- **Level 10**: Customer Admin
- **Level 20**: Reseller
- **Level 30**: End User
- **Level 40**: Approver
- **Level 50**: Auditor

### 2. Database Schema

#### Core Tables

**tbl_customers**
- Primary tenant/customer table
- Contains quotas, limits, billing info
- Branding and configuration
- License management
- SMSC and routing policies

**tbl_roles**
- Predefined system roles
- Custom customer roles
- Hierarchy level definition
- Scope (GLOBAL, CUSTOMER, RESELLER)

**tbl_permissions**
- Granular permission definitions
- Module and action based
- 60+ predefined permissions covering:
  - User Management
  - Campaign Management
  - Template Management
  - Contact Management
  - Reporting
  - SMSC/Routing
  - Billing
  - System Configuration
  - Audit

**tbl_role_permissions**
- Default permission mappings for roles
- Allow/Deny flag

**tbl_user_permissions_override**
- User-specific permission overrides
- Expiration support
- Audit trail (granted_by, granted_at)

**tbl_customer_config**
- Customer-specific configuration key-value store
- Support for encrypted values

**tbl_permission_audit**
- Complete audit log of permission checks
- Permission grants/revokes
- IP and user agent tracking

#### Schema Files
- `/database/multitenant_schema.sql` - Complete SQL schema with triggers and functions

### 3. Permission Resolution Algorithm

```python
def check_permission(user_id, permission_code):
    # Step 1: Check user-level override
    override = get_user_permission_override(user_id, permission_code)
    if override exists and not expired:
        return override.allow

    # Step 2: Check role-level permission
    role_permission = get_role_permission(user.role_id, permission_code)
    if role_permission exists:
        return role_permission.allow

    # Step 3: Check customer-level config (optional)
    customer_config = get_customer_config(user.customer_id, permission_code)
    if customer_config exists:
        return customer_config.value

    # Step 4: Default deny
    return False
```

### 4. Python Models

**Location**: `/src/models/multitenant.py`

Models implemented:
- `Customer` - Tenant/customer entity
- `Role` - Role definitions
- `Permission` - Permission definitions
- `RolePermission` - Role-permission mappings
- `UserPermissionOverride` - User-specific overrides
- `CustomerConfig` - Customer configuration
- `PermissionAudit` - Audit log

All models include proper relationships, indexes, and constraints.

### 5. Services

#### Permission Service
**Location**: `/src/services/permission_service.py`

Key methods:
```python
check_permission(user_id, permission_code) -> bool
check_permissions(user_id, permission_codes) -> dict
grant_permission(user_id, permission_code, granted_by)
revoke_permission(user_id, permission_code, revoked_by)
get_user_permissions(user_id) -> list
assign_role_permissions(role_id, permission_codes)
get_role_permissions(role_id) -> list
get_permission_audit_log(customer_id, user_id) -> list
```

Features:
- Permission resolution with inheritance
- Audit logging of all permission checks
- Grant/revoke with expiration support
- Bulk permission assignment

#### Customer Service
**Location**: `/src/services/customer_service.py`

Key methods:
```python
create_customer(...) -> Customer
update_customer(customer_id, **kwargs) -> Customer
get_customer(customer_id) -> Customer
list_customers(status, license_type) -> list
suspend_customer(customer_id, reason)
activate_customer(customer_id)
delete_customer(customer_id)
get_customer_statistics(customer_id) -> dict
set_customer_config(customer_id, key, value)
get_customer_config(customer_id, key)
update_customer_balance(customer_id, amount)
extend_license(customer_id, days)
```

Features:
- Complete customer lifecycle management
- Quota and limit enforcement
- License management
- Balance management (prepaid/postpaid)
- Configuration management
- Usage statistics

### 6. Permission Matrix

| Module | Action | Global Admin | Customer Admin | Reseller | End User | Approver | Auditor |
|--------|--------|--------------|----------------|----------|----------|----------|---------|
| User Mgmt | Create/Edit/Delete | ✔ | ✔ (tenant) | ✔ (sub) | ❌ | ❌ | ❌ |
| Permissions | Assign/Revoke | ✔ | ✔ (tenant) | ❌ | ❌ | ❌ | ❌ |
| Campaigns | Create/Edit/Delete | ✔ | ✔ | ✔ | ✔ | ❌ | View |
| Campaigns | Approve/Reject | ✔ | ✔ | ❌ | ❌ | ✔ | View |
| Reports | View/Export | ✔ | ✔ | ✔ | ✔ | ✔ | ✔ |
| SMSC | Add/Edit | ✔ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Quota | Assign Credits/TPS | ✔ | ✔ | ✔ (sub) | ❌ | ❌ | ❌ |
| Templates | Create/Share | ✔ | ✔ | ✔ | ✔ | View | View |
| System Config | Backup/Scheduler | ✔ | ❌ | ❌ | ❌ | ❌ | ❌ |
| Logs | View/Export | ✔ | ✔ (tenant) | View | ❌ | View | ✔ |

### 7. API Endpoints (To Be Implemented)

#### Customer Management (Global Admin)
```
POST   /api/v1/admin/customers              # Create customer
GET    /api/v1/admin/customers              # List all customers
GET    /api/v1/admin/customers/{id}         # Get customer details
PUT    /api/v1/admin/customers/{id}         # Update customer
DELETE /api/v1/admin/customers/{id}         # Delete customer
POST   /api/v1/admin/customers/{id}/suspend # Suspend customer
POST   /api/v1/admin/customers/{id}/activate # Activate customer
GET    /api/v1/admin/customers/{id}/stats   # Customer statistics
POST   /api/v1/admin/customers/{id}/config  # Set customer config
GET    /api/v1/admin/customers/{id}/config  # Get customer config
```

#### Customer Admin Portal
```
GET    /api/v1/customer/{id}/users              # List users
POST   /api/v1/customer/{id}/users              # Create user
PUT    /api/v1/customer/{id}/users/{user_id}    # Update user
DELETE /api/v1/customer/{id}/users/{user_id}    # Delete user

GET    /api/v1/customer/{id}/permissions        # Get all permissions
POST   /api/v1/customer/{id}/permissions/assign # Assign permissions
POST   /api/v1/customer/{id}/permissions/revoke # Revoke permissions

GET    /api/v1/customer/{id}/audit              # View audit logs
GET    /api/v1/customer/{id}/activity           # Customer activity
```

#### Permission Management
```
GET    /api/v1/permissions                   # List all permissions
GET    /api/v1/permissions/user/{user_id}    # Get user permissions
POST   /api/v1/permissions/grant             # Grant permission
POST   /api/v1/permissions/revoke            # Revoke permission
GET    /api/v1/permissions/audit             # Permission audit log

GET    /api/v1/roles                         # List roles
GET    /api/v1/roles/{role_id}/permissions   # Get role permissions
POST   /api/v1/roles/{role_id}/permissions   # Assign role permissions
```

### 8. Web UI Components (To Be Implemented)

#### Global Admin Control Panel
**Route**: `/admin/customers`

Features:
- Customer overview dashboard
- Create/edit customer wizard
- Assign customer admin
- Suspend/resume tenant
- View tenant activity
- Export tenant data
- Clone tenant template

#### Customer Admin Portal
**Route**: `/customer/{id}/admin`

Features:
- User management (add/edit/delete/disable)
- Role assignment
- Permission control matrix
- Reseller quota control
- Campaign approval settings
- Customer-specific routing
- Data access control
- Customer logs and alerts

#### Permission Editor
**Route**: `/customer/{id}/permissions`

Features:
- Tree view of all modules
- Toggle switches for Read/Write/Delete/Approve
- "Apply to All Users" button
- "Clone Role Permissions"
- Export/Import permission profiles
- Temporary role lockout

### 9. Security Features

1. **Tenant Isolation**
   - All queries filtered by `customer_id`
   - Row-level security enforcement
   - Separate database schemas (optional)

2. **Permission Checking**
   - Every API endpoint checks permissions
   - Audit logging of all permission checks
   - IP and user agent tracking

3. **Role Hierarchy**
   - Users can only manage lower-level users
   - Permission propagation control
   - Inheritance logic prevents privilege escalation

4. **Encrypted Storage**
   - Sensitive configuration encrypted
   - API keys hashed
   - Password bcrypt hashing

5. **Audit Trail**
   - Complete audit log of all actions
   - Permission grants/revokes logged
   - User activity tracking
   - SIEM integration ready

### 10. Usage Examples

#### Creating a New Customer

```python
from src.services.customer_service import customer_service

# Global admin creates new customer
customer = customer_service.create_customer(
    db=db,
    customer_code="UMNIAH",
    customer_name="Umniah Telecom",
    company_name="Umniah Mobile Company",
    contact_email="admin@umniah.com",
    license_type="PREMIUM",
    max_users=500,
    max_tps=5000,
    created_by="global_admin"
)

# Set customer configuration
customer_service.set_customer_config(
    db=db,
    customer_id=customer.customer_id,
    config_key="require_campaign_approval",
    config_value="true",
    config_type="BOOLEAN"
)
```

#### Checking Permissions

```python
from src.services.permission_service import permission_service

# Check if user can create campaign
can_create = permission_service.check_permission(
    db=db,
    user_id=123,
    permission_code="CAMPAIGN_CREATE"
)

# Check multiple permissions
permissions = permission_service.check_permissions(
    db=db,
    user_id=123,
    permission_codes=["CAMPAIGN_CREATE", "CAMPAIGN_EDIT", "CAMPAIGN_DELETE"]
)
```

#### Granting Permissions

```python
# Grant permission to specific user
permission_service.grant_permission(
    db=db,
    user_id=456,
    permission_code="REPORT_EXPORT",
    granted_by_user_id=123,
    expires_at=datetime(2025, 12, 31),
    reason="Temporary access for Q4 reporting"
)

# Assign permissions to role
permission_service.assign_role_permissions(
    db=db,
    role_id=3,  # End User role
    permission_codes=[
        "CAMPAIGN_VIEW",
        "CAMPAIGN_CREATE",
        "TEMPLATE_VIEW",
        "REPORT_VIEW"
    ]
)
```

### 11. Database Functions

#### check_user_permission(user_id, permission_code)
SQL function for efficient permission checking directly in database.

#### get_customer_user_hierarchy(customer_id)
Returns complete user hierarchy tree for a customer.

### 12. Triggers

- **update_customer_activity**: Updates `last_activity_at` when campaigns or messages created
- Automatic timestamp updates on all tables

### 13. Implementation Status

✅ **Completed**:
- Multi-tenant database schema
- Python models for all entities
- Permission service with resolution logic
- Customer service with full lifecycle
- SQL functions and triggers
- Audit logging framework

🚧 **In Progress**:
- FastAPI endpoints for customer management
- FastAPI endpoints for permission management
- Web UI components

📋 **To Do**:
- Customer admin control panel UI
- Permission editor UI
- Tenant switching in web portal
- Bulk user import/export
- Role template library
- Dynamic approval chains

### 14. Configuration

#### Customer Configuration Keys

Common configuration keys:
- `require_campaign_approval` - Boolean, requires maker-checker for campaigns
- `max_campaign_size` - Integer, maximum recipients per campaign
- `allowed_sending_hours` - JSON, allowed time windows
- `dlr_callback_url` - String, webhook for delivery reports
- `default_sender_id` - String, default sender ID
- `enable_api_access` - Boolean, allow API access
- `daily_message_limit` - Integer, daily message cap

### 15. Best Practices

1. **Always filter by customer_id**: All queries must include customer_id to enforce isolation

2. **Check permissions before operations**: Use permission service in all endpoints

3. **Log sensitive operations**: Use audit logging for all permission changes

4. **Validate user hierarchy**: Users can only manage users at lower levels

5. **Use transactions**: Wrap related operations in database transactions

6. **Handle license expiry**: Check customer status and expiry before operations

7. **Monitor quotas**: Check TPS, user, and campaign limits

8. **Encrypt sensitive data**: Use encryption for API keys and secrets

### 16. Migration Guide

#### Migrating Existing Data

1. Create Global Admin customer (PROTEI_HQ)
2. Assign existing admin user to global customer
3. Create customer for each existing account
4. Migrate users with customer_id assignment
5. Assign default roles based on current permissions
6. Migrate campaigns, messages, templates with customer_id

```sql
-- Example migration script
INSERT INTO tbl_customers (customer_code, customer_name, status)
VALUES ('EXISTING', 'Existing Customer', 'ACTIVE');

UPDATE users SET customer_id = (SELECT customer_id FROM tbl_customers WHERE customer_code = 'EXISTING');
UPDATE campaigns SET customer_id = (SELECT customer_id FROM tbl_customers WHERE customer_code = 'EXISTING');
-- etc.
```

### 17. Performance Considerations

1. **Indexes**: All customer_id foreign keys have indexes
2. **Query optimization**: Use customer_id in WHERE clauses
3. **Caching**: Permission results can be cached (with TTL)
4. **Partitioning**: Consider table partitioning by customer_id for very large deployments
5. **Connection pooling**: Separate connection pools per customer (optional)

### 18. Troubleshooting

#### Common Issues

**Permission denied errors**:
- Check user has active account
- Verify role has permission
- Check for conflicting overrides
- Review audit log for permission checks

**Customer isolation violations**:
- Ensure all queries filter by customer_id
- Check foreign key constraints
- Verify user belongs to correct customer

**License expiry**:
- Check customer expiry_date
- Verify customer status is ACTIVE
- Extend license if needed

### 19. Future Enhancements

- [ ] Multi-database support (database per tenant)
- [ ] Dynamic permission creation
- [ ] Permission templates/presets
- [ ] Bulk user operations
- [ ] Customer data export/import
- [ ] Advanced approval workflows
- [ ] Permission delegation
- [ ] Time-based permission grants
- [ ] IP whitelist per customer
- [ ] Custom role creation UI

### 20. References

- Database Schema: `/database/multitenant_schema.sql`
- Python Models: `/src/models/multitenant.py`
- Permission Service: `/src/services/permission_service.py`
- Customer Service: `/src/services/customer_service.py`
- Main README: `/README.md`

---

**Version**: 1.0.0
**Last Updated**: 2025-01-16
**Status**: Core Implementation Complete
**© 2025 Protei Corporation**
