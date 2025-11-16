import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Button,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  IconButton,
  Chip,
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Switch,
  FormControlLabel,
  Grid
} from '@mui/material';
import {
  Add as AddIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  Lock as LockIcon,
  Key as KeyIcon
} from '@mui/icons-material';

/**
 * User Accounts Management Page
 * Add/Edit/Delete users, Assign roles, Set working hours and TPS
 */
function UserAccounts() {
  const [users, setUsers] = useState([]);
  const [openDialog, setOpenDialog] = useState(false);
  const [selectedUser, setSelectedUser] = useState(null);
  const [formData, setFormData] = useState({
    username: '',
    email: '',
    role: 'USER',
    account_type: 'PREPAID',
    tps_limit: 100,
    max_tps: 500,
    working_hours_start: '00:00',
    working_hours_end: '23:59',
    two_factor_enabled: false,
    is_active: true,
    assigned_smsc: '',
  });

  useEffect(() => {
    fetchUsers();
  }, []);

  const fetchUsers = async () => {
    // Mock data - replace with actual API call
    setUsers([
      {
        id: 1,
        username: 'admin',
        email: 'admin@protei.com',
        role: 'ADMIN',
        account_type: 'POSTPAID',
        tps_limit: 1000,
        two_factor_enabled: true,
        is_active: true,
        last_login: '2025-01-15 10:30:00'
      },
      {
        id: 2,
        username: 'reseller1',
        email: 'reseller@example.com',
        role: 'RESELLER',
        account_type: 'PREPAID',
        tps_limit: 500,
        two_factor_enabled: false,
        is_active: true,
        last_login: '2025-01-15 09:15:00'
      },
      {
        id: 3,
        username: 'user1',
        email: 'user@example.com',
        role: 'USER',
        account_type: 'PREPAID',
        tps_limit: 100,
        two_factor_enabled: true,
        is_active: true,
        last_login: '2025-01-14 16:45:00'
      },
    ]);
  };

  const handleOpenDialog = (user = null) => {
    if (user) {
      setSelectedUser(user);
      setFormData(user);
    } else {
      setSelectedUser(null);
      setFormData({
        username: '',
        email: '',
        role: 'USER',
        account_type: 'PREPAID',
        tps_limit: 100,
        max_tps: 500,
        working_hours_start: '00:00',
        working_hours_end: '23:59',
        two_factor_enabled: false,
        is_active: true,
        assigned_smsc: '',
      });
    }
    setOpenDialog(true);
  };

  const handleCloseDialog = () => {
    setOpenDialog(false);
    setSelectedUser(null);
  };

  const handleSave = async () => {
    // Save user logic here
    console.log('Saving user:', formData);
    handleCloseDialog();
    fetchUsers();
  };

  const handleDelete = async (userId) => {
    if (window.confirm('Are you sure you want to delete this user?')) {
      // Delete user logic here
      console.log('Deleting user:', userId);
      fetchUsers();
    }
  };

  const getRoleColor = (role) => {
    const colors = {
      ADMIN: 'error',
      RESELLER: 'warning',
      USER: 'primary',
      APPROVER: 'info',
    };
    return colors[role] || 'default';
  };

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5">User Accounts</Typography>
        <Button
          variant="contained"
          startIcon={<AddIcon />}
          onClick={() => handleOpenDialog()}
        >
          Add User
        </Button>
      </Box>

      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Username</TableCell>
              <TableCell>Email</TableCell>
              <TableCell>Role</TableCell>
              <TableCell>Account Type</TableCell>
              <TableCell>TPS Limit</TableCell>
              <TableCell>2FA</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Last Login</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {users.map((user) => (
              <TableRow key={user.id}>
                <TableCell>{user.username}</TableCell>
                <TableCell>{user.email}</TableCell>
                <TableCell>
                  <Chip label={user.role} color={getRoleColor(user.role)} size="small" />
                </TableCell>
                <TableCell>{user.account_type}</TableCell>
                <TableCell>{user.tps_limit}</TableCell>
                <TableCell>
                  {user.two_factor_enabled ? (
                    <Chip label="Enabled" color="success" size="small" icon={<LockIcon />} />
                  ) : (
                    <Chip label="Disabled" size="small" />
                  )}
                </TableCell>
                <TableCell>
                  <Chip
                    label={user.is_active ? 'Active' : 'Inactive'}
                    color={user.is_active ? 'success' : 'default'}
                    size="small"
                  />
                </TableCell>
                <TableCell>{user.last_login}</TableCell>
                <TableCell align="right">
                  <IconButton size="small" onClick={() => handleOpenDialog(user)}>
                    <EditIcon />
                  </IconButton>
                  <IconButton size="small" color="error" onClick={() => handleDelete(user.id)}>
                    <DeleteIcon />
                  </IconButton>
                  <IconButton size="small">
                    <KeyIcon />
                  </IconButton>
                </TableCell>
              </TableRow>
            ))}
          </TableBody>
        </Table>
      </TableContainer>

      {/* Add/Edit User Dialog */}
      <Dialog open={openDialog} onClose={handleCloseDialog} maxWidth="md" fullWidth>
        <DialogTitle>
          {selectedUser ? 'Edit User' : 'Add New User'}
        </DialogTitle>
        <DialogContent>
          <Grid container spacing={2} sx={{ mt: 1 }}>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Username"
                value={formData.username}
                onChange={(e) => setFormData({ ...formData, username: e.target.value })}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Email"
                type="email"
                value={formData.email}
                onChange={(e) => setFormData({ ...formData, email: e.target.value })}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth>
                <InputLabel>Role</InputLabel>
                <Select
                  value={formData.role}
                  label="Role"
                  onChange={(e) => setFormData({ ...formData, role: e.target.value })}
                >
                  <MenuItem value="ADMIN">Admin</MenuItem>
                  <MenuItem value="RESELLER">Reseller</MenuItem>
                  <MenuItem value="USER">User</MenuItem>
                  <MenuItem value="APPROVER">Approver</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6}>
              <FormControl fullWidth>
                <InputLabel>Account Type</InputLabel>
                <Select
                  value={formData.account_type}
                  label="Account Type"
                  onChange={(e) => setFormData({ ...formData, account_type: e.target.value })}
                >
                  <MenuItem value="PREPAID">Prepaid</MenuItem>
                  <MenuItem value="POSTPAID">Postpaid</MenuItem>
                </Select>
              </FormControl>
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="TPS Limit"
                type="number"
                value={formData.tps_limit}
                onChange={(e) => setFormData({ ...formData, tps_limit: parseInt(e.target.value) })}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Max TPS"
                type="number"
                value={formData.max_tps}
                onChange={(e) => setFormData({ ...formData, max_tps: parseInt(e.target.value) })}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Working Hours Start"
                type="time"
                value={formData.working_hours_start}
                onChange={(e) => setFormData({ ...formData, working_hours_start: e.target.value })}
                InputLabelProps={{ shrink: true }}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Working Hours End"
                type="time"
                value={formData.working_hours_end}
                onChange={(e) => setFormData({ ...formData, working_hours_end: e.target.value })}
                InputLabelProps={{ shrink: true }}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <TextField
                fullWidth
                label="Assigned SMSC"
                value={formData.assigned_smsc}
                onChange={(e) => setFormData({ ...formData, assigned_smsc: e.target.value })}
                helperText="Leave empty for auto-routing"
              />
            </Grid>
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.two_factor_enabled}
                    onChange={(e) => setFormData({ ...formData, two_factor_enabled: e.target.checked })}
                  />
                }
                label="Enable Two-Factor Authentication"
              />
            </Grid>
            <Grid item xs={12}>
              <FormControlLabel
                control={
                  <Switch
                    checked={formData.is_active}
                    onChange={(e) => setFormData({ ...formData, is_active: e.target.checked })}
                  />
                }
                label="Active Account"
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions>
          <Button onClick={handleCloseDialog}>Cancel</Button>
          <Button variant="contained" onClick={handleSave}>
            {selectedUser ? 'Update' : 'Create'}
          </Button>
        </DialogActions>
      </Dialog>
    </Box>
  );
}

export default UserAccounts;
