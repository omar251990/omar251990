import React, { useState } from 'react';
import { Outlet, useNavigate, useLocation } from 'react-router-dom';
import {
  Box,
  Drawer,
  AppBar,
  Toolbar,
  List,
  Typography,
  Divider,
  IconButton,
  ListItem,
  ListItemButton,
  ListItemIcon,
  ListItemText,
  Collapse,
  Avatar,
  Menu,
  MenuItem,
  Badge,
  Tooltip
} from '@mui/material';
import {
  Menu as MenuIcon,
  Dashboard as DashboardIcon,
  People as PeopleIcon,
  Campaign as CampaignIcon,
  Message as MessageIcon,
  Contacts as ContactsIcon,
  Description as DescriptionIcon,
  Assessment as AssessmentIcon,
  Settings as SettingsIcon,
  Security as SecurityIcon,
  Notifications as NotificationsIcon,
  AccountCircle,
  Logout,
  ExpandLess,
  ExpandMore,
  Api as ApiIcon,
  Router as RouterIcon,
  Build as BuildIcon,
  BugReport as BugReportIcon,
  MonitorHeart as MonitorIcon
} from '@mui/icons-material';

const drawerWidth = 280;

/**
 * Main Layout Component
 * Provides navigation structure for the entire application
 */
function MainLayout() {
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileOpen, setMobileOpen] = useState(false);
  const [openMenus, setOpenMenus] = useState({});
  const [anchorEl, setAnchorEl] = useState(null);
  const [notifAnchorEl, setNotifAnchorEl] = useState(null);

  // Navigation menu structure
  const menuItems = [
    {
      title: 'Dashboard',
      icon: <DashboardIcon />,
      path: '/dashboard',
    },
    {
      title: 'User Management',
      icon: <PeopleIcon />,
      children: [
        { title: 'User Accounts', path: '/users/accounts' },
        { title: 'Reseller Accounts', path: '/users/resellers' },
        { title: 'Privileges Matrix', path: '/users/privileges' },
        { title: 'Audit Trail', path: '/users/audit' },
      ],
    },
    {
      title: 'Account & Routing',
      icon: <RouterIcon />,
      children: [
        { title: 'SMSC Connections', path: '/routing/smsc' },
        { title: 'Routing Rules', path: '/routing/rules' },
        { title: 'Sender ID Management', path: '/routing/senders' },
        { title: 'Working Hours', path: '/routing/hours' },
      ],
    },
    {
      title: 'Campaign Management',
      icon: <CampaignIcon />,
      children: [
        { title: 'Create Campaign', path: '/campaigns/create' },
        { title: 'Campaign List', path: '/campaigns/list' },
        { title: 'Approval Queue', path: '/campaigns/approval' },
        { title: 'Campaign Monitoring', path: '/campaigns/monitor' },
      ],
    },
    {
      title: 'Message Templates',
      icon: <MessageIcon />,
      path: '/templates',
    },
    {
      title: 'Contact & Profiles',
      icon: <ContactsIcon />,
      children: [
        { title: 'Contact Lists', path: '/contacts/lists' },
        { title: 'Hidden Lists', path: '/contacts/hidden' },
        { title: 'Subscriber Profiling', path: '/contacts/profiling' },
        { title: 'Profile Groups', path: '/contacts/groups' },
      ],
    },
    {
      title: 'API & Integration',
      icon: <ApiIcon />,
      children: [
        { title: 'API Documentation', path: '/api/docs' },
        { title: 'API Key Management', path: '/api/keys' },
        { title: 'SMPP Accounts', path: '/api/smpp' },
      ],
    },
    {
      title: 'Reports & Analytics',
      icon: <AssessmentIcon />,
      children: [
        { title: 'Real-Time Reports', path: '/reports/realtime' },
        { title: 'Category Reports', path: '/reports/category' },
        { title: 'Profile Reports', path: '/reports/profile' },
        { title: 'Consumption Reports', path: '/reports/consumption' },
        { title: 'Alert Reports', path: '/reports/alerts' },
      ],
    },
    {
      title: 'Monitoring & Health',
      icon: <MonitorIcon />,
      children: [
        { title: 'System Dashboard', path: '/monitoring/system' },
        { title: 'Process Monitor', path: '/monitoring/processes' },
        { title: 'Log Viewer', path: '/monitoring/logs' },
        { title: 'Alerting System', path: '/monitoring/alerts' },
      ],
    },
    {
      title: 'Simulation & Testing',
      icon: <BugReportIcon />,
      children: [
        { title: 'SMS Simulator', path: '/testing/simulator' },
        { title: 'Load Tester', path: '/testing/load' },
      ],
    },
    {
      title: 'Configuration',
      icon: <SettingsIcon />,
      children: [
        { title: 'System Parameters', path: '/config/system' },
        { title: 'Scheduler Settings', path: '/config/scheduler' },
        { title: 'Backup & Restore', path: '/config/backup' },
        { title: 'Language Settings', path: '/config/language' },
        { title: 'Theme Customization', path: '/config/theme' },
      ],
    },
    {
      title: 'Security & Audit',
      icon: <SecurityIcon />,
      children: [
        { title: 'Audit Logs', path: '/security/audit' },
        { title: 'Login Logs', path: '/security/login' },
        { title: 'Blocked Users', path: '/security/blocked' },
        { title: 'Maker-Checker History', path: '/security/maker-checker' },
      ],
    },
  ];

  const handleDrawerToggle = () => {
    setMobileOpen(!mobileOpen);
  };

  const handleMenuClick = (item) => {
    if (item.children) {
      setOpenMenus({
        ...openMenus,
        [item.title]: !openMenus[item.title],
      });
    } else {
      navigate(item.path);
      setMobileOpen(false);
    }
  };

  const handleProfileMenuOpen = (event) => {
    setAnchorEl(event.currentTarget);
  };

  const handleProfileMenuClose = () => {
    setAnchorEl(null);
  };

  const handleNotificationOpen = (event) => {
    setNotifAnchorEl(event.currentTarget);
  };

  const handleNotificationClose = () => {
    setNotifAnchorEl(null);
  };

  const handleLogout = () => {
    // Clear auth and redirect to login
    localStorage.removeItem('auth-storage');
    navigate('/login');
  };

  const isActive = (path) => {
    return location.pathname === path;
  };

  const drawer = (
    <div>
      <Toolbar sx={{ backgroundColor: 'primary.main', color: 'white' }}>
        <Typography variant="h6" noWrap component="div">
          Protei_Bulk
        </Typography>
      </Toolbar>
      <Divider />
      <List>
        {menuItems.map((item) => (
          <div key={item.title}>
            <ListItem disablePadding>
              <ListItemButton
                onClick={() => handleMenuClick(item)}
                selected={!item.children && isActive(item.path)}
              >
                <ListItemIcon>{item.icon}</ListItemIcon>
                <ListItemText primary={item.title} />
                {item.children && (openMenus[item.title] ? <ExpandLess /> : <ExpandMore />)}
              </ListItemButton>
            </ListItem>
            {item.children && (
              <Collapse in={openMenus[item.title]} timeout="auto" unmountOnExit>
                <List component="div" disablePadding>
                  {item.children.map((child) => (
                    <ListItemButton
                      key={child.path}
                      sx={{ pl: 4 }}
                      onClick={() => {
                        navigate(child.path);
                        setMobileOpen(false);
                      }}
                      selected={isActive(child.path)}
                    >
                      <ListItemText primary={child.title} />
                    </ListItemButton>
                  ))}
                </List>
              </Collapse>
            )}
          </div>
        ))}
      </List>
    </div>
  );

  return (
    <Box sx={{ display: 'flex' }}>
      {/* AppBar */}
      <AppBar
        position="fixed"
        sx={{
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          ml: { sm: `${drawerWidth}px` },
        }}
      >
        <Toolbar>
          <IconButton
            color="inherit"
            aria-label="open drawer"
            edge="start"
            onClick={handleDrawerToggle}
            sx={{ mr: 2, display: { sm: 'none' } }}
          >
            <MenuIcon />
          </IconButton>
          <Typography variant="h6" noWrap component="div" sx={{ flexGrow: 1 }}>
            Enterprise Bulk Messaging Platform
          </Typography>

          {/* Notifications */}
          <Tooltip title="Notifications">
            <IconButton color="inherit" onClick={handleNotificationOpen}>
              <Badge badgeContent={3} color="error">
                <NotificationsIcon />
              </Badge>
            </IconButton>
          </Tooltip>

          {/* Profile Menu */}
          <Tooltip title="Account">
            <IconButton onClick={handleProfileMenuOpen} sx={{ ml: 1 }}>
              <Avatar sx={{ width: 32, height: 32, bgcolor: 'secondary.main' }}>
                A
              </Avatar>
            </IconButton>
          </Tooltip>
        </Toolbar>
      </AppBar>

      {/* Profile Menu */}
      <Menu
        anchorEl={anchorEl}
        open={Boolean(anchorEl)}
        onClose={handleProfileMenuClose}
        onClick={handleProfileMenuClose}
      >
        <MenuItem onClick={() => navigate('/profile')}>
          <ListItemIcon>
            <AccountCircle fontSize="small" />
          </ListItemIcon>
          My Profile
        </MenuItem>
        <MenuItem onClick={() => navigate('/settings')}>
          <ListItemIcon>
            <SettingsIcon fontSize="small" />
          </ListItemIcon>
          Settings
        </MenuItem>
        <Divider />
        <MenuItem onClick={handleLogout}>
          <ListItemIcon>
            <Logout fontSize="small" />
          </ListItemIcon>
          Logout
        </MenuItem>
      </Menu>

      {/* Notifications Menu */}
      <Menu
        anchorEl={notifAnchorEl}
        open={Boolean(notifAnchorEl)}
        onClose={handleNotificationClose}
        PaperProps={{
          style: {
            maxHeight: 400,
            width: '350px',
          },
        }}
      >
        <MenuItem>
          <ListItemText
            primary="Low Balance Alert"
            secondary="Account balance below threshold"
          />
        </MenuItem>
        <MenuItem>
          <ListItemText
            primary="Campaign Completed"
            secondary="Spring Promo campaign finished"
          />
        </MenuItem>
        <MenuItem>
          <ListItemText
            primary="SMSC Connection Issue"
            secondary="SMSC-01 connection failed"
          />
        </MenuItem>
      </Menu>

      {/* Drawer */}
      <Box
        component="nav"
        sx={{ width: { sm: drawerWidth }, flexShrink: { sm: 0 } }}
      >
        {/* Mobile drawer */}
        <Drawer
          variant="temporary"
          open={mobileOpen}
          onClose={handleDrawerToggle}
          ModalProps={{
            keepMounted: true, // Better mobile performance
          }}
          sx={{
            display: { xs: 'block', sm: 'none' },
            '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
          }}
        >
          {drawer}
        </Drawer>
        {/* Desktop drawer */}
        <Drawer
          variant="permanent"
          sx={{
            display: { xs: 'none', sm: 'block' },
            '& .MuiDrawer-paper': { boxSizing: 'border-box', width: drawerWidth },
          }}
          open
        >
          {drawer}
        </Drawer>
      </Box>

      {/* Main content */}
      <Box
        component="main"
        sx={{
          flexGrow: 1,
          p: 3,
          width: { sm: `calc(100% - ${drawerWidth}px)` },
          minHeight: '100vh',
          backgroundColor: 'background.default',
        }}
      >
        <Toolbar />
        <Outlet />
      </Box>
    </Box>
  );
}

export default MainLayout;
