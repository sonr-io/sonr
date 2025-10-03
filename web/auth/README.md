# Auth Frontend

This is the Next.js authentication frontend for Highway WebAuthn gateway. It provides a modern, responsive interface for passwordless authentication using WebAuthn passkeys.

## Features

- **Passwordless Authentication**: WebAuthn registration and login with passkeys
- **Modern UI**: Clean, responsive design with Tailwind CSS
- **Session Management**: Persistent sessions with automatic validation
- **Browser Compatibility**: WebAuthn feature detection and fallback messaging
- **Accessibility**: ARIA labels, keyboard navigation, and screen reader support
- **TypeScript**: Full type safety with custom hooks and components

## Pages

### Home (`/`)

- Landing page with feature highlights
- Navigation to registration and login
- Automatic redirect to dashboard if authenticated

### Registration (`/register`)

- WebAuthn passkey registration flow
- Username and display name input
- Real-time validation and error handling

### Login (`/login`)

- WebAuthn passkey authentication
- Username input with passkey verification
- Remember me functionality

### Dashboard (`/dashboard`)

- User profile information
- Account security features
- Session management controls

## Components

### Custom Hooks

- `useWebAuthn`: WebAuthn registration and authentication logic
- `useSession`: Session management and user state

### UI Components (from `@sonr.io/ui`)

- `Button`: Configurable button with loading states
- `Input`: Form input with validation and help text
- `ErrorAlert`: Dismissible error notifications

## Development

To start the development server:

```bash
pnpm dev
```

This will start the Next.js development server with hot reload at `http://localhost:3000`.

## Testing

```bash
# Run tests
pnpm test

# Test WebAuthn in browser
# Visit http://localhost:3000 and test registration/login flow
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
# API endpoint
NEXT_PUBLIC_API_URL=https://api.yourdomain.com

# Development
NEXT_PUBLIC_API_URL=http://localhost:8080
```

### Next.js Configuration

The app uses:

- Static export for Cloudflare Pages
- Unoptimized images for static hosting
- Trailing slash for proper routing

## Development Roadmap

### ✅ **Phase 3: WebAuthn Frontend Implementation** (Current)

- [x] **Next.js 14 Setup**: App directory with TypeScript configuration
- [x] **WebAuthn Integration**: Complete registration and authentication flows
- [x] **Custom Hooks**: useWebAuthn and useSession for state management
- [x] **UI Components**: Reusable Button, Input, and ErrorAlert components
- [x] **Authentication Pages**: Registration, login, and dashboard interfaces
- [x] **Session Management**: Persistent sessions with browser compatibility
- [x] **Responsive Design**: Mobile-first approach with Tailwind CSS
- [x] **Accessibility**: ARIA labels, keyboard navigation, and screen reader support

### 🚧 **Phase 5: Production Readiness** (Next)

- [ ] **Performance Optimization**: Code splitting and bundle optimization
- [ ] **Testing Suite**: Unit tests for components and hooks
- [ ] **E2E Testing**: Cypress or Playwright for full user flow testing
- [ ] **Error Boundaries**: React error boundaries for graceful error handling
- [ ] **Analytics**: User behavior tracking and conversion metrics
- [ ] **Internationalization**: Multi-language support

### 🔮 **Future Enhancements**

- [ ] **Advanced UI**: Dark mode, animations, and enhanced UX
- [ ] **Progressive Web App**: PWA features for mobile experience
- [ ] **Multi-factor Authentication**: Additional security options
- [ ] **Account Management**: Profile editing and security settings
- [ ] **Admin Dashboard**: User management and system monitoring
- [ ] **Mobile App**: React Native or native mobile applications

## Architecture

The frontend follows modern React patterns:

- **App Router**: Next.js 13+ app directory structure
- **Custom Hooks**: Reusable logic for WebAuthn and session management
- **Component Library**: Shared UI components with TypeScript
- **State Management**: React hooks for local state, no external state library
- **Styling**: Tailwind CSS with custom component styling

## Security

- **WebAuthn**: Cryptographic authentication without passwords
- **Session Validation**: Automatic session validation and renewal
- **Input Validation**: Client-side validation with server-side verification
- **HTTPS Only**: Secure connection required for WebAuthn
- **Content Security Policy**: CSP headers for XSS protection

## Browser Support

- **Chrome**: Full WebAuthn support
- **Firefox**: Full WebAuthn support
- **Safari**: Full WebAuthn support (iOS 14+)
- **Edge**: Full WebAuthn support
- **Mobile**: Android and iOS with platform authenticators
