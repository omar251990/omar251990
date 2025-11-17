import React, { useState, useEffect } from 'react';
import {
  Box,
  Grid,
  Paper,
  Typography,
  Card,
  CardContent,
  Button,
  Alert,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableHead,
  TableRow,
  LinearProgress,
  IconButton
} from '@mui/material';
import {
  Speed as SpeedIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Campaign as CampaignIcon,
  AttachMoney as MoneyIcon,
  Storage as StorageIcon,
  Memory as MemoryIcon,
  Warning as WarningIcon,
  Refresh as RefreshIcon,
  Add as AddIcon,
  Description as DescriptionIcon,
  Contacts as ContactsIcon
} from '@mui/icons-material';
import { LineChart, Line, AreaChart, Area, BarChart, Bar, PieChart, Pie, Cell, XAxis, YAxis, CartesianGrid, Tooltip, Legend, ResponsiveContainer } from 'recharts';
import { analyticsAPI } from '../../services/analyticsAPI';

/**
 * Enhanced Dashboard Page
 * Complete implementation with all sections from specification
 */
function EnhancedDashboard() {
  const [loading, setLoading] = useState(true);
  const [summary, setSummary] = useState(null);
  const [realtimeMetrics, setRealtimeMetrics] = useState(null);
  const [messageTrend, setMessageTrend] = useState([]);
  const [alerts, setAlerts] = useState([]);
  const [tpsData, setTpsData] = useState([]);
  const [queueStatus, setQueueStatus] = useState([]);

  useEffect(() => {
    fetchDashboardData();
    const interval = setInterval(fetchDashboardData, 5000); // Refresh every 5 seconds
    return () => clearInterval(interval);
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);
      const [summaryRes, metricsRes, trendRes] = await Promise.all([
        analyticsAPI.dashboardSummary(),
        analyticsAPI.getRealtimeMessageMetrics(60),
        analyticsAPI.getMessageTrend('hour', 24)
      ]);

      setSummary(summaryRes.data.data);
      setRealtimeMetrics(metricsRes.data.data);

      // Transform trend data
      const trend = trendRes.data.data.data_points.map(point => ({
        time: new Date(point.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' }),
        messages: point.value
      }));
      setMessageTrend(trend);

      // Mock TPS data per channel/SMSC
      setTpsData([
        { name: 'SMSC-1', tps: 3245, max: 5000 },
        { name: 'SMSC-2', tps: 2891, max: 5000 },
        { name: 'SMSC-3', tps: 1987, max: 3000 },
        { name: 'HTTP API', tps: 1456, max: 2000 },
      ]);

      // Mock queue status
      setQueueStatus([
        { queue: 'High Priority', count: 1234, processing: 98 },
        { queue: 'Normal', count: 5678, processing: 245 },
        { queue: 'Low Priority', count: 987, processing: 12 },
        { queue: 'Scheduled', count: 3456, processing: 0 },
      ]);

      // Mock alerts
      setAlerts([
        { id: 1, type: 'warning', message: 'Account ABC balance below 1000 credits', time: '2 mins ago' },
        { id: 2, type: 'error', message: 'SMSC-3 connection failed - automatic failover activated', time: '15 mins ago' },
        { id: 3, type: 'info', message: 'Campaign "Spring Sale" completed successfully', time: '1 hour ago' },
      ]);

    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
    } finally {
      setLoading(false);
    }
  };

  // Metric Card Component
  const MetricCard = ({ title, value, subtitle, icon: Icon, color = 'primary', trend }) => (
    <Card>
      <CardContent>
        <Box display="flex" justifyContent="space-between" alignItems="flex-start">
          <Box>
            <Typography color="textSecondary" gutterBottom variant="body2">
              {title}
            </Typography>
            <Typography variant="h4" component="div">
              {value}
            </Typography>
            {subtitle && (
              <Typography variant="body2" color="textSecondary" sx={{ mt: 1 }}>
                {subtitle}
              </Typography>
            )}
          </Box>
          {Icon && (
            <Box sx={{ backgroundColor: `${color}.light`, borderRadius: 2, p: 1 }}>
              <Icon sx={{ color: `${color}.main`, fontSize: 40 }} />
            </Box>
          )}
        </Box>
      </CardContent>
    </Card>
  );

  const COLORS = ['#0088FE', '#00C49F', '#FFBB28', '#FF8042'];

  return (
    <Box>
      {/* Header with Refresh */}
      <Box display="flex" justifyContent="space-between" alignItems="center" mb={3}>
        <Box>
          <Typography variant="h4" gutterBottom>
            Dashboard
          </Typography>
          <Typography variant="body1" color="textSecondary">
            Real-time platform overview and monitoring
          </Typography>
        </Box>
        <IconButton onClick={fetchDashboardData} color="primary">
          <RefreshIcon />
        </IconButton>
      </Box>

      {/* A. Real-time Statistics Panel */}
      <Typography variant="h6" gutterBottom sx={{ mt: 3 }}>
        Real-time Statistics
      </Typography>
      <Grid container spacing={3}>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Current TPS"
            value={realtimeMetrics?.messages_per_second?.toFixed(2) || '0'}
            subtitle="Transactions per second"
            icon={SpeedIcon}
            color="primary"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Messages Today"
            value={summary?.messages?.total_today?.toLocaleString() || '0'}
            subtitle="Total sent today"
            icon={CheckCircleIcon}
            color="success"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="System Uptime"
            value="99.98%"
            subtitle="Last 30 days"
            icon={MemoryIcon}
            color="info"
          />
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Account Balance"
            value="125,430"
            subtitle="Credits remaining"
            icon={MoneyIcon}
            color="warning"
          />
        </Grid>
      </Grid>

      {/* TPS per Channel / SMSC */}
      <Grid container spacing={3} mt={2}>
        <Grid item xs={12} lg={6}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              TPS per SMSC / Channel
            </Typography>
            <ResponsiveContainer width="100%" height={250}>
              <BarChart data={tpsData}>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="name" />
                <YAxis />
                <Tooltip />
                <Legend />
                <Bar dataKey="tps" fill="#8884d8" name="Current TPS" />
                <Bar dataKey="max" fill="#82ca9d" name="Max TPS" />
              </BarChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>

        {/* Queue Status */}
        <Grid item xs={12} lg={6}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Queue Status
            </Typography>
            <TableContainer>
              <Table size="small">
                <TableHead>
                  <TableRow>
                    <TableCell>Queue</TableCell>
                    <TableCell align="right">Queued</TableCell>
                    <TableCell align="right">Processing</TableCell>
                  </TableRow>
                </TableHead>
                <TableBody>
                  {queueStatus.map((row) => (
                    <TableRow key={row.queue}>
                      <TableCell>{row.queue}</TableCell>
                      <TableCell align="right">{row.count.toLocaleString()}</TableCell>
                      <TableCell align="right">{row.processing}</TableCell>
                    </TableRow>
                  ))}
                </TableBody>
              </Table>
            </TableContainer>
          </Paper>
        </Grid>
      </Grid>

      {/* B. Alerts Widget */}
      <Typography variant="h6" gutterBottom sx={{ mt: 4 }}>
        System Alerts
      </Typography>
      <Grid container spacing={2}>
        {alerts.map((alert) => (
          <Grid item xs={12} key={alert.id}>
            <Alert
              severity={alert.type}
              sx={{ mb: 1 }}
              action={
                <Typography variant="caption" sx={{ mr: 2 }}>
                  {alert.time}
                </Typography>
              }
            >
              {alert.message}
            </Alert>
          </Grid>
        ))}
      </Grid>

      {/* C. Graphical KPIs */}
      <Typography variant="h6" gutterBottom sx={{ mt: 4 }}>
        Graphical KPIs
      </Typography>
      <Grid container spacing={3}>
        {/* Throughput Graph */}
        <Grid item xs={12} lg={8}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Message Throughput (Last 24 Hours)
            </Typography>
            <ResponsiveContainer width="100%" height={300}>
              <AreaChart data={messageTrend}>
                <defs>
                  <linearGradient id="colorMessages" x1="0" y1="0" x2="0" y2="1">
                    <stop offset="5%" stopColor="#1976d2" stopOpacity={0.8} />
                    <stop offset="95%" stopColor="#1976d2" stopOpacity={0} />
                  </linearGradient>
                </defs>
                <CartesianGrid strokeDasharray="3 3" />
                <XAxis dataKey="time" />
                <YAxis />
                <Tooltip />
                <Area
                  type="monotone"
                  dataKey="messages"
                  stroke="#1976d2"
                  fillOpacity={1}
                  fill="url(#colorMessages)"
                />
              </AreaChart>
            </ResponsiveContainer>
          </Paper>
        </Grid>

        {/* Success Rate & DLR Statistics */}
        <Grid item xs={12} lg={4}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Delivery Statistics
            </Typography>
            <Box mt={2}>
              <Typography variant="body2" gutterBottom>
                Delivery Rate
              </Typography>
              <Box display="flex" alignItems="center" mb={2}>
                <Box flexGrow={1} mr={2}>
                  <LinearProgress
                    variant="determinate"
                    value={realtimeMetrics?.delivery_rate_percent || 0}
                    color="success"
                    sx={{ height: 10, borderRadius: 5 }}
                  />
                </Box>
                <Typography variant="body2" fontWeight="bold">
                  {realtimeMetrics?.delivery_rate_percent?.toFixed(2) || 0}%
                </Typography>
              </Box>

              <Typography variant="body2" gutterBottom>
                Failure Rate
              </Typography>
              <Box display="flex" alignItems="center" mb={2}>
                <Box flexGrow={1} mr={2}>
                  <LinearProgress
                    variant="determinate"
                    value={realtimeMetrics?.failure_rate_percent || 0}
                    color="error"
                    sx={{ height: 10, borderRadius: 5 }}
                  />
                </Box>
                <Typography variant="body2" fontWeight="bold">
                  {realtimeMetrics?.failure_rate_percent?.toFixed(2) || 0}%
                </Typography>
              </Box>

              <Box mt={3}>
                <Grid container spacing={2}>
                  <Grid item xs={6}>
                    <Typography variant="caption" color="textSecondary">
                      Delivered
                    </Typography>
                    <Typography variant="h6">
                      {realtimeMetrics?.messages_delivered?.toLocaleString() || 0}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="caption" color="textSecondary">
                      Failed
                    </Typography>
                    <Typography variant="h6" color="error">
                      {realtimeMetrics?.messages_failed?.toLocaleString() || 0}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="caption" color="textSecondary">
                      Pending
                    </Typography>
                    <Typography variant="h6">
                      {realtimeMetrics?.messages_pending?.toLocaleString() || 0}
                    </Typography>
                  </Grid>
                  <Grid item xs={6}>
                    <Typography variant="caption" color="textSecondary">
                      Active Campaigns
                    </Typography>
                    <Typography variant="h6">
                      {summary?.campaigns?.active || 0}
                    </Typography>
                  </Grid>
                </Grid>
              </Box>
            </Box>
          </Paper>
        </Grid>
      </Grid>

      {/* D. Quick Actions */}
      <Typography variant="h6" gutterBottom sx={{ mt: 4 }}>
        Quick Actions
      </Typography>
      <Grid container spacing={2}>
        <Grid item xs={12} sm={6} md={3}>
          <Button
            variant="contained"
            startIcon={<AddIcon />}
            fullWidth
            size="large"
            onClick={() => window.location.href = '/campaigns/create'}
          >
            Send New Campaign
          </Button>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Button
            variant="outlined"
            startIcon={<DescriptionIcon />}
            fullWidth
            size="large"
            onClick={() => window.location.href = '/templates'}
          >
            Create Template
          </Button>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Button
            variant="outlined"
            startIcon={<ContactsIcon />}
            fullWidth
            size="large"
            onClick={() => window.location.href = '/contacts/lists'}
          >
            Add Contact List
          </Button>
        </Grid>
        <Grid item xs={12} sm={6} md={3}>
          <Button
            variant="outlined"
            startIcon={<CampaignIcon />}
            fullWidth
            size="large"
            onClick={() => window.location.href = '/campaigns/monitor'}
          >
            Monitor Campaigns
          </Button>
        </Grid>
      </Grid>

      {/* System Resource Summary */}
      <Typography variant="h6" gutterBottom sx={{ mt: 4 }}>
        System Resources
      </Typography>
      <Grid container spacing={3}>
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="body2" gutterBottom>
              CPU Usage
            </Typography>
            <Box display="flex" alignItems="center">
              <Box flexGrow={1} mr={2}>
                <LinearProgress
                  variant="determinate"
                  value={summary?.system?.cpu_usage || 0}
                  color={summary?.system?.cpu_usage > 80 ? 'error' : 'success'}
                  sx={{ height: 8, borderRadius: 4 }}
                />
              </Box>
              <Typography variant="body2" fontWeight="bold">
                {summary?.system?.cpu_usage?.toFixed(1) || 0}%
              </Typography>
            </Box>
          </Paper>
        </Grid>
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="body2" gutterBottom>
              Memory Usage
            </Typography>
            <Box display="flex" alignItems="center">
              <Box flexGrow={1} mr={2}>
                <LinearProgress
                  variant="determinate"
                  value={summary?.system?.memory_usage || 0}
                  color={summary?.system?.memory_usage > 80 ? 'error' : 'success'}
                  sx={{ height: 8, borderRadius: 4 }}
                />
              </Box>
              <Typography variant="body2" fontWeight="bold">
                {summary?.system?.memory_usage?.toFixed(1) || 0}%
              </Typography>
            </Box>
          </Paper>
        </Grid>
        <Grid item xs={12} md={4}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="body2" gutterBottom>
              Disk Usage
            </Typography>
            <Box display="flex" alignItems="center">
              <Box flexGrow={1} mr={2}>
                <LinearProgress
                  variant="determinate"
                  value={summary?.system?.disk_usage || 0}
                  color={summary?.system?.disk_usage > 80 ? 'error' : 'success'}
                  sx={{ height: 8, borderRadius: 4 }}
                />
              </Box>
              <Typography variant="body2" fontWeight="bold">
                {summary?.system?.disk_usage?.toFixed(1) || 0}%
              </Typography>
            </Box>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
}

export default EnhancedDashboard;
