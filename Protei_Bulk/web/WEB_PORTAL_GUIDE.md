# Protei_Bulk Web Portal - Complete Guide

## Overview

The Protei_Bulk Web Portal is a comprehensive enterprise messaging management system built with React 18 and Material-UI. It provides complete functionality for managing users, campaigns, templates, contacts, and system monitoring.

## Features Implemented

### ✅ A. Dashboard
- **Real-time Statistics Panel**
  - TPS (Transactions Per Second) per channel/SMSC
  - Queue status monitoring
  - System uptime and load tracking
  - Account balance and message counters
- **Alerts Widget**
  - System alerts (CPU, disk, faults)
  - Prepaid balance warnings
  - Campaign delivery failure notifications
- **Graphical KPIs**
  - Throughput graphs per user/SMSC
  - Success rate and DLR statistics
  - Top sender IDs and routes visualization
- **Quick Actions**
  - Send New Campaign
  - Create Template
  - Add Contact List
  - Monitor Campaigns

### ✅ B. User Management
- **User Accounts Page** (`/users/accounts`)
  - Add/Edit/Delete users
  - Assign roles (Admin, Reseller, User, Approver)
  - Set working hours and max TPS limits
  - Assign specific SMSC or route
  - Enable/disable two-factor authentication
  - Account activation/deactivation

### ✅ C. Campaign Management
- **Create Campaign** (`/campaigns/create`)
  - Multi-step wizard interface
  - Channel selection: SMS, USSD, Email, Push, WhatsApp, Telegram
  - Recipient selection:
    - Upload file (CSV/TXT/Excel)
    - Select from Contact List
    - Select Profile Group
    - Manual MSISDN entry
  - Template or custom message composition
  - Sender ID definition
  - Schedule configuration (immediate, scheduled, recurring)
  - Priority setting
  - Max message/day restriction

- **Campaign List** (`/campaigns/list`)
  - View all campaigns by status (Draft, Scheduled, Running, Paused, Completed, Failed)
  - Real-time progress tracking
  - Campaign actions: View, Edit, Pause, Resume, Stop, Cancel, Duplicate
  - Delivery statistics and progress bars
  - Filter and search capabilities

### ✅ D. Message Templates
- **Template Manager** (`/templates`)
  - Create/Edit/Delete templates
  - Multi-language support (EN/AR)
  - Variables support (e.g., %NAME%, %CODE%, %BALANCE%)
  - Categorization (OTP, Marketing, Alert, Info)
  - Template visibility control (Public, User, Reseller)

### ✅ E. Contact & Profile Management
- **Contact Lists** (`/contacts/lists`)
  - Create/Import/Export contact groups
  - Upload Excel/CSV/TXT files (supports 1M+ records)
  - Grouping by tag or segment
  - Contact count tracking
  - Last updated timestamps

### 🚧 F. Additional Sections (Placeholder Routes)
All navigation menu items are functional with placeholder pages ready for implementation:
- Reseller Accounts
- Privileges Matrix
- Audit Trail
- SMSC Connections
- Routing Rules
- Sender ID Management
- Working Hours
- Approval Queue
- Campaign Monitoring
- Hidden Lists
- Subscriber Profiling
- Profile Groups
- API Documentation
- API Key Management
- SMPP Accounts
- Real-Time Reports
- Category Reports
- Profile Reports
- Consumption Reports
- Alert Reports
- System Dashboard
- Process Monitor
- Log Viewer
- Alerting System
- SMS Simulator
- Load Tester
- System Parameters
- Scheduler Settings
- Backup & Restore
- Language Settings
- Theme Customization
- Audit Logs
- Login Logs
- Blocked Users
- Maker-Checker History

## Architecture

### Directory Structure

```
web/
├── src/
│   ├── components/
│   │   └── Layout/
│   │       └── MainLayout.jsx       # Main navigation layout
│   ├── pages/
│   │   ├── Dashboard/
│   │   │   └── EnhancedDashboard.jsx  # Complete dashboard
│   │   ├── Users/
│   │   │   └── UserAccounts.jsx       # User management
│   │   ├── Campaigns/
│   │   │   ├── CreateCampaign.jsx     # Campaign creation wizard
│   │   │   └── CampaignList.jsx       # Campaign listing
│   │   ├── Templates/
│   │   │   └── MessageTemplates.jsx   # Template management
│   │   └── Contacts/
│   │       └── ContactLists.jsx       # Contact list management
│   ├── services/
│   │   ├── api.js                     # Base API service
│   │   └── analyticsAPI.js            # Analytics API
│   ├── store/
│   │   └── authStore.js               # Auth state management
│   └── App.jsx                        # Main app with routing
├── package.json
└── README.md
```

### Navigation Structure

The portal uses a collapsible sidebar navigation with the following structure:

