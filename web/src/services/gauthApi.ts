// GAuth API Service
const GAUTH_BASE_URL = process.env.NEXT_PUBLIC_GAUTH_API_URL || 'http://localhost:8082';

export interface CreateOrganizationRequest {
    name: string;
    initial_user_email: string;
    initial_user_public_key?: string;
}

export interface CreateOrganizationResponse {
    success: boolean;
    data: {
        organization: {
            id: string;
            name: string;
            version: string;
            created_at: string;
            updated_at: string;
        };
        status: string;
        user_id: string;
    };
    message?: string;
}

export interface Organization {
    id: string;
    name: string;
    version: string;
    created_at: string;
    updated_at: string;
}

export interface User {
    id: string;
    organization_id: string;
    username: string;
    email: string;
    public_key?: string;
    tags: string[];
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface ListOrganizationsResponse {
    success: boolean;
    data: {
        organizations: Organization[];
        next_page_token?: string;
    };
}

export interface GetOrganizationResponse {
    success: boolean;
    data: {
        organization: Organization;
    };
}

export interface CreateUserRequest {
    username: string;
    email: string;
    public_key?: string;
    tags?: string[];
}

export interface CreateUserResponse {
    success: boolean;
    data: {
        user: User;
    };
}

export interface ListUsersResponse {
    success: boolean;
    data: {
        users: User[];
        next_page_token?: string;
    };
}

// Google OAuth Types
export interface GoogleOAuthLoginRequest {
    organization_id: string;
    state?: string;
}

export interface GoogleOAuthLoginResponse {
    success: boolean;
    data: {
        auth_url: string;
        state: string;
    };
}

export interface GoogleOAuthCallbackResponse {
    success: boolean;
    data: {
        session_token: string;
        expires_at: string;
        user: {
            id: string;
            organization_id: string;
            username: string;
            email: string;
            public_key?: string;
            is_active: boolean;
            created_at: string;
            updated_at: string;
        };
        auth_method: {
            id: string;
            type: string;
            name: string;
        };
    };
}

export interface GoogleOAuthRefreshResponse {
    success: boolean;
    data: {
        message: string;
    };
}

// Wallet Types
export interface Wallet {
    id: string;
    organization_id: string;
    name: string;
    public_key: string;
    seed_phrase?: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface CreateWalletRequest {
    name: string;
    accounts: CreateWalletAccount[];
    mnemonic_length?: number;
    tags?: string[];
}

export interface CreateWalletAccount {
    curve: string;
    path_format: string;
    path: string;
    address_format: string;
}

export interface CreateWalletResponse {
    success: boolean;
    data: {
        wallet: Wallet;
    };
}

export interface ListWalletsResponse {
    success: boolean;
    data: {
        wallets: Wallet[];
        next_page_token?: string;
    };
}

// Private Key Types
export interface PrivateKey {
    id: string;
    organization_id: string;
    wallet_id: string;
    name: string;
    public_key: string;
    curve: string;
    path: string;
    tags: string[];
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface CreatePrivateKeyRequest {
    wallet_id: string;
    name: string;
    curve: string;
    private_key_material?: string;
    tags?: string[];
}

export interface CreatePrivateKeyResponse {
    success: boolean;
    data: {
        private_key: PrivateKey;
    };
}

export interface ListPrivateKeysResponse {
    success: boolean;
    data: {
        private_keys: PrivateKey[];
        next_page_token?: string;
    };
}

// Session Management Types
export interface SessionInfo {
    session_id: string;
    user_id: string;
    organization_id: string;
    email: string;
    role: string;
    oauth_provider?: string;
    created_at: string;
    last_activity: string;
    expires_at: string;
}

export interface SessionInfoResponse {
    success: boolean;
    data: SessionInfo;
}

export interface SessionRefreshResponse {
    success: boolean;
    data: {
        message: string;
        expires_at: string;
    };
}

export interface SessionLogoutResponse {
    success: boolean;
    data: {
        message: string;
    };
}

export interface SessionListResponse {
    success: boolean;
    data: {
        sessions: SessionInfo[];
        count: number;
    };
}

export interface SessionValidateResponse {
    success: boolean;
    data: SessionInfo;
}

export interface ApiError {
    success: false;
    error: string;
    message: string;
}

class GAuthApiService {
    private baseUrl: string;

    constructor(baseUrl: string = GAUTH_BASE_URL) {
        this.baseUrl = baseUrl;
    }

    private async makeRequest<T>(
        endpoint: string,
        options: RequestInit = {}
    ): Promise<T> {
        const url = `${this.baseUrl}${endpoint}`;

        // Get session token from localStorage for authenticated requests
        const sessionToken = localStorage.getItem('gauth_session_token');

        const defaultOptions: RequestInit = {
            headers: {
                'Content-Type': 'application/json',
                ...(sessionToken && { 'Authorization': `Bearer ${sessionToken}` }),
                ...options.headers,
            },
            credentials: 'include', // Include cookies for session management
        };

        const response = await fetch(url, {
            ...defaultOptions,
            ...options,
        });

        if (!response.ok) {
            const errorData = await response.json().catch(() => ({}));
            throw new Error(errorData.message || `HTTP error! status: ${response.status}`);
        }

        return response.json();
    }

