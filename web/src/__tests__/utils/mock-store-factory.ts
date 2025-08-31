import { configureStore } from '@reduxjs/toolkit';
import authReducer, { AuthState, AuthSession, AuthUser } from '@/store/authSlice';
import usersReducer, { UsersState, User } from '@/store/usersSlice';
import walletReducer, { WalletState } from '@/store/slices/walletSlice';
import userReducer, { UserState } from '@/store/slices/userSlice';
import themeReducer, { ThemeState } from '@/store/slices/themeSlice';
import privateKeysReducer, { PrivateKeysState } from '@/store/privateKeysSlice';
import { gauthApi } from '@/services/gauthApi'; // For RTK Query middleware

// Test data fixtures
export const mockAuthUser: AuthUser = {
    id: 'user-1',
    organization_id: 'org-1',
    username: 'testuser',
    email: 'test@example.com',
    public_key: '0x1234567890abcdef',
    is_active: true,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
};

export const mockAuthSession: AuthSession = {
    session_token: 'mock-session-token',
    expires_at: new Date(Date.now() + 24 * 60 * 60 * 1000).toISOString(), // 24 hours from now
    user: mockAuthUser,
    auth_method: {
        id: 'auth-method-1',
        type: 'oauth',
        name: 'Google OAuth',
    },
};

export const mockUsers: User[] = [
    {
        id: 'user-1',
        organization_id: 'org-1',
        username: 'testuser',
        email: 'test@example.com',
        public_key: '0x1234567890abcdef',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
    },
    {
        id: 'user-2',
        organization_id: 'org-1',
        username: 'inactiveuser',
        email: 'inactive@example.com',
        public_key: '0xfedcba0987654321',
        is_active: false,
        created_at: '2024-01-02T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
    },
];

export const mockWallets = [
    {
        id: 'wallet-1',
        organization_id: 'org-1',
        name: 'Test Wallet 1',
        public_key: '0x1234567890abcdef1234567890abcdef12345678',
        seed_phrase: 'test seed phrase one',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
    },
    {
        id: 'wallet-2',
        organization_id: 'org-1',
        name: 'Test Wallet 2',
        public_key: '0xfedcba0987654321fedcba0987654321fedcba09',
        seed_phrase: 'test seed phrase two',
        is_active: false,
        created_at: '2024-01-02T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
    },
];

export const mockPrivateKeys = [
    {
        id: 'pkey-1',
        organization_id: 'org-1',
        wallet_id: 'wallet-1',
        name: 'Test Private Key 1',
        public_key: '0x1234567890abcdef1234567890abcdef12345678',
        curve: 'secp256k1',
        path: "m/44'/60'/0'/0/0",
        tags: ['main'],
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
    },
    {
        id: 'pkey-2',
        organization_id: 'org-1',
        wallet_id: 'wallet-1',
        name: 'Test Private Key 2',
        public_key: '0xfedcba0987654321fedcba0987654321fedcba09',
        curve: 'secp256k1',
        path: "m/44'/60'/0'/0/1",
        tags: ['backup'],
        is_active: false,
        created_at: '2024-01-02T00:00:00Z',
        updated_at: '2024-01-02T00:00:00Z',
    },
];

// Default state configurations
export const defaultAuthState: AuthState = {
    isAuthenticated: true,
    session: mockAuthSession,
    loading: false,
    error: null,
};

export const defaultUsersState: UsersState = {
    users: mockUsers,
    loading: false,
    error: null,
    currentPage: 1,
    totalPages: 1,
    totalUsers: mockUsers.length,
};

export const defaultWalletState: WalletState = {
    wallets: mockWallets,
    loading: false,
    error: null,
};

export const defaultUserState: UserState = {
    user: mockAuthUser,
    loading: false,
    error: null,
};

export const defaultThemeState: ThemeState = {
    theme: 'light',
};

export const defaultPrivateKeysState: PrivateKeysState = {
    privateKeys: mockPrivateKeys,
    loading: false,
    error: null,
    currentPage: 1,
    totalPages: 1,
    totalPrivateKeys: mockPrivateKeys.length,
};

// Mock store factory
export interface MockStoreConfig {
    auth?: Partial<AuthState>;
    users?: Partial<UsersState>;
    wallets?: Partial<WalletState>;
    user?: Partial<UserState>;
    theme?: Partial<ThemeState>;
    privateKeys?: Partial<PrivateKeysState>;
}

export const createMockStore = (config: MockStoreConfig = {}) => {
    const store = configureStore({
        reducer: {
            auth: authReducer,
            users: usersReducer,
            wallets: walletReducer,
            user: userReducer,
            theme: themeReducer,
            privateKeys: privateKeysReducer,
            [gauthApi.reducerPath]: gauthApi.reducer,
        },
        preloadedState: {
            auth: { ...defaultAuthState, ...config.auth },
            users: { ...defaultUsersState, ...config.users },
            wallets: { ...defaultWalletState, ...config.wallets },
            user: { ...defaultUserState, ...config.user },
            theme: { ...defaultThemeState, ...config.theme },
            privateKeys: { ...defaultPrivateKeysState, ...config.privateKeys },
            [gauthApi.reducerPath]: gauthApi.reducer,
        },
    });

    return store;
};

// Common test scenarios
export const mockStoreScenarios = {
    // Authenticated user with data
    authenticatedWithData: () => createMockStore({
        auth: { isAuthenticated: true, session: mockAuthSession },
        users: { users: mockUsers, loading: false, totalUsers: mockUsers.length },
    }),

    // Authenticated user with no data
    authenticatedNoData: () => createMockStore({
        auth: { isAuthenticated: true, session: mockAuthSession },
        users: { users: [], loading: false, totalUsers: 0 },
    }),

    // Loading state
    loading: () => createMockStore({
        auth: { isAuthenticated: true, session: mockAuthSession },
        users: { users: [], loading: true, totalUsers: 0 },
    }),

    // Error state
    error: () => createMockStore({
        auth: { isAuthenticated: true, session: mockAuthSession },
        users: { users: [], loading: false, error: 'Failed to fetch users', totalUsers: 0 },
    }),

    // Not authenticated
    notAuthenticated: () => createMockStore({
        auth: { isAuthenticated: false, session: null },
        users: { users: [], loading: false, totalUsers: 0 },
    }),

    // User with inactive status
    withInactiveUser: () => createMockStore({
        auth: { isAuthenticated: true, session: mockAuthSession },
        users: {
            users: [mockUsers[1]], // Second user is inactive
            loading: false,
            totalUsers: 1
        },
    }),
};

// Helper to get store state for assertions
export const getStoreState = (store: any) => store.getState();

// Helper to dispatch actions in tests
export const dispatchAction = (store: any, action: any) => store.dispatch(action);
