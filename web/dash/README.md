# Dashboard App

This is the user dashboard for Highway WebAuthn authentication gateway. It provides a comprehensive interface for user management, console interactions, and system monitoring, integrating with the Blockchain x/svc module.

## Features

- **User Management**: Complete user account management and profile editing
- **Console Interface**: Interactive console for system operations
- **System Monitoring**: Real-time monitoring of service health and performance
- **Blockchain Integration**: Direct integration with Sonr Blockchain x/svc module
- **Authentication**: WebAuthn-based secure authentication
- **Responsive Design**: Mobile-first approach with modern UI/UX

## Development

To start the development server:

```bash
pnpm dev
```

This will start the Next.js development server with hot reload at `http://localhost:3001`.

## Testing

```bash
# Run tests
pnpm test

# Test dashboard functionality
# Visit http://localhost:3001 and test management features
```

## Deployment

To deploy to Cloudflare Pages:

```bash
pnpm deploy
```

The app is configured for static export and optimized for Cloudflare Pages deployment.

## Configuration

### Environment Variables

```env
# API endpoints
NEXT_PUBLIC_API_URL=https://api.yourdomain.com
NEXT_PUBLIC_BLOCKCHAIN_URL=https://blockchain.yourdomain.com

# Development
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_BLOCKCHAIN_URL=http://localhost:8080
```

## Development Roadmap

### 🔮 **Future Implementation** (Planned)

- [ ] **Dashboard Framework**: Next.js 14 setup with TypeScript configuration
- [ ] **User Management**: User account creation, editing, and deletion
- [ ] **Console Interface**: Interactive console for system operations
- [ ] **System Monitoring**: Real-time service health and performance metrics
- [ ] **Blockchain Integration**: Direct integration with Sonr Blockchain x/svc module
- [ ] **Authentication**: WebAuthn-based secure authentication
- [ ] **Responsive Design**: Mobile-first approach with modern UI/UX

### 🚧 **Production Readiness** (Next)

- [ ] **Performance Optimization**: Code splitting and bundle optimization
- [ ] **Testing Suite**: Unit tests for components and hooks
- [ ] **E2E Testing**: Cypress or Playwright for full user flow testing
- [ ] **Error Boundaries**: React error boundaries for graceful error handling
- [ ] **Analytics**: User behavior tracking and conversion metrics
- [ ] **Security**: Role-based access control and audit logging

### 🔮 **Future Enhancements**

- [ ] **Advanced Analytics**: Detailed system analytics and reporting
- [ ] **Multi-tenant Support**: Organization-based dashboard isolation
- [ ] **Advanced UI**: Dark mode, animations, and enhanced UX
- [ ] **Mobile App**: React Native or native mobile applications
- [ ] **API Management**: Advanced API key management and monitoring
- [ ] **Backup & Recovery**: System backup and disaster recovery features

## Architecture

The dashboard follows modern React patterns:

- **App Router**: Next.js 13+ app directory structure
- **Custom Hooks**: Reusable logic for dashboard operations
- **Component Library**: Shared UI components from `@sonr.io/ui`
- **State Management**: React hooks for local state management
- **Styling**: Tailwind CSS with custom dashboard styling

## Security

- **WebAuthn Authentication**: Secure admin authentication
- **Role-based Access**: Different permission levels for dashboard features
- **Input Validation**: Client-side validation with server-side verification
- **Audit Logging**: Complete audit trail for dashboard operations
- **Session Management**: Secure session handling and timeout

## Integration

The dashboard integrates with:

- **API Gateway**: Provides data for dashboard views
- **Blockchain Service**: Direct blockchain operations and monitoring
- **Monitoring**: System metrics and performance data
- **Authentication**: User authentication and session management
