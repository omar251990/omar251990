import React, { useState, useEffect } from 'react';
import {
  Box,
  Paper,
  Typography,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  Chip,
  IconButton,
  Button,
  TextField,
  Select,
  MenuItem,
  FormControl,
  InputLabel,
  Grid,
  LinearProgress
} from '@mui/material';
import {
  Visibility as ViewIcon,
  Edit as EditIcon,
  Delete as DeleteIcon,
  FileCopy as CopyIcon,
  Pause as PauseIcon,
  PlayArrow as PlayIcon,
  Stop as StopIcon
} from '@mui/icons-material';

/**
 * Campaign List Page
 * View, manage, and monitor all campaigns
 */
function CampaignList() {
  const [campaigns, setCampaigns] = useState([]);
  const [filterStatus, setFilterStatus] = useState('ALL');
  const [searchQuery, setSearchQuery] = useState('');

  useEffect(() => {
    fetchCampaigns();
  }, []);

  const fetchCampaigns = async () => {
    // Mock data - replace with actual API call
    setCampaigns([
      {
        id: 1,
        name: 'Spring Promotion 2025',
        status: 'RUNNING',
        channel: 'SMS',
        total_recipients: 50000,
        processed: 35420,
        delivered: 34251,
        failed: 1169,
        pending: 14580,
        created_at: '2025-01-15 09:00:00',
        sender_id: 'ACME',
        priority: 'NORMAL'
      },
      {
        id: 2,
        name: 'OTP Verification Batch',
        status: 'COMPLETED',
        channel: 'SMS',
        total_recipients: 12000,
        processed: 12000,
        delivered: 11876,
        failed: 124,
        pending: 0,
        created_at: '2025-01-14 18:30:00',
        sender_id: 'OTP',
        priority: 'HIGH'
      },
      {
        id: 3,
        name: 'Customer Feedback Survey',
        status: 'SCHEDULED',
        channel: 'EMAIL',
        total_recipients: 8500,
        processed: 0,
        delivered: 0,
        failed: 0,
        pending: 8500,
        created_at: '2025-01-15 11:00:00',
        scheduled_for: '2025-01-16 10:00:00',
        sender_id: 'SURVEY',
        priority: 'LOW'
      },
      {
        id: 4,
        name: 'Payment Reminders',
        status: 'PAUSED',
        channel: 'SMS',
        total_recipients: 25000,
        processed: 10250,
        delivered: 9875,
        failed: 375,
        pending: 14750,
        created_at: '2025-01-14 14:00:00',
        sender_id: 'PAYMENTS',
        priority: 'NORMAL'
      },
    ]);
  };

  const getStatusColor = (status) => {
    const colors = {
      DRAFT: 'default',
      SCHEDULED: 'info',
      RUNNING: 'primary',
      PAUSED: 'warning',
      COMPLETED: 'success',
      FAILED: 'error',
      CANCELLED: 'default'
    };
    return colors[status] || 'default';
  };

  const getPriorityColor = (priority) => {
    const colors = {
      CRITICAL: 'error',
      HIGH: 'warning',
      NORMAL: 'info',
      LOW: 'default'
    };
    return colors[priority] || 'default';
  };

  const handleAction = (action, campaignId) => {
    console.log(`${action} campaign:`, campaignId);
    // Implement actions: view, edit, delete, pause, resume, stop, duplicate
    fetchCampaigns();
  };

  const filteredCampaigns = campaigns.filter(campaign => {
    const matchesStatus = filterStatus === 'ALL' || campaign.status === filterStatus;
    const matchesSearch = campaign.name.toLowerCase().includes(searchQuery.toLowerCase());
    return matchesStatus && matchesSearch;
  });

  return (
    <Box>
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Typography variant="h5">Campaign List</Typography>
        <Button
          variant="contained"
          onClick={() => window.location.href = '/campaigns/create'}
        >
          Create New Campaign
        </Button>
      </Box>

      {/* Filters */}
      <Paper sx={{ p: 2, mb: 3 }}>
        <Grid container spacing={2} alignItems="center">
          <Grid item xs={12} md={6}>
            <TextField
              fullWidth
              label="Search Campaigns"
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              size="small"
            />
          </Grid>
          <Grid item xs={12} md={3}>
            <FormControl fullWidth size="small">
              <InputLabel>Status Filter</InputLabel>
              <Select
                value={filterStatus}
                label="Status Filter"
                onChange={(e) => setFilterStatus(e.target.value)}
              >
                <MenuItem value="ALL">All Statuses</MenuItem>
                <MenuItem value="DRAFT">Draft</MenuItem>
                <MenuItem value="SCHEDULED">Scheduled</MenuItem>
                <MenuItem value="RUNNING">Running</MenuItem>
                <MenuItem value="PAUSED">Paused</MenuItem>
                <MenuItem value="COMPLETED">Completed</MenuItem>
                <MenuItem value="FAILED">Failed</MenuItem>
              </Select>
            </FormControl>
          </Grid>
          <Grid item xs={12} md={3}>
            <Typography variant="body2" color="textSecondary">
              Total: {filteredCampaigns.length} campaigns
            </Typography>
          </Grid>
        </Grid>
      </Paper>

      {/* Campaigns Table */}
      <TableContainer component={Paper}>
        <Table>
          <TableHead>
            <TableRow>
              <TableCell>Campaign Name</TableCell>
              <TableCell>Status</TableCell>
              <TableCell>Channel</TableCell>
              <TableCell>Progress</TableCell>
              <TableCell>Delivered</TableCell>
              <TableCell>Failed</TableCell>
              <TableCell>Sender ID</TableCell>
              <TableCell>Priority</TableCell>
              <TableCell>Created</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {filteredCampaigns.map((campaign) => {
              const progress = (campaign.processed / campaign.total_recipients) * 100;
              const deliveryRate = campaign.processed > 0
                ? (campaign.delivered / campaign.processed) * 100
                : 0;

              return (
                <TableRow key={campaign.id}>
                  <TableCell>
                    <Typography variant="body2" fontWeight="bold">
                      {campaign.name}
                    </Typography>
                    <Typography variant="caption" color="textSecondary">
                      {campaign.total_recipients.toLocaleString()} recipients
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Chip
                      label={campaign.status}
                      color={getStatusColor(campaign.status)}
                      size="small"
                    />
                  </TableCell>
                  <TableCell>{campaign.channel}</TableCell>
                  <TableCell sx={{ minWidth: 150 }}>
                    <Box>
                      <Box display="flex" justifyContent="space-between" mb={0.5}>
                        <Typography variant="caption">
                          {campaign.processed.toLocaleString()} / {campaign.total_recipients.toLocaleString()}
                        </Typography>
                        <Typography variant="caption">
                          {progress.toFixed(1)}%
                        </Typography>
                      </Box>
                      <LinearProgress
                        variant="determinate"
                        value={progress}
                        color={campaign.status === 'RUNNING' ? 'primary' : 'inherit'}
                      />
                    </Box>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="success.main">
                      {campaign.delivered.toLocaleString()}
                    </Typography>
                    <Typography variant="caption" color="textSecondary">
                      {deliveryRate.toFixed(1)}%
                    </Typography>
                  </TableCell>
                  <TableCell>
                    <Typography variant="body2" color="error.main">
                      {campaign.failed.toLocaleString()}
                    </Typography>
                  </TableCell>
                  <TableCell>{campaign.sender_id}</TableCell>
                  <TableCell>
                    <Chip
                      label={campaign.priority}
                      color={getPriorityColor(campaign.priority)}
                      size="small"
                      variant="outlined"
                    />
                  </TableCell>
                  <TableCell>
                    <Typography variant="caption">
                      {campaign.created_at}
                    </Typography>
                  </TableCell>
                  <TableCell align="right">
                    <IconButton
                      size="small"
                      onClick={() => handleAction('view', campaign.id)}
                      title="View Details"
                    >
                      <ViewIcon />
                    </IconButton>
                    {campaign.status === 'DRAFT' && (
                      <IconButton
                        size="small"
                        onClick={() => handleAction('edit', campaign.id)}
                        title="Edit"
                      >
                        <EditIcon />
                      </IconButton>
                    )}
                    {campaign.status === 'RUNNING' && (
                      <IconButton
                        size="small"
                        onClick={() => handleAction('pause', campaign.id)}
                        title="Pause"
                      >
                        <PauseIcon />
                      </IconButton>
                    )}
                    {campaign.status === 'PAUSED' && (
                      <IconButton
                        size="small"
                        color="primary"
                        onClick={() => handleAction('resume', campaign.id)}
                        title="Resume"
                      >
                        <PlayIcon />
                      </IconButton>
                    )}
                    {['RUNNING', 'PAUSED'].includes(campaign.status) && (
                      <IconButton
                        size="small"
                        color="error"
                        onClick={() => handleAction('stop', campaign.id)}
                        title="Stop"
                      >
                        <StopIcon />
                      </IconButton>
                    )}
                    <IconButton
                      size="small"
                      onClick={() => handleAction('duplicate', campaign.id)}
                      title="Duplicate"
                    >
                      <CopyIcon />
                    </IconButton>
                    {campaign.status === 'DRAFT' && (
                      <IconButton
                        size="small"
                        color="error"
                        onClick={() => handleAction('delete', campaign.id)}
                        title="Delete"
                      >
                        <DeleteIcon />
                      </IconButton>
                    )}
                  </TableCell>
                </TableRow>
              );
            })}
          </TableBody>
        </Table>
      </TableContainer>
    </Box>
  );
}

export default CampaignList;
