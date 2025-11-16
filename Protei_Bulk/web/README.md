# Protei_Bulk Web Dashboard

Modern React-based web interface for the Protei_Bulk enterprise messaging platform.

## Features

- **Dashboard**: Real-time statistics and system overview
- **Message Management**: Send single/bulk messages, view history
- **Campaign Management**: Create, schedule, and monitor campaigns
- **User Management**: RBAC-based user administration
- **Reports & Analytics**: Comprehensive reporting with export
- **Settings**: System configuration and preferences

## Technology Stack

- **React 18**: Modern React with hooks
- **Material-UI (MUI)**: Component library
- **React Router**: Client-side routing
- **Axios**: HTTP client
- **Zustand**: State management
- **React Query**: Server state management
- **Recharts**: Data visualization
- **Formik + Yup**: Form handling and validation
- **Socket.IO**: Real-time updates

## Quick Start

### Prerequisites

- Node.js 16+ and npm/yarn
- Protei_Bulk backend running on http://localhost:8080

### Installation

```bash
# Install dependencies
npm install

# Start development server
npm start
```

The application will open at http://localhost:3000

### Build for Production

```bash
# Create production build
npm run build

# The build folder contains the optimized production build
```

## Configuration

Create a `.env` file in the web directory:

```env
REACT_APP_API_URL=http://localhost:8080/api/v1
REACT_APP_WS_URL=ws://localhost:8080
```

## Project Structure

```
web/
├── public/              # Static assets
├── src/
│   ├── components/      # Reusable components
│   │   ├── Layout/      # Layout components
│   │   ├── Charts/      # Chart components
│   │   └── Common/      # Common UI components
│   ├── pages/           # Page components
│   │   ├── Auth/        # Authentication pages
│   │   ├── Dashboard/   # Dashboard
│   │   ├── Messages/    # Message management
│   │   ├── Campaigns/   # Campaign management
│   │   ├── Users/       # User management
│   │   └── Reports/     # Reports
│   ├── services/        # API services
│   ├── store/           # State management
│   ├── utils/           # Utility functions
│   ├── App.jsx          # Main app component
│   └── index.js         # Entry point
├── package.json
└── README.md
```

## Features by Module

### Dashboard
- Real-time message statistics
- System health indicators
- Active campaigns overview
- Recent activity feed
- Quick actions

### Messages
- Send single message
- Bulk message upload (CSV/Excel)
- Message templates
- Message history and search
- Delivery status tracking
- Filter by status, date, sender

### Campaigns
- Create campaign wizard
- Schedule campaigns
- Upload recipient lists
- Profile-based targeting
- Campaign monitoring (real-time)
- Pause/resume/stop campaigns
- Campaign analytics

### Users
- User list with search
- Create/edit users
- Role assignment (RBAC)
- Account management
- API key generation
- 2FA setup

### Reports
- Message reports (by date, account, status)
- Campaign performance
- System utilization
- Custom report builder
- Export to Excel/CSV/PDF
- Scheduled reports

## Development

### Available Scripts

```bash
# Start dev server
npm start

# Run tests
npm test

# Build production
npm run build

# Lint code
npm run lint
```

### Code Style

This project uses ESLint and Prettier for code formatting.

```bash
# Run linter
npm run lint

# Format code
npm run format
```

## Deployment

### Nginx Configuration

```nginx
server {
    listen 80;
    server_name dashboard.protei.com;

    root /var/www/protei-bulk-dashboard/build;
    index index.html;

    location / {
        try_files $uri $uri/ /index.html;
    }

    location /api/ {
        proxy_pass http://localhost:8080;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_cache_bypass $http_upgrade;
    }
}
```

### Docker Deployment

```dockerfile
# Build stage
FROM node:18-alpine as build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY . .
RUN npm run build

# Production stage
FROM nginx:alpine
COPY --from=build /app/build /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf
EXPOSE 80
CMD ["nginx", "-g", "daemon off;"]
```

## API Integration

The dashboard connects to the Protei_Bulk API at `/api/v1`. All API calls are made through the `api` service (`src/services/api.js`).

### Authentication

JWT-based authentication with automatic token refresh:

```javascript
import { useAuthStore } from './store/authStore';

const { login, logout } = useAuthStore();

// Login
await login('username', 'password');

// Logout
logout();
```

### Making API Calls

```javascript
import { messagesAPI } from './services/api';

// Send message
const result = await messagesAPI.send({
  from: '1234',
  to: '9876543210',
  text: 'Hello World'
});

// Get message list
const messages = await messagesAPI.list({
  page: 1,
  limit: 50,
  status: 'DELIVERED'
});
```

## Real-time Updates

The dashboard uses Socket.IO for real-time updates:

```javascript
import io from 'socket.io-client';

const socket = io(process.env.REACT_APP_WS_URL);

socket.on('campaign:progress', (data) => {
  // Update campaign progress
  console.log('Campaign progress:', data);
});

socket.on('message:delivered', (data) => {
  // Update message status
  console.log('Message delivered:', data);
});
```

## Browser Support

- Chrome (latest)
- Firefox (latest)
- Safari (latest)
- Edge (latest)

## License

© 2025 Protei Corporation. All rights reserved.
