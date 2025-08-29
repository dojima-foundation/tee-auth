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
    user: AuthUser | null;
    session: AuthSession | null;
    loading: boolean;
    error: string | null;
}

export interface LoginCredentials {
    organization_id: string;
    state?: string;
}
