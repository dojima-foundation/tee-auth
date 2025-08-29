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
    organization_id: string;
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

        const defaultOptions: RequestInit = {
            headers: {
                'Content-Type': 'application/json',
                ...options.headers,
            },
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

    async getUsers(organizationId?: string): Promise<ListUsersResponse> {
        const params = organizationId ? `?organization_id=${organizationId}` : '';
        return this.makeRequest<ListUsersResponse>(`/api/v1/users${params}`);
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
}

export const gauthApi = new GAuthApiService();