1. **Dashboard** - Real-time overview
2. **User Management** - Users, resellers, privileges, audit
3. **Account & Routing** - SMSC, routing rules, sender IDs
4. **Campaign Management** - Create, list, approval, monitoring
5. **Message Templates** - Template library
6. **Contact & Profiles** - Lists, hidden lists, profiling, groups
7. **API & Integration** - Documentation, keys, SMPP
8. **Reports & Analytics** - Various report types
9. **Monitoring & Health** - System monitoring
10. **Simulation & Testing** - Simulators and load testers
11. **Configuration** - System settings
12. **Security & Audit** - Security logs and history

## Installation

### Prerequisites
- Node.js 16+ and npm
- Protei_Bulk backend running on http://localhost:8080

### Setup

```bash
cd web/

# Install dependencies
npm install

# Configure API endpoint
echo "REACT_APP_API_URL=http://localhost:8080/api/v1" > .env
echo "REACT_APP_WS_URL=ws://localhost:8080" >> .env

# Start development server
npm start
```

The application will open at http://localhost:3000

### Build for Production

```bash
npm run build
```

The optimized production build will be in the `build/` directory.

## Usage Guide

### Dashboard
- Navigate to `/dashboard` to view real-time statistics
- Monitor TPS, queue status, and system health
- View recent alerts and notifications
- Use quick actions for common tasks

### Creating a Campaign
1. Click "Create Campaign" or navigate to `/campaigns/create`
2. **Step 1**: Select channel (SMS, Email, etc.) and sender ID
3. **Step 2**: Choose recipient source (file upload, contact list, profile group, or manual entry)
4. **Step 3**: Compose message (use template or write custom)
5. **Step 4**: Set schedule (immediate, scheduled, or recurring)
6. **Step 5**: Review and submit

### Managing Users
1. Navigate to `/users/accounts`
2. Click "Add User" to create a new user
3. Fill in user details:
   - Username and email
   - Role (Admin, Reseller, User, Approver)
   - Account type (Prepaid/Postpaid)
   - TPS limits
   - Working hours
   - 2FA settings
4. Click "Create" to save

### Creating Templates
1. Navigate to `/templates`
2. Click "Create Template"
3. Enter template name and content
4. Use variables like %NAME%, %CODE% for dynamic content
5. Select category and language
6. Set visibility level
7. Click "Create"

### Managing Contact Lists
1. Navigate to `/contacts/lists`
2. Click "Create List" or "Import Contacts"
3. For import: Upload CSV/TXT/Excel file
4. Add tags for organization
5. View, edit, or export lists as needed

## API Integration

The web portal connects to the backend API automatically. All API calls use:

```javascript
import { analyticsAPI } from './services/analyticsAPI';

// Get real-time metrics
const metrics = await analyticsAPI.getRealtimeMessageMetrics(60);

// Get dashboard summary
const summary = await analyticsAPI.dashboardSummary();

// Get campaign metrics
const campaign = await analyticsAPI.getCampaignMetrics(campaignId);
```

## Real-time Updates

The dashboard refreshes automatically every 5 seconds to show:
- Current TPS
- Message statistics
- System resource usage
- Active campaigns
- Recent alerts

## Customization

### Theme
The application uses Material-UI theming. Customize in `App.jsx`:

```javascript
const theme = createTheme({
  palette: {
    mode: 'light',
    primary: {
      main: '#1976d2',
    },
    secondary: {
      main: '#dc004e',
    },
  },
});
```

### Adding New Pages
1. Create page component in `src/pages/`
2. Add route in `App.jsx`
3. Add menu item in `MainLayout.jsx`

## Development

### Code Structure
- **Components**: Reusable UI components
- **Pages**: Full page components
- **Services**: API integration layer
- **Store**: State management with Zustand

### Best Practices
- Use functional components with hooks
- Implement proper error handling
- Add loading states for async operations
- Follow Material-UI design patterns
- Keep components modular and reusable

## Troubleshooting

### API Connection Issues
```bash
# Check if backend is running
curl http://localhost:8080/api/v1/health

# Verify CORS settings in backend config/api.conf
enable_cors = true
cors_origins = http://localhost:3000
```

### Port Already in Use
```bash
# Use a different port
PORT=3001 npm start
```

### Dependencies Issues
```bash
# Clear cache and reinstall
rm -rf node_modules package-lock.json
npm install
```

## Performance

- React 18 with automatic batching
- Lazy loading for code splitting
- Optimized re-renders with React.memo
- Efficient state management with Zustand
- Real-time updates every 5 seconds (configurable)

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## Security

- Authentication required for all routes
- JWT token-based session management
- XSS protection
- CSRF protection
- Secure API communication

## Future Enhancements

- [ ] Dark mode theme
- [ ] Multi-language support (i18n)
- [ ] Export reports to PDF/Excel
- [ ] Advanced filtering and search
- [ ] Real-time notifications via WebSocket
- [ ] Drag-and-drop campaign builder
- [ ] Advanced analytics dashboard
- [ ] Mobile responsive improvements

## Support

For issues or questions:
- Check the main README.md
- View API documentation at http://localhost:8080/api/docs
- Contact the development team

---

**Protei_Bulk Web Portal** - Enterprise Messaging Management
Version 1.0.0 | © 2025 Protei Corporation
