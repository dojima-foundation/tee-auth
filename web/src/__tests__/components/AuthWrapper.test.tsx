import { render, screen } from '../utils/test-utils.helper';
import { AuthWrapper } from '@/components/AuthWrapper';
import { useAuth } from '@/lib/auth-context';

// Mock the auth context
jest.mock('@/lib/auth-context');
jest.mock('@/components/SessionLoading', () => ({
    SessionLoading: () => <div data-testid="session-loading">Loading...</div>
}));

const mockUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;

describe('AuthWrapper', () => {
    const TestChild = () => <div data-testid="test-child">Test Content</div>;

    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('shows loading component when authentication is loading', () => {
        mockUseAuth.mockReturnValue({
            loading: true,
            isAuthenticated: false,
            user: null,
            session: null,
            error: null,
            loginWithGoogle: jest.fn(),
            handleOAuthCallback: jest.fn(),
            setSession: jest.fn(),
            logout: jest.fn(),
            clearError: jest.fn(),
            refreshSession: jest.fn(),
            validateSession: jest.fn(),
            getSessionInfo: jest.fn(),
            listSessions: jest.fn(),
            destroySession: jest.fn(),
        });

        render(
            <AuthWrapper>
                <TestChild />
            </AuthWrapper>
        );

        expect(screen.getByTestId('session-loading')).toBeInTheDocument();
        expect(screen.queryByTestId('test-child')).not.toBeInTheDocument();
    });

    it('renders children when authentication is complete and user is authenticated', () => {
        mockUseAuth.mockReturnValue({
            loading: false,
            isAuthenticated: true,
            user: {
                id: 'user-123',
                organization_id: 'org-456',
                username: 'testuser',
                email: 'test@example.com',
                is_active: true,
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString(),
            },
            session: {
                session_token: 'token-123',
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
            error: null,
            loginWithGoogle: jest.fn(),
            handleOAuthCallback: jest.fn(),
            setSession: jest.fn(),
            logout: jest.fn(),
            clearError: jest.fn(),
            refreshSession: jest.fn(),
            validateSession: jest.fn(),
            getSessionInfo: jest.fn(),
            listSessions: jest.fn(),
            destroySession: jest.fn(),
        });

        render(
            <AuthWrapper>
                <TestChild />
            </AuthWrapper>
        );

        expect(screen.queryByTestId('session-loading')).not.toBeInTheDocument();
        expect(screen.getByTestId('test-child')).toBeInTheDocument();
        expect(screen.getByText('Test Content')).toBeInTheDocument();
    });

    it('renders children when authentication is complete but user is not authenticated', () => {
        mockUseAuth.mockReturnValue({
            loading: false,
            isAuthenticated: false,
            user: null,
            session: null,
            error: 'Authentication failed',
            loginWithGoogle: jest.fn(),
            handleOAuthCallback: jest.fn(),
            setSession: jest.fn(),
            logout: jest.fn(),
            clearError: jest.fn(),
            refreshSession: jest.fn(),
            validateSession: jest.fn(),
            getSessionInfo: jest.fn(),
            listSessions: jest.fn(),
            destroySession: jest.fn(),
        });

        render(
            <AuthWrapper>
                <TestChild />
            </AuthWrapper>
        );

        expect(screen.queryByTestId('session-loading')).not.toBeInTheDocument();
        expect(screen.getByTestId('test-child')).toBeInTheDocument();
    });

    it('handles multiple children correctly', () => {
        mockUseAuth.mockReturnValue({
            loading: false,
            isAuthenticated: true,
            user: {
                id: 'user-123',
                organization_id: 'org-456',
                username: 'testuser',
                email: 'test@example.com',
                is_active: true,
                created_at: new Date().toISOString(),
                updated_at: new Date().toISOString(),
            },
            session: {
                session_token: 'token-123',
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
            error: null,
            loginWithGoogle: jest.fn(),
            handleOAuthCallback: jest.fn(),
            setSession: jest.fn(),
            logout: jest.fn(),
            clearError: jest.fn(),
            refreshSession: jest.fn(),
            validateSession: jest.fn(),
            getSessionInfo: jest.fn(),
            listSessions: jest.fn(),
            destroySession: jest.fn(),
        });

        render(
            <AuthWrapper>
                <div data-testid="child-1">Child 1</div>
                <div data-testid="child-2">Child 2</div>
                <TestChild />
            </AuthWrapper>
        );

        expect(screen.getByTestId('child-1')).toBeInTheDocument();
        expect(screen.getByTestId('child-2')).toBeInTheDocument();
        expect(screen.getByTestId('test-child')).toBeInTheDocument();
    });

    it('calls useAuth hook correctly', () => {
        mockUseAuth.mockReturnValue({
            loading: false,
            isAuthenticated: true,
            user: null,
            session: null,
            error: null,
            loginWithGoogle: jest.fn(),
            handleOAuthCallback: jest.fn(),
            setSession: jest.fn(),
            logout: jest.fn(),
            clearError: jest.fn(),
            refreshSession: jest.fn(),
            validateSession: jest.fn(),
            getSessionInfo: jest.fn(),
            listSessions: jest.fn(),
            destroySession: jest.fn(),
        });

        render(
            <AuthWrapper>
                <TestChild />
            </AuthWrapper>
        );

        expect(mockUseAuth).toHaveBeenCalledTimes(1);
    });
});