    async createOrganization(data: CreateOrganizationRequest): Promise<CreateOrganizationResponse> {
        return this.makeRequest<CreateOrganizationResponse>('/api/v1/organizations', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async getOrganizations(): Promise<ListOrganizationsResponse> {
        return this.makeRequest<ListOrganizationsResponse>('/api/v1/organizations');
    }

    async getOrganization(id: string): Promise<GetOrganizationResponse> {
        return this.makeRequest<GetOrganizationResponse>(`/api/v1/organizations/${id}`);
    }

    async createUser(data: CreateUserRequest): Promise<CreateUserResponse> {
        return this.makeRequest<CreateUserResponse>('/api/v1/users', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async getUsers(): Promise<ListUsersResponse> {
        // Get organization_id from session data
        const sessionData = localStorage.getItem('gauth_session_data');
        if (!sessionData) {
            throw new Error('No session data found');
        }

        const session = JSON.parse(sessionData);
        const organizationId = session.user?.organization_id;

        if (!organizationId) {
            throw new Error('Organization ID not found in session');
        }

        return this.makeRequest<ListUsersResponse>(`/api/v1/users?organization_id=${encodeURIComponent(organizationId)}`);
    }

    // Google OAuth Methods
    async initiateGoogleOAuth(data: GoogleOAuthLoginRequest): Promise<GoogleOAuthLoginResponse> {
        return this.makeRequest<GoogleOAuthLoginResponse>('/api/v1/auth/google/login', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async handleGoogleOAuthCallback(code: string, state: string): Promise<GoogleOAuthCallbackResponse> {
        const params = new URLSearchParams({ code, state });
        return this.makeRequest<GoogleOAuthCallbackResponse>(`/api/v1/auth/google/callback?${params}`);
    }

    async refreshGoogleOAuthToken(authMethodId: string): Promise<GoogleOAuthRefreshResponse> {
        return this.makeRequest<GoogleOAuthRefreshResponse>(`/api/v1/auth/google/refresh/${authMethodId}`, {
            method: 'POST',
        });
    }

    // Wallet Methods
    async createWallet(data: CreateWalletRequest): Promise<CreateWalletResponse> {
        return this.makeRequest<CreateWalletResponse>('/api/v1/wallets', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async getWallets(): Promise<ListWalletsResponse> {
        // Get organization_id from session data
        const sessionData = localStorage.getItem('gauth_session_data');
        if (!sessionData) {
            throw new Error('No session data found');
        }

        const session = JSON.parse(sessionData);
        const organizationId = session.user?.organization_id;

        if (!organizationId) {
            throw new Error('Organization ID not found in session');
        }

        return this.makeRequest<ListWalletsResponse>(`/api/v1/wallets?organization_id=${encodeURIComponent(organizationId)}`);
    }

    // Private Key Methods
    async createPrivateKey(data: CreatePrivateKeyRequest): Promise<CreatePrivateKeyResponse> {
        return this.makeRequest<CreatePrivateKeyResponse>('/api/v1/private-keys', {
            method: 'POST',
            body: JSON.stringify(data),
        });
    }

    async getPrivateKeys(): Promise<ListPrivateKeysResponse> {
        // Get organization_id from session data
        const sessionData = localStorage.getItem('gauth_session_data');
        if (!sessionData) {
            throw new Error('No session data found');
        }

        const session = JSON.parse(sessionData);
        const organizationId = session.user?.organization_id;

        if (!organizationId) {
            throw new Error('Organization ID not found in session');
        }

        return this.makeRequest<ListPrivateKeysResponse>(`/api/v1/private-keys?organization_id=${encodeURIComponent(organizationId)}`);
    }

    // Session Management Methods
    async getSessionInfo(): Promise<SessionInfoResponse> {
        return this.makeRequest<SessionInfoResponse>('/api/v1/sessions/info');
    }

    async refreshSession(): Promise<SessionRefreshResponse> {
        return this.makeRequest<SessionRefreshResponse>('/api/v1/sessions/refresh', {
            method: 'POST',
        });
    }

    async logoutSession(): Promise<SessionLogoutResponse> {
        return this.makeRequest<SessionLogoutResponse>('/api/v1/sessions/logout', {
            method: 'POST',
        });
    }

    async validateSession(): Promise<SessionValidateResponse> {
        return this.makeRequest<SessionValidateResponse>('/api/v1/sessions/validate');
    }

    async listSessions(): Promise<SessionListResponse> {
        return this.makeRequest<SessionListResponse>('/api/v1/sessions/list');
    }

    async destroySession(sessionId: string): Promise<SessionLogoutResponse> {
        return this.makeRequest<SessionLogoutResponse>(`/api/v1/sessions/${sessionId}`, {
            method: 'DELETE',
        });
    }
}

export const gauthApi = new GAuthApiService();
