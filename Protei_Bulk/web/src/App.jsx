import React from 'react';
import { BrowserRouter as Router, Routes, Route, Navigate } from 'react-router-dom';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';

// Layouts
import MainLayout from './components/Layout/MainLayout';

// Pages
import EnhancedDashboard from './pages/Dashboard/EnhancedDashboard';

// User Management
import UserAccounts from './pages/Users/UserAccounts';

// Campaign Management
import CreateCampaign from './pages/Campaigns/CreateCampaign';
import CampaignList from './pages/Campaigns/CampaignList';

// Templates & Contacts
import MessageTemplates from './pages/Templates/MessageTemplates';
import ContactLists from './pages/Contacts/ContactLists';

// Placeholder components for routes not yet implemented
const PlaceholderPage = ({ title }) => (
  <div>
    <h2>{title}</h2>
    <p>This page is under development.</p>
  </div>
);

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
  typography: {
    fontFamily: 'Roboto, Arial, sans-serif',
  },
});

function PrivateRoute({ children }) {
  // For demo purposes, always allow access
  // In production, check authentication status
  return children;
}

function App() {
  return (
    <ThemeProvider theme={theme}>
      <CssBaseline />
      <Router>
        <Routes>
          {/* Main Application Routes */}
          <Route
            element={
              <PrivateRoute>
                <MainLayout />
              </PrivateRoute>
            }
          >
            {/* Dashboard */}
            <Route path="/" element={<Navigate to="/dashboard" replace />} />
            <Route path="/dashboard" element={<EnhancedDashboard />} />

            {/* User Management */}
            <Route path="/users/accounts" element={<UserAccounts />} />
            <Route path="/users/resellers" element={<PlaceholderPage title="Reseller Accounts" />} />
            <Route path="/users/privileges" element={<PlaceholderPage title="Privileges Matrix" />} />
            <Route path="/users/audit" element={<PlaceholderPage title="Audit Trail" />} />

            {/* Account & Routing */}
            <Route path="/routing/smsc" element={<PlaceholderPage title="SMSC Connections" />} />
            <Route path="/routing/rules" element={<PlaceholderPage title="Routing Rules" />} />
            <Route path="/routing/senders" element={<PlaceholderPage title="Sender ID Management" />} />
            <Route path="/routing/hours" element={<PlaceholderPage title="Working Hours" />} />

            {/* Campaign Management */}
            <Route path="/campaigns/create" element={<CreateCampaign />} />
            <Route path="/campaigns/list" element={<CampaignList />} />
            <Route path="/campaigns/approval" element={<PlaceholderPage title="Approval Queue" />} />
            <Route path="/campaigns/monitor" element={<PlaceholderPage title="Campaign Monitoring" />} />

            {/* Message Templates */}
            <Route path="/templates" element={<MessageTemplates />} />

            {/* Contact & Profiles */}
            <Route path="/contacts/lists" element={<ContactLists />} />
            <Route path="/contacts/hidden" element={<PlaceholderPage title="Hidden Lists" />} />
            <Route path="/contacts/profiling" element={<PlaceholderPage title="Subscriber Profiling" />} />
            <Route path="/contacts/groups" element={<PlaceholderPage title="Profile Groups" />} />

            {/* API & Integration */}
            <Route path="/api/docs" element={<PlaceholderPage title="API Documentation" />} />
            <Route path="/api/keys" element={<PlaceholderPage title="API Key Management" />} />
            <Route path="/api/smpp" element={<PlaceholderPage title="SMPP Accounts" />} />

            {/* Reports & Analytics */}
            <Route path="/reports/realtime" element={<PlaceholderPage title="Real-Time Reports" />} />
            <Route path="/reports/category" element={<PlaceholderPage title="Category Reports" />} />
            <Route path="/reports/profile" element={<PlaceholderPage title="Profile Reports" />} />
            <Route path="/reports/consumption" element={<PlaceholderPage title="Consumption Reports" />} />
            <Route path="/reports/alerts" element={<PlaceholderPage title="Alert Reports" />} />

            {/* Monitoring & Health */}
            <Route path="/monitoring/system" element={<PlaceholderPage title="System Dashboard" />} />
            <Route path="/monitoring/processes" element={<PlaceholderPage title="Process Monitor" />} />
            <Route path="/monitoring/logs" element={<PlaceholderPage title="Log Viewer" />} />
            <Route path="/monitoring/alerts" element={<PlaceholderPage title="Alerting System" />} />

            {/* Simulation & Testing */}
            <Route path="/testing/simulator" element={<PlaceholderPage title="SMS Simulator" />} />
            <Route path="/testing/load" element={<PlaceholderPage title="Load Tester" />} />

            {/* Configuration */}
            <Route path="/config/system" element={<PlaceholderPage title="System Parameters" />} />
            <Route path="/config/scheduler" element={<PlaceholderPage title="Scheduler Settings" />} />
            <Route path="/config/backup" element={<PlaceholderPage title="Backup & Restore" />} />
            <Route path="/config/language" element={<PlaceholderPage title="Language Settings" />} />
            <Route path="/config/theme" element={<PlaceholderPage title="Theme Customization" />} />

            {/* Security & Audit */}
            <Route path="/security/audit" element={<PlaceholderPage title="Audit Logs" />} />
            <Route path="/security/login" element={<PlaceholderPage title="Login Logs" />} />
            <Route path="/security/blocked" element={<PlaceholderPage title="Blocked Users" />} />
            <Route path="/security/maker-checker" element={<PlaceholderPage title="Maker-Checker History" />} />

            {/* Profile & Settings */}
            <Route path="/profile" element={<PlaceholderPage title="My Profile" />} />
            <Route path="/settings" element={<PlaceholderPage title="Settings" />} />

            {/* 404 */}
            <Route path="*" element={<PlaceholderPage title="Page Not Found" />} />
          </Route>

          {/* Login Route (if needed) */}
          <Route path="/login" element={<PlaceholderPage title="Login" />} />
        </Routes>
      </Router>
    </ThemeProvider>
  );
}

export default App;
