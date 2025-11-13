// Protei_Monitoring v2.0 - Dashboard JavaScript
// Professional monitoring system with real-time updates

class ProteiDashboard {
    constructor() {
        this.apiBase = '/api';
        this.token = localStorage.getItem('auth_token');
        this.charts = {};
        this.wsConnection = null;
        this.currentPage = 'dashboard';
        this.refreshInterval = 5000; // 5 seconds
        this.timers = [];

        this.init();
    }

    init() {
        // Check authentication
        if (!this.token && window.location.pathname !== '/login.html') {
            window.location.href = '/login.html';
            return;
        }

        // Setup event listeners
        this.setupEventListeners();

        // Initialize WebSocket for real-time updates
        this.initWebSocket();

        // Load initial data
        this.loadDashboard();

        // Start auto-refresh
        this.startAutoRefresh();
    }

    setupEventListeners() {
        // Navigation
        document.querySelectorAll('.nav-item').forEach(item => {
            item.addEventListener('click', (e) => {
                e.preventDefault();
                const page = item.dataset.page;
                this.navigateTo(page);
            });
        });

        // Logout
        const logoutBtn = document.getElementById('logout-btn');
        if (logoutBtn) {
            logoutBtn.addEventListener('click', () => this.logout());
        }

        // Search
        const searchInput = document.getElementById('search-input');
        if (searchInput) {
            searchInput.addEventListener('input', (e) => this.handleSearch(e.target.value));
        }
    }

