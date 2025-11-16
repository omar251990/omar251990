import React, { useState, useEffect } from 'react';
import {
  Box,
  Grid,
  Paper,
  Typography,
  Card,
  CardContent,
  LinearProgress,
  Alert
} from '@mui/material';
import {
  TrendingUp as TrendingUpIcon,
  TrendingDown as TrendingDownIcon,
  Send as SendIcon,
  CheckCircle as CheckCircleIcon,
  Error as ErrorIcon,
  Campaign as CampaignIcon,
  Speed as SpeedIcon
} from '@mui/icons-material';
import { LineChart, Line, AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';
import { analyticsAPI } from '../../services/analyticsAPI';

/**
 * Dashboard Page
 * Displays real-time metrics, trends, and system overview
 */
function DashboardPage() {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(null);
  const [summary, setSummary] = useState(null);
  const [realtimeMetrics, setRealtimeMetrics] = useState(null);
  const [messageTrend, setMessageTrend] = useState([]);

  // Fetch dashboard data
  useEffect(() => {
    fetchDashboardData();
    const interval = setInterval(fetchDashboardData, 5000); // Refresh every 5 seconds

    return () => clearInterval(interval);
  }, []);

  const fetchDashboardData = async () => {
    try {
      setLoading(true);

      // Fetch summary and real-time metrics in parallel
      const [summaryRes, metricsRes, trendRes] = await Promise.all([
        analyticsAPI.dashboardSummary(),
        analyticsAPI.getRealtimeMessageMetrics(60),
        analyticsAPI.getMessageTrend('hour', 24)
      ]);

      setSummary(summaryRes.data.data);
      setRealtimeMetrics(metricsRes.data.data);

      // Transform trend data for chart
      const trendData = trendRes.data.data.data_points.map(point => ({
        time: new Date(point.timestamp).toLocaleTimeString(),
        messages: point.value
      }));
      setMessageTrend(trendData);

      setError(null);
    } catch (err) {
      console.error('Failed to fetch dashboard data:', err);
      setError('Failed to load dashboard data');
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
            {trend && (
              <Box display="flex" alignItems="center" mt={1}>
                {trend > 0 ? (
                  <TrendingUpIcon fontSize="small" color="success" />
                ) : (
                  <TrendingDownIcon fontSize="small" color="error" />
                )}
                <Typography variant="caption" sx={{ ml: 0.5 }}>
                  {Math.abs(trend)}%
                </Typography>
              </Box>
            )}
          </Box>
          {Icon && (
            <Box
              sx={{
                backgroundColor: `${color}.light`,
                borderRadius: 2,
                p: 1
              }}
            >
              <Icon sx={{ color: `${color}.main`, fontSize: 40 }} />
            </Box>
          )}
        </Box>
      </CardContent>
    </Card>
  );

  // System Health Card
  const SystemHealthCard = ({ title, value, max = 100, color }) => {
    const percentage = (value / max) * 100;
    const getColor = () => {
      if (percentage < 70) return 'success';
      if (percentage < 85) return 'warning';
      return 'error';
    };

    return (
      <Box mb={2}>
        <Box display="flex" justifyContent="space-between" mb={0.5}>
          <Typography variant="body2">{title}</Typography>
          <Typography variant="body2" fontWeight="bold">
            {value.toFixed(1)}%
          </Typography>
        </Box>
        <LinearProgress
          variant="determinate"
          value={percentage}
          color={color || getColor()}
          sx={{ height: 8, borderRadius: 1 }}
        />
      </Box>
    );
  };

  if (loading && !summary) {
    return (
      <Box display="flex" justifyContent="center" alignItems="center" minHeight="80vh">
        <LinearProgress sx={{ width: 200 }} />
      </Box>
    );
  }

  if (error && !summary) {
    return (
      <Box p={3}>
        <Alert severity="error">{error}</Alert>
      </Box>
    );
  }

  return (
    <Box p={3}>
      <Typography variant="h4" gutterBottom>
        Dashboard
      </Typography>
      <Typography variant="body1" color="textSecondary" gutterBottom>
        Real-time platform overview and metrics
      </Typography>

      {/* Key Metrics Row */}
      <Grid container spacing={3} mt={2}>
        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Messages Today"
            value={summary?.messages?.total_today?.toLocaleString() || '0'}
            subtitle="Total messages sent"
            icon={SendIcon}
            color="primary"
          />
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Current TPS"
            value={realtimeMetrics?.messages_per_second?.toFixed(2) || '0'}
            subtitle="Transactions per second"
            icon={SpeedIcon}
            color="info"
          />
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Delivery Rate"
            value={`${realtimeMetrics?.delivery_rate_percent?.toFixed(2) || '0'}%`}
            subtitle="Last hour"
            icon={CheckCircleIcon}
            color="success"
          />
        </Grid>

        <Grid item xs={12} sm={6} md={3}>
          <MetricCard
            title="Active Campaigns"
            value={summary?.campaigns?.active || '0'}
            subtitle="Running campaigns"
            icon={CampaignIcon}
            color="secondary"
          />
        </Grid>
      </Grid>

      {/* Charts and Detailed Metrics */}
      <Grid container spacing={3} mt={2}>
        {/* Message Volume Trend */}
        <Grid item xs={12} lg={8}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Message Volume (Last 24 Hours)
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

        {/* System Health */}
        <Grid item xs={12} lg={4}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              System Health
            </Typography>
            <Box mt={2}>
              <SystemHealthCard
                title="CPU Usage"
                value={summary?.system?.cpu_usage || 0}
              />
              <SystemHealthCard
                title="Memory Usage"
                value={summary?.system?.memory_usage || 0}
              />
              <SystemHealthCard
                title="Disk Usage"
                value={summary?.system?.disk_usage || 0}
              />
              <Box mt={2}>
                <Typography variant="body2" color="textSecondary">
                  Active Connections: {summary?.system?.active_connections || 0}
                </Typography>
              </Box>
            </Box>
          </Paper>
        </Grid>

        {/* Performance Metrics */}
        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Performance Metrics
            </Typography>
            <Grid container spacing={2} mt={1}>
              <Grid item xs={6}>
                <Typography variant="body2" color="textSecondary">
                  Avg Delivery Time
                </Typography>
                <Typography variant="h6">
                  {realtimeMetrics?.avg_delivery_time_seconds?.toFixed(2) || '0'}s
                </Typography>
              </Grid>
              <Grid item xs={6}>
                <Typography variant="body2" color="textSecondary">
                  P95 Delivery Time
                </Typography>
                <Typography variant="h6">
                  {realtimeMetrics?.p95_delivery_time_seconds?.toFixed(2) || '0'}s
                </Typography>
              </Grid>
              <Grid item xs={6}>
                <Typography variant="body2" color="textSecondary">
                  Messages Delivered
                </Typography>
                <Typography variant="h6">
                  {realtimeMetrics?.messages_delivered?.toLocaleString() || '0'}
                </Typography>
              </Grid>
              <Grid item xs={6}>
                <Typography variant="body2" color="textSecondary">
                  Messages Failed
                </Typography>
                <Typography variant="h6" color="error">
                  {realtimeMetrics?.messages_failed?.toLocaleString() || '0'}
                </Typography>
              </Grid>
            </Grid>
          </Paper>
        </Grid>

        {/* Quick Stats */}
        <Grid item xs={12} md={6}>
          <Paper sx={{ p: 3 }}>
            <Typography variant="h6" gutterBottom>
              Message Status Breakdown
            </Typography>
            <Grid container spacing={2} mt={1}>
              <Grid item xs={6}>
                <Box display="flex" alignItems="center" mb={1}>
                  <CheckCircleIcon color="success" sx={{ mr: 1 }} />
                  <Box>
                    <Typography variant="body2" color="textSecondary">
                      Delivered
                    </Typography>
                    <Typography variant="h6">
                      {realtimeMetrics?.messages_delivered || 0}
                    </Typography>
                  </Box>
                </Box>
              </Grid>
              <Grid item xs={6}>
                <Box display="flex" alignItems="center" mb={1}>
                  <ErrorIcon color="error" sx={{ mr: 1 }} />
                  <Box>
                    <Typography variant="body2" color="textSecondary">
                      Failed
                    </Typography>
                    <Typography variant="h6">
                      {realtimeMetrics?.messages_failed || 0}
                    </Typography>
                  </Box>
                </Box>
              </Grid>
              <Grid item xs={6}>
                <Box display="flex" alignItems="center">
                  <SendIcon color="info" sx={{ mr: 1 }} />
                  <Box>
                    <Typography variant="body2" color="textSecondary">
                      Pending
                    </Typography>
                    <Typography variant="h6">
                      {realtimeMetrics?.messages_pending || 0}
                    </Typography>
                  </Box>
                </Box>
              </Grid>
              <Grid item xs={6}>
                <Box display="flex" alignItems="center">
                  <ErrorIcon color="warning" sx={{ mr: 1 }} />
                  <Box>
                    <Typography variant="body2" color="textSecondary">
                      Rejected
                    </Typography>
                    <Typography variant="h6">
                      {realtimeMetrics?.messages_rejected || 0}
                    </Typography>
                  </Box>
                </Box>
              </Grid>
            </Grid>
          </Paper>
        </Grid>
      </Grid>
    </Box>
  );
}

export default DashboardPage;
