import { createSlice, PayloadAction } from '@reduxjs/toolkit';

export interface AuthUser {
    id: string;
    organization_id: string;
    username: string;
    email: string;
    public_key?: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface AuthMethod {
    id: string;
    type: string;
    name: string;
}

export interface AuthSession {
    session_token: string;
    expires_at: string;
    user: AuthUser;
    auth_method: AuthMethod;
}

export interface AuthState {
    isAuthenticated: boolean;
    session: AuthSession | null;
    loading: boolean;
    error: string | null;
}

const initialState: AuthState = {
    isAuthenticated: false,
    session: null,
    loading: false,
    error: null,
};

const authSlice = createSlice({
    name: 'auth',
    initialState,
    reducers: {
        setAuthLoading: (state, action: PayloadAction<boolean>) => {
            state.loading = action.payload;
        },
        setAuthSession: (state, action: PayloadAction<AuthSession>) => {
            state.session = action.payload;
            state.isAuthenticated = true;
            state.loading = false;
            state.error = null;
        },
        setAuthError: (state, action: PayloadAction<string>) => {
            state.error = action.payload;
            state.loading = false;
        },
        clearAuth: (state) => {
            state.isAuthenticated = false;
            state.session = null;
            state.loading = false;
            state.error = null;
        },
        updateUser: (state, action: PayloadAction<Partial<AuthUser>>) => {
            if (state.session?.user) {
                state.session.user = { ...state.session.user, ...action.payload };
            }
        },
    },
});

export const {
    setAuthLoading,
    setAuthSession,
    setAuthError,
    clearAuth,
    updateUser,
} = authSlice.actions;

export default authSlice.reducer;

// Selectors
export const selectAuth = (state: { auth: AuthState }) => state.auth;
export const selectIsAuthenticated = (state: { auth: AuthState }) => state.auth.isAuthenticated;
export const selectAuthSession = (state: { auth: AuthState }) => state.auth.session;
export const selectAuthUser = (state: { auth: AuthState }) => state.auth.session?.user;
export const selectOrganizationId = (state: { auth: AuthState }) => state.auth.session?.user?.organization_id;
export const selectAuthLoading = (state: { auth: AuthState }) => state.auth.loading;
export const selectAuthError = (state: { auth: AuthState }) => state.auth.error;
