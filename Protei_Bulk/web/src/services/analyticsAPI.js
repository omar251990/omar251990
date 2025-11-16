import api from './api';

/**
 * Analytics API service
 * Provides methods for accessing analytics endpoints
 */

export const analyticsAPI = {
  // ========== Real-time Metrics ==========

  /**
   * Get real-time message metrics
   * @param {number} windowSeconds - Time window in seconds (default: 60)
   */
  getRealtimeMessageMetrics: (windowSeconds = 60) =>
    api.get('/analytics/metrics/messages/realtime', { params: { window_seconds: windowSeconds } }),

  /**
   * Get campaign metrics
   * @param {string} campaignId - Campaign ID
   */
  getCampaignMetrics: (campaignId) =>
    api.get(`/analytics/metrics/campaigns/${campaignId}`),

  /**
   * Get system resource metrics
   */
  getSystemMetrics: () =>
    api.get('/analytics/metrics/system'),

  /**
   * Get account usage metrics
   * @param {string} accountId - Account ID
   */
  getAccountMetrics: (accountId) =>
    api.get(`/analytics/metrics/accounts/${accountId}`),

  // ========== Trend Analysis ==========

  /**
   * Get message volume trend
   * @param {string} granularity - minute, hour, or day
   * @param {number} hours - Hours to look back
   */
  getMessageTrend: (granularity = 'hour', hours = 24) =>
    api.get('/analytics/trends/messages', { params: { granularity, hours } }),

  // ========== Predictive Analytics ==========

  /**
   * Predict message volume
   * @param {number} hoursAhead - Hours to predict ahead
   */
  predictMessageVolume: (hoursAhead = 24) =>
    api.get('/analytics/predictions/message-volume', { params: { hours_ahead: hoursAhead } }),

  // ========== Report Generation ==========

  /**
   * Generate message delivery report
   * @param {Object} params - Report parameters
   * @param {string} params.startDate - Start date (ISO format)
   * @param {string} params.endDate - End date (ISO format)
   * @param {string} params.format - csv, json, excel, or pdf
   * @param {string} params.accountId - Filter by account (optional)
   * @param {string} params.status - Filter by status (optional)
   * @param {string} params.campaignId - Filter by campaign (optional)
   */
  generateMessageReport: (params) =>
    api.post('/analytics/reports/messages', null, { params }),

  /**
   * Generate campaign performance report
   * @param {string} campaignId - Campaign ID
   * @param {string} format - csv, json, excel, or pdf
   */
  generateCampaignReport: (campaignId, format = 'json') =>
    api.post(`/analytics/reports/campaigns/${campaignId}`, null, { params: { format } }),

  /**
   * Generate account usage report
   * @param {string} accountId - Account ID
   * @param {string} startDate - Start date (ISO format)
   * @param {string} endDate - End date (ISO format)
   * @param {string} format - csv, json, excel, or pdf
   */
  generateAccountUsageReport: (accountId, startDate, endDate, format = 'excel') =>
    api.post(`/analytics/reports/accounts/${accountId}/usage`, null, {
      params: { start_date: startDate, end_date: endDate, format }
    }),

  // ========== Dashboard Summary ==========

  /**
   * Get comprehensive dashboard summary
   */
  dashboardSummary: () =>
    api.get('/analytics/dashboard/summary'),
};

export default analyticsAPI;