    // API Methods
    async apiCall(endpoint, method = 'GET', data = null) {
        const options = {
            method,
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${this.token}`
            }
        };

        if (data) {
            options.body = JSON.stringify(data);
        }

        try {
            const response = await fetch(`${this.apiBase}${endpoint}`, options);

            if (response.status === 401) {
                this.logout();
                return null;
            }

            if (!response.ok) {
                throw new Error(`API error: ${response.statusText}`);
            }

            return await response.json();
        } catch (error) {
            console.error('API call failed:', error);
            this.showNotification('API Error', error.message, 'danger');
            return null;
        }
    }

    // WebSocket for real-time updates
    initWebSocket() {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/ws?token=${this.token}`;

        this.wsConnection = new WebSocket(wsUrl);

        this.wsConnection.onopen = () => {
            console.log('WebSocket connected');
            this.showNotification('Connected', 'Real-time updates active', 'success');
        };

        this.wsConnection.onmessage = (event) => {
            const data = JSON.parse(event.data);
            this.handleRealtimeUpdate(data);
        };

        this.wsConnection.onerror = (error) => {
            console.error('WebSocket error:', error);
        };

        this.wsConnection.onclose = () => {
            console.log('WebSocket disconnected, reconnecting...');
            setTimeout(() => this.initWebSocket(), 5000);
        };
    }

    handleRealtimeUpdate(data) {
        switch (data.type) {
            case 'kpi_update':
                this.updateKPIs(data.payload);
                break;
            case 'alarm':
                this.handleNewAlarm(data.payload);
                break;
            case 'session_update':
                this.updateSessionTable(data.payload);
                break;
            case 'resource_update':
                this.updateResourceGraphs(data.payload);
                break;
        }
    }

    // Dashboard Loading
    async loadDashboard() {
        this.showLoading();

        try {
            // Load all dashboard data in parallel
            const [kpis, sessions, alarms, resources, license] = await Promise.all([
                this.apiCall('/kpi'),
                this.apiCall('/sessions?limit=100'),
                this.apiCall('/alarms?status=active'),
                this.apiCall('/resources'),
                this.apiCall('/license')
            ]);

            // Update UI components
            this.updateKPICards(kpis);
            this.updateSessionsTable(sessions);
            this.updateAlarmsPanel(alarms);
            this.updateResourceGraphs(resources);
            this.updateLicenseInfo(license);

            // Initialize charts
            this.initializeCharts();

        } catch (error) {
            console.error('Failed to load dashboard:', error);
            this.showNotification('Error', 'Failed to load dashboard data', 'danger');
        } finally {
            this.hideLoading();
        }
    }

    updateKPICards(kpis) {
        if (!kpis) return;

        // Total Sessions
        const totalSessions = document.getElementById('total-sessions');
        if (totalSessions) {
            totalSessions.textContent = this.formatNumber(kpis.total_sessions || 0);
        }

        // Success Rate
        const successRate = document.getElementById('success-rate');
        if (successRate) {
            const rate = kpis.success_rate || 0;
            successRate.textContent = rate.toFixed(2) + '%';

            // Update trend
            const trend = document.getElementById('success-rate-trend');
            if (trend && kpis.success_rate_change) {
                const change = kpis.success_rate_change;
                trend.textContent = Math.abs(change).toFixed(2) + '%';
                trend.className = 'stat-change ' + (change >= 0 ? 'positive' : 'negative');
            }
        }

        // Active Alarms
        const activeAlarms = document.getElementById('active-alarms');
        if (activeAlarms) {
            activeAlarms.textContent = kpis.active_alarms || 0;
        }

        // Average TPS
        const avgTps = document.getElementById('avg-tps');
        if (avgTps) {
            avgTps.textContent = this.formatNumber(kpis.avg_tps || 0);
        }
    }

    updateSessionsTable(sessions) {
        const tbody = document.getElementById('sessions-tbody');
        if (!tbody || !sessions) return;

        tbody.innerHTML = sessions.map(session => `
            <tr>
                <td><a href="#" onclick="dashboard.viewSession('${session.tid}')">${session.tid}</a></td>
                <td>${session.imsi || '-'}</td>
                <td>${session.msisdn || '-'}</td>
                <td><span class="badge info">${session.protocol}</span></td>
                <td>${session.procedure}</td>
                <td><span class="badge ${this.getStatusClass(session.status)}">${session.status}</span></td>
                <td>${this.formatTimestamp(session.timestamp)}</td>
            </tr>
        `).join('');
    }

    updateAlarmsPanel(alarms) {
        const container = document.getElementById('alarms-container');
        if (!container || !alarms) return;

        // Update badge count
        const badge = document.querySelector('.nav-item[data-page="alarms"] .nav-badge');
        if (badge) {
            badge.textContent = alarms.length;
        }

        container.innerHTML = alarms.slice(0, 5).map(alarm => `
            <div class="alarm-item alarm-${alarm.severity}">
                <div class="alarm-header">
                    <span class="badge ${alarm.severity}">${alarm.severity.toUpperCase()}</span>
                    <span class="alarm-time">${this.formatTimestamp(alarm.timestamp)}</span>
                </div>
                <div class="alarm-message">${alarm.message}</div>
                <div class="alarm-source">${alarm.source}</div>
            </div>
        `).join('');
    }

    updateResourceGraphs(resources) {
        if (!resources) return;

        // Update CPU chart
        if (this.charts.cpu) {
            this.updateChart(this.charts.cpu, resources.cpu);
        }

        // Update memory chart
        if (this.charts.memory) {
            this.updateChart(this.charts.memory, resources.memory);
        }

        // Update network chart
        if (this.charts.network) {
            this.updateChart(this.charts.network, resources.network);
        }
    }

    updateLicenseInfo(license) {
        if (!license) return;

        const container = document.getElementById('license-info');
        if (!container) return;

        const daysRemaining = Math.floor((new Date(license.expiry) - new Date()) / (1000 * 60 * 60 * 24));
        const statusClass = daysRemaining < 30 ? 'warning' : 'success';

        container.innerHTML = `
            <div class="license-card">
                <div class="license-status">
                    <span class="badge ${statusClass}">
                        ${daysRemaining} days remaining
                    </span>
                </div>
                <div class="license-details">
                    <p><strong>Customer:</strong> ${license.customer}</p>
                    <p><strong>Expiry:</strong> ${this.formatDate(license.expiry)}</p>
                    <p><strong>Max Subscribers:</strong> ${this.formatNumber(license.max_subscribers)}</p>
                    <p><strong>Max TPS:</strong> ${this.formatNumber(license.max_tps)}</p>
                </div>
                <div class="license-features">
                    <strong>Enabled Features:</strong>
                    <div class="feature-badges">
                        ${this.generateFeatureBadges(license)}
                    </div>
                </div>
            </div>
        `;
    }

    generateFeatureBadges(license) {
        const features = [];
        if (license.enable_2g) features.push('2G');
        if (license.enable_3g) features.push('3G');
        if (license.enable_4g) features.push('4G');
        if (license.enable_5g) features.push('5G');
        if (license.enable_map) features.push('MAP');
        if (license.enable_cap) features.push('CAP');
        if (license.enable_diameter) features.push('Diameter');
        if (license.enable_gtp) features.push('GTP');

        return features.map(f => `<span class="badge success">${f}</span>`).join(' ');
    }

    // Charts initialization
    initializeCharts() {
        // TPS Chart
        const tpsCtx = document.getElementById('tps-chart');
        if (tpsCtx) {
            this.charts.tps = new Chart(tpsCtx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: 'Transactions Per Second',
                        data: [],
                        borderColor: '#2563eb',
                        backgroundColor: 'rgba(37, 99, 235, 0.1)',
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            display: false
                        }
                    },
                    scales: {
                        y: {
                            beginAtZero: true
                        }
                    }
                }
            });
        }

        // Protocol Distribution Chart
        const protocolCtx = document.getElementById('protocol-chart');
        if (protocolCtx) {
            this.charts.protocol = new Chart(protocolCtx, {
                type: 'doughnut',
                data: {
                    labels: ['MAP', 'Diameter', 'GTP', 'HTTP', 'CAP', 'INAP'],
                    datasets: [{
                        data: [30, 25, 20, 15, 5, 5],
                        backgroundColor: [
                            '#2563eb', '#7c3aed', '#10b981',
                            '#f59e0b', '#ef4444', '#06b6d4'
                        ]
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    plugins: {
                        legend: {
                            position: 'bottom'
                        }
                    }
                }
            });
        }

        // Success Rate Chart
        const successCtx = document.getElementById('success-chart');
        if (successCtx) {
            this.charts.success = new Chart(successCtx, {
                type: 'bar',
                data: {
                    labels: ['Attach', 'Registration', 'PDU Session', 'Handover', 'Update Location'],
                    datasets: [{
                        label: 'Success Rate (%)',
                        data: [98.5, 99.2, 97.8, 96.5, 99.0],
                        backgroundColor: '#10b981'
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            max: 100
                        }
                    }
                }
            });
        }

        // CPU Chart
        const cpuCtx = document.getElementById('cpu-chart');
        if (cpuCtx) {
            this.charts.cpu = new Chart(cpuCtx, {
                type: 'line',
                data: {
                    labels: [],
                    datasets: [{
                        label: 'CPU Usage (%)',
                        data: [],
                        borderColor: '#7c3aed',
                        backgroundColor: 'rgba(124, 58, 237, 0.1)',
                        tension: 0.4,
                        fill: true
                    }]
                },
                options: {
                    responsive: true,
                    maintainAspectRatio: false,
                    scales: {
                        y: {
                            beginAtZero: true,
                            max: 100
                        }
                    }
                }
            });
        }
    }

    updateChart(chart, data) {
        if (!chart || !data) return;

        // Add new data point
        const now = new Date().toLocaleTimeString();
        chart.data.labels.push(now);
        chart.data.datasets[0].data.push(data);

        // Keep only last 20 points
        if (chart.data.labels.length > 20) {
            chart.data.labels.shift();
            chart.data.datasets[0].data.shift();
        }

        chart.update('none'); // Update without animation for performance
    }

    // Navigation
    navigateTo(page) {
        // Update active nav item
        document.querySelectorAll('.nav-item').forEach(item => {
            item.classList.remove('active');
        });
        document.querySelector(`.nav-item[data-page="${page}"]`)?.classList.add('active');

        this.currentPage = page;
        this.loadPage(page);
    }

    async loadPage(page) {
        const content = document.getElementById('content');
        if (!content) return;

        this.showLoading();

        try {
            switch (page) {
                case 'dashboard':
                    await this.loadDashboard();
                    break;
                case 'sessions':
                    await this.loadSessionsPage();
                    break;
                case 'configuration':
                    await this.loadConfigurationPage();
                    break;
                case 'alarms':
                    await this.loadAlarmsPage();
                    break;
                case 'users':
                    await this.loadUsersPage();
                    break;
                case 'logs':
                    await this.loadLogsPage();
                    break;
                case 'resources':
                    await this.loadResourcesPage();
                    break;
                default:
                    console.warn('Unknown page:', page);
            }
        } finally {
            this.hideLoading();
        }
    }

    // Utility Methods
    formatNumber(num) {
        if (num >= 1000000) {
            return (num / 1000000).toFixed(1) + 'M';
        } else if (num >= 1000) {
            return (num / 1000).toFixed(1) + 'K';
        }
        return num.toString();
    }

    formatTimestamp(timestamp) {
        const date = new Date(timestamp);
        return date.toLocaleString();
    }

    formatDate(date) {
        return new Date(date).toLocaleDateString();
    }

    getStatusClass(status) {
        const statusMap = {
            'success': 'success',
            'completed': 'success',
            'failed': 'danger',
            'error': 'danger',
            'active': 'info',
            'pending': 'warning'
        };
        return statusMap[status.toLowerCase()] || 'info';
    }

    showLoading() {
        const loader = document.getElementById('loading-overlay');
        if (loader) loader.style.display = 'flex';
    }

    hideLoading() {
        const loader = document.getElementById('loading-overlay');
        if (loader) loader.style.display = 'none';
    }

    showNotification(title, message, type = 'info') {
        // Create notification element
        const notification = document.createElement('div');
        notification.className = `notification notification-${type}`;
        notification.innerHTML = `
            <div class="notification-header">
                <strong>${title}</strong>
                <button onclick="this.parentElement.parentElement.remove()">×</button>
            </div>
            <div class="notification-body">${message}</div>
        `;

        // Add to container
        let container = document.getElementById('notifications-container');
        if (!container) {
            container = document.createElement('div');
            container.id = 'notifications-container';
            container.style.cssText = 'position: fixed; top: 80px; right: 20px; z-index: 9999;';
            document.body.appendChild(container);
        }

        container.appendChild(notification);

        // Auto remove after 5 seconds
        setTimeout(() => notification.remove(), 5000);
    }

    startAutoRefresh() {
        // Refresh KPIs every 5 seconds
        this.timers.push(setInterval(() => {
            if (this.currentPage === 'dashboard') {
                this.apiCall('/kpi').then(kpis => this.updateKPICards(kpis));
            }
        }, this.refreshInterval));

        // Refresh resource graphs every 10 seconds
        this.timers.push(setInterval(() => {
            this.apiCall('/resources').then(resources => this.updateResourceGraphs(resources));
        }, 10000));
    }

    logout() {
        this.apiCall('/auth/logout', 'POST').then(() => {
            localStorage.removeItem('auth_token');
            window.location.href = '/login.html';
        });
    }

    handleSearch(query) {
        console.log('Searching for:', query);
        // TODO: Implement search functionality
    }

    async viewSession(tid) {
        // TODO: Show session details modal
        const session = await this.apiCall(`/sessions/${tid}`);
        console.log('Session details:', session);
    }
}

// Initialize dashboard when DOM is ready
let dashboard;
if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => {
        dashboard = new ProteiDashboard();
    });
} else {
    dashboard = new ProteiDashboard();
}
