import { gauthApi } from '@/services/gauthApi';
import type {
    CreateOrganizationRequest,
    CreateUserRequest,
    CreateWalletRequest,
    CreatePrivateKeyRequest,
    GoogleOAuthLoginRequest,
} from '@/services/gauthApi';

// Mock fetch globally
global.fetch = jest.fn();
const mockFetch = fetch as jest.MockedFunction<typeof fetch>;

// Mock localStorage
const mockLocalStorage = {
    getItem: jest.fn(),
    setItem: jest.fn(),
    removeItem: jest.fn(),
    clear: jest.fn(),
};
Object.defineProperty(window, 'localStorage', {
    value: mockLocalStorage,
});

describe('GAuthApiService', () => {
    const baseUrl = 'http://localhost:8082';
    const mockSessionToken = 'mock-session-token';

    beforeEach(() => {
        jest.clearAllMocks();
        mockLocalStorage.getItem.mockReturnValue(mockSessionToken);
    });

    describe('Organization Management', () => {
        it('creates organization successfully', async () => {
            const orgData: CreateOrganizationRequest = {
                name: 'Test Organization',
                initial_user_email: 'admin@test.com',
                initial_user_public_key: '0x1234567890abcdef',
            };

            const mockResponse = {
                success: true,
                data: {
                    organization: {
                        id: 'org-123',
                        name: 'Test Organization',
                        version: '1.0.0',
                        created_at: new Date().toISOString(),
                        updated_at: new Date().toISOString(),
                    },
                    status: 'created',
                    user_id: 'user-123',
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.createOrganization(orgData);

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/organizations`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer mock-session-token',
                },
                credentials: 'include',
                body: JSON.stringify(orgData),
            });

            expect(result).toEqual(mockResponse);
        });

        it('gets organizations successfully', async () => {
            const mockResponse = {
                success: true,
                data: {
                    organizations: [
                        {
                            id: 'org-123',
                            name: 'Test Organization',
                            version: '1.0.0',
                            created_at: new Date().toISOString(),
                            updated_at: new Date().toISOString(),
                        },
                    ],
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.getOrganizations();

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/organizations`, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer mock-session-token',
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });

        it('gets specific organization successfully', async () => {
            const orgId = 'org-123';
            const mockResponse = {
                success: true,
                data: {
                    organization: {
                        id: orgId,
                        name: 'Test Organization',
                        version: '1.0.0',
                        created_at: new Date().toISOString(),
                        updated_at: new Date().toISOString(),
                    },
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.getOrganization(orgId);

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/organizations/${orgId}`, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer mock-session-token',
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });
    });

    describe('User Management', () => {
        it('creates user successfully', async () => {
            const userData: CreateUserRequest = {
                username: 'testuser',
                email: 'test@example.com',
                public_key: '0x1234567890abcdef',
                tags: ['admin'],
            };

            const mockResponse = {
                success: true,
                data: {
                    user: {
                        id: 'user-123',
                        organization_id: 'org-456',
                        username: 'testuser',
                        email: 'test@example.com',
                        public_key: '0x1234567890abcdef',
                        tags: ['admin'],
                        is_active: true,
                        created_at: new Date().toISOString(),
                        updated_at: new Date().toISOString(),
                    },
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.createUser(userData);

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/users`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
                body: JSON.stringify(userData),
            });

            expect(result).toEqual(mockResponse);
        });

        it('gets users successfully', async () => {
            const mockResponse = {
                success: true,
                data: {
                    users: [
                        {
                            id: 'user-123',
                            organization_id: 'org-456',
                            username: 'testuser',
                            email: 'test@example.com',
                            public_key: '0x1234567890abcdef',
                            tags: ['admin'],
                            is_active: true,
                            created_at: new Date().toISOString(),
                            updated_at: new Date().toISOString(),
                        },
                    ],
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.getUsers();

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/users`, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });
    });

    describe('Google OAuth', () => {
        it('initiates Google OAuth successfully', async () => {
            const oauthData: GoogleOAuthLoginRequest = {
                organization_id: 'org-123',
                state: 'random-state',
            };

            const mockResponse = {
                success: true,
                data: {
                    auth_url: 'https://accounts.google.com/oauth/authorize?client_id=...',
                    state: 'random-state',
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.initiateGoogleOAuth(oauthData);

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/auth/google/login`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': 'Bearer mock-session-token',
                },
                credentials: 'include',
                body: JSON.stringify(oauthData),
            });

            expect(result).toEqual(mockResponse);
        });

        it('handles Google OAuth callback successfully', async () => {
            const code = 'auth-code';
            const state = 'random-state';

            const mockResponse = {
                success: true,
                data: {
                    session_token: 'new-session-token',
                    expires_at: new Date(Date.now() + 3600 * 1000).toISOString(),
                    user: {
                        id: 'user-123',
                        organization_id: 'org-456',
                        username: 'testuser',
                        email: 'test@example.com',
                        is_active: true,
                        created_at: new Date().toISOString(),
                        updated_at: new Date().toISOString(),
                    },
                    auth_method: {
                        id: 'auth-1',
                        type: 'oauth',
                        name: 'Google',
                    },
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.handleGoogleOAuthCallback(code, state);

            expect(mockFetch).toHaveBeenCalledWith(
                `${baseUrl}/api/v1/auth/google/callback?code=${code}&state=${state}`,
                {
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer mock-session-token',
                    },
                    credentials: 'include',
                }
            );

            expect(result).toEqual(mockResponse);
        });

        it('refreshes Google OAuth token successfully', async () => {
            const authMethodId = 'auth-123';

            const mockResponse = {
                success: true,
                data: {
                    message: 'Token refreshed successfully',
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.refreshGoogleOAuthToken(authMethodId);

            expect(mockFetch).toHaveBeenCalledWith(
                `${baseUrl}/api/v1/auth/google/refresh/${authMethodId}`,
                {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': 'Bearer mock-session-token',
                    },
                    credentials: 'include',
                }
            );

            expect(result).toEqual(mockResponse);
        });
    });

    describe('Wallet Management', () => {
        it('creates wallet successfully', async () => {
            const walletData: CreateWalletRequest = {
                name: 'Test Wallet',
                accounts: [
                    {
                        curve: 'secp256k1',
                        path_format: 'bip44',
                        path: "m/44'/60'/0'/0/0",
                        address_format: 'ethereum',
                    },
                ],
                mnemonic_length: 12,
                tags: ['main'],
            };

            const mockResponse = {
                success: true,
                data: {
                    wallet: {
                        id: 'wallet-123',
                        organization_id: 'org-456',
                        name: 'Test Wallet',
                        public_key: '0x1234567890abcdef',
                        seed_phrase: 'test seed phrase',
                        is_active: true,
                        created_at: new Date().toISOString(),
                        updated_at: new Date().toISOString(),
                    },
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.createWallet(walletData);

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/wallets`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
                body: JSON.stringify(walletData),
            });

            expect(result).toEqual(mockResponse);
        });

        it('gets wallets successfully', async () => {
            const mockResponse = {
                success: true,
                data: {
                    wallets: [
                        {
                            id: 'wallet-123',
                            organization_id: 'org-456',
                            name: 'Test Wallet',
                            public_key: '0x1234567890abcdef',
                            is_active: true,
                            created_at: new Date().toISOString(),
                            updated_at: new Date().toISOString(),
                        },
                    ],
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.getWallets();

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/wallets`, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });
    });

    describe('Private Key Management', () => {
        it('creates private key successfully', async () => {
            const privateKeyData: CreatePrivateKeyRequest = {
                wallet_id: 'wallet-123',
                name: 'Test Private Key',
                curve: 'secp256k1',
                private_key_material: 'test-private-key',
                tags: ['main'],
            };

            const mockResponse = {
                success: true,
                data: {
                    private_key: {
                        id: 'pkey-123',
                        organization_id: 'org-456',
                        wallet_id: 'wallet-123',
                        name: 'Test Private Key',
                        public_key: '0x1234567890abcdef',
                        curve: 'secp256k1',
                        path: "m/44'/60'/0'/0/0",
                        tags: ['main'],
                        is_active: true,
                        created_at: new Date().toISOString(),
                        updated_at: new Date().toISOString(),
                    },
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.createPrivateKey(privateKeyData);

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/private-keys`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
                body: JSON.stringify(privateKeyData),
            });

            expect(result).toEqual(mockResponse);
        });

        it('gets private keys successfully', async () => {
            const mockResponse = {
                success: true,
                data: {
                    private_keys: [
                        {
                            id: 'pkey-123',
                            organization_id: 'org-456',
                            wallet_id: 'wallet-123',
                            name: 'Test Private Key',
                            public_key: '0x1234567890abcdef',
                            curve: 'secp256k1',
                            path: "m/44'/60'/0'/0/0",
                            tags: ['main'],
                            is_active: true,
                            created_at: new Date().toISOString(),
                            updated_at: new Date().toISOString(),
                        },
                    ],
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.getPrivateKeys();

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/private-keys`, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });
    });

    describe('Session Management', () => {
        it('gets session info successfully', async () => {
            const mockResponse = {
                success: true,
                data: {
                    session_id: 'session-123',
                    user_id: 'user-123',
                    organization_id: 'org-456',
                    email: 'test@example.com',
                    role: 'admin',
                    oauth_provider: 'google',
                    created_at: new Date().toISOString(),
                    last_activity: new Date().toISOString(),
                    expires_at: new Date(Date.now() + 3600 * 1000).toISOString(),
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.getSessionInfo();

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/sessions/info`, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });

        it('refreshes session successfully', async () => {
            const mockResponse = {
                success: true,
                data: {
                    message: 'Session refreshed',
                    expires_at: new Date(Date.now() + 3600 * 1000).toISOString(),
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.refreshSession();

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/sessions/refresh`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });

        it('logs out session successfully', async () => {
            const mockResponse = {
                success: true,
                data: {
                    message: 'Logged out successfully',
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.logoutSession();

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/sessions/logout`, {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });

        it('validates session successfully', async () => {
            const mockResponse = {
                success: true,
                data: {
                    session_id: 'session-123',
                    user_id: 'user-123',
                    organization_id: 'org-456',
                    email: 'test@example.com',
                    role: 'admin',
                    oauth_provider: 'google',
                    created_at: new Date().toISOString(),
                    last_activity: new Date().toISOString(),
                    expires_at: new Date(Date.now() + 3600 * 1000).toISOString(),
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.validateSession();

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/sessions/validate`, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });

        it('lists sessions successfully', async () => {
            const mockResponse = {
                success: true,
                data: {
                    sessions: [
                        {
                            session_id: 'session-123',
                            user_id: 'user-123',
                            organization_id: 'org-456',
                            email: 'test@example.com',
                            role: 'admin',
                            oauth_provider: 'google',
                            created_at: new Date().toISOString(),
                            last_activity: new Date().toISOString(),
                            expires_at: new Date(Date.now() + 3600 * 1000).toISOString(),
                        },
                    ],
                    count: 1,
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.listSessions();

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/sessions/list`, {
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });

        it('destroys session successfully', async () => {
            const sessionId = 'session-123';
            const mockResponse = {
                success: true,
                data: {
                    message: 'Session destroyed successfully',
                },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.destroySession(sessionId);

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/sessions/${sessionId}`, {
                method: 'DELETE',
                headers: {
                    'Content-Type': 'application/json',
                    'Authorization': `Bearer ${mockSessionToken}`,
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });
    });

    describe('Error Handling', () => {
        it('handles HTTP errors correctly', async () => {
            const errorResponse = {
                success: false,
                error: 'Bad Request',
                message: 'Invalid request data',
            };

            mockFetch.mockResolvedValueOnce({
                ok: false,
                status: 400,
                json: async () => errorResponse,
            } as Response);

            await expect(gauthApi.getUsers()).rejects.toThrow('Invalid request data');
        });

        it('handles network errors correctly', async () => {
            mockFetch.mockRejectedValueOnce(new Error('Network error'));

            await expect(gauthApi.getUsers()).rejects.toThrow('Network error');
        });

        it('handles JSON parsing errors correctly', async () => {
            mockFetch.mockResolvedValueOnce({
                ok: false,
                status: 500,
                json: async () => {
                    throw new Error('Invalid JSON');
                },
            } as Response);

            await expect(gauthApi.getUsers()).rejects.toThrow('HTTP error! status: 500');
        });

        it('works without session token for unauthenticated requests', async () => {
            mockLocalStorage.getItem.mockReturnValue(null);

            const mockResponse = {
                success: true,
                data: { organizations: [] },
            };

            mockFetch.mockResolvedValueOnce({
                ok: true,
                json: async () => mockResponse,
            } as Response);

            const result = await gauthApi.getOrganizations();

            expect(mockFetch).toHaveBeenCalledWith(`${baseUrl}/api/v1/organizations`, {
                headers: {
                    'Content-Type': 'application/json',
                },
                credentials: 'include',
            });

            expect(result).toEqual(mockResponse);
        });
    });
});
