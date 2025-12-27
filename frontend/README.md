# Fjord Bank Frontend

## Purpose

React single-page application for banking customers. Handles authentication, account management, and transfers.

## Running

```bash
cd frontend
npm install
npm run dev
```

Opens at http://localhost:5173

## Architecture

```
src/
  ├── api/
  │   └── client.ts        → API client with token management
  ├── components/
  │   ├── ui/              → Shadcn/ui components (button, card, input, label)
  │   └── ProtectedRoute.tsx → Auth guard for routes
  ├── context/
  │   └── AuthContext.tsx  → Global auth state
  ├── pages/
  │   ├── LoginPage.tsx    → Login form
  │   ├── RegisterPage.tsx → Registration form
  │   ├── DashboardPage.tsx → Account list and balances
  │   └── TransferPage.tsx → Transfer form
  ├── App.tsx              → Router setup
  └── main.tsx             → Entry point
```

## Technology Stack

- **Vite 5** - Build tool with hot module replacement
- **React 18** - UI library
- **TypeScript** - Type safety
- **React Router v7** - Client-side routing
- **Tailwind CSS v3** - Utility-first styling
- **Shadcn/ui** - Copy-paste component library

## Authentication Flow

1. **Login** calls `/auth/login`, receives access token in body
2. **Access token** stored in memory (not localStorage)
3. **Refresh token** stored as HttpOnly cookie (automatic)
4. **API client** adds `Authorization: Bearer <token>` to requests
5. **On 401**, client calls `/auth/refresh` automatically
6. **AuthContext** provides `isAuthenticated`, `login()`, `logout()`

## Key Components

### AuthContext
Global auth state provider. Checks for existing session on mount by attempting token refresh.

### ProtectedRoute
Wraps routes that require authentication. Redirects to `/login` if not authenticated.

### API Client
- `login(email, password)` - Authenticate, store token
- `logout()` - Clear token, call logout endpoint
- `getAccounts()` - Fetch customer's accounts
- `createTransfer(...)` - Submit transfer with idempotency key

## Design Decisions

**Why memory for access token:** XSS attacks can read localStorage. Memory is cleared on tab close, but refresh token cookie handles persistence.

**Why React Context for auth:** Single source of truth for auth state. Avoids prop drilling. Any component can access `useAuth()`.

**Why Shadcn/ui:** Components are copied into project, not imported from npm. Full control, no version conflicts, matches Tailwind styling.

**Why Vite over CRA:** Faster dev server, instant HMR, smaller production bundles, ESM-native.
