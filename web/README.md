# Next.js Web Application with Redux Saga

A modern Next.js application with Redux Saga for state management, featuring a dashboard interface with theme support.

## 🚀 Features

- **Next.js 15.5.2** with App Router and TypeScript
- **Redux Toolkit** + **Redux Saga** for advanced state management
- **Tailwind CSS 4.0** with comprehensive theme system
- **shadcn/ui** components for consistent UI
- **Theme Support** (Light, Dark, System)
- **Dashboard Interface** with navigation and interactive components
- **Mock API Integration** with Redux Saga for async operations

## 📦 Tech Stack

- **Framework**: Next.js 15.5.2
- **Language**: TypeScript
- **Styling**: Tailwind CSS 4.0
- **State Management**: Redux Toolkit + Redux Saga
- **UI Components**: shadcn/ui + Radix UI
- **Icons**: Lucide React
- **Build Tool**: Turbopack

## 🏗️ Project Structure

```
web/
├── src/
│   ├── app/                    # Next.js App Router
│   │   ├── dashboard/          # Dashboard page
│   │   ├── globals.css         # Global styles
│   │   └── layout.tsx          # Root layout with providers
│   ├── components/             # React components
│   │   ├── ui/                 # shadcn/ui components
│   │   ├── DashboardNavbar.tsx # Main navigation
│   │   ├── ThemeProvider.tsx   # Theme context (legacy)
│   │   └── ThemeToggle.tsx     # Theme switcher
│   ├── store/                  # Redux store
│   │   ├── slices/             # Redux Toolkit slices
│   │   │   ├── userSlice.ts    # User state management
│   │   │   ├── themeSlice.ts   # Theme state management
│   │   │   └── walletSlice.ts  # Wallet/Private key state
│   │   ├── sagas/              # Redux Saga effects
│   │   │   ├── userSaga.ts     # User async operations
│   │   │   ├── themeSaga.ts    # Theme async operations
│   │   │   └── walletSaga.ts   # Wallet async operations
│   │   ├── hooks.ts            # Typed Redux hooks
│   │   └── index.ts            # Store configuration
│   └── lib/                    # Utilities
│       └── utils.ts            # Helper functions
├── tailwind.config.ts          # Tailwind configuration
└── package.json                # Dependencies
```

## 🔧 Redux Saga Implementation

### Store Configuration
The application uses Redux Toolkit with Redux Saga middleware for handling complex async operations:

```typescript
// Store setup with Saga middleware
const sagaMiddleware = createSagaMiddleware()

export const store = configureStore({
  reducer: {
    user: userReducer,
    theme: themeReducer,
    wallet: walletReducer,
  },
  middleware: (getDefaultMiddleware) =>
    getDefaultMiddleware({
      thunk: false, // Disable thunk since we're using saga
    }).concat(sagaMiddleware),
})
```

### State Slices
- **User Slice**: Manages current user, users list, loading states, and errors
- **Theme Slice**: Handles theme preferences and system theme detection
- **Wallet Slice**: Manages wallets, private keys, and selection states

### Saga Effects
Each slice has corresponding sagas that handle:
- **API calls** with proper error handling
- **Loading states** management
- **Optimistic updates**
- **Complex workflows** with multiple async operations

### Usage Example
```typescript
// In components
const dispatch = useAppDispatch();
const { users, loading } = useAppSelector((state) => state.user);

// Dispatch saga actions
dispatch(userActions.fetchUsers());
dispatch(walletActions.createWallet({ name: 'New Wallet' }));
```

## 🎨 Theme System

The application features a comprehensive theme system with:
- **Light/Dark/System** theme options
- **CSS Variables** for consistent theming
- **Redux Saga** integration for theme persistence
- **System preference** detection

### Theme Colors
- Primary, Secondary, Neutral colors (50-950 shades)
- Semantic colors (background, foreground, card, border, etc.)
- Success, Warning, Error states

## 🧪 Testing

```bash
# Run development server
npm run dev

# Build for production
npm run build

# Start production server
npm start
```

## 📱 Dashboard Features

### Navigation
- **Logo** with Turnkey branding
- **9-dots menu** with dropdown options
- **Theme toggle** with loading states

### Interactive Sections
- **Users**: Display, create, and manage users
- **Wallets**: View and create wallets with balances
- **Private Keys**: Manage cryptographic keys
- **Theme**: Current theme status and controls
- **Redux State**: Real-time state monitoring

### Menu Actions
- **Users**: Fetches user list via Redux Saga
- **Wallet**: Loads wallet data with loading indicators
- **Private Keys**: Retrieves private key information

## 🔄 State Management Flow

1. **Component** dispatches saga action
2. **Saga** intercepts action and performs async operation
3. **Saga** dispatches success/error actions to slice
4. **Slice** updates state with new data
5. **Component** re-renders with updated state

## 🚀 Getting Started

1. **Install dependencies**:
   ```bash
   npm install
   ```

2. **Start development server**:
   ```bash
   npm run dev
   ```

3. **Open browser**:
   Navigate to `http://localhost:3000`

4. **Explore dashboard**:
   - Visit `/dashboard` for the main interface
   - Try the theme toggle in the navigation
   - Click the 9-dots menu to trigger saga actions
   - Use the interactive buttons to test Redux state

## 📚 Key Concepts

### Redux Saga vs Thunk
- **Saga**: Complex async workflows, cancellation, race conditions
- **Thunk**: Simple async operations, easier learning curve
- **This app**: Uses Saga for advanced state management patterns

### Generator Functions
Sagas use ES6 generators for better control flow:
```typescript
function* fetchUserSaga() {
  try {
    yield put(setLoading(true))
    const user = yield call(api.getUser)
    yield put(setUser(user))
  } catch (error) {
    yield put(setError(error.message))
  } finally {
    yield put(setLoading(false))
  }
}
```

### Effects
- `call`: Invoke async functions
- `put`: Dispatch actions
- `takeLatest`: Handle latest action, cancel previous
- `select`: Access current state
- `delay`: Add delays for UX

## 🔮 Future Enhancements

- **Real API Integration**: Replace mock APIs with actual endpoints
- **Authentication**: Add login/logout functionality
- **Real-time Updates**: WebSocket integration with sagas
- **Advanced Workflows**: Complex multi-step operations
- **Testing**: Unit tests for sagas and components
- **Performance**: Optimize saga patterns for large datasets
