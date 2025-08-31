import { render, screen, waitFor } from '../utils/test-utils.helper';
import { ProtectedRoute } from '@/components/ProtectedRoute';
import { useAuth } from '@/lib/auth-context';
import { useRouter } from 'next/navigation';

// Mock the auth context and router
jest.mock('@/lib/auth-context');
jest.mock('next/navigation', () => ({
    useRouter: jest.fn(),
}));

const mockUseAuth = useAuth as jest.MockedFunction<typeof useAuth>;
const mockUseRouter = useRouter as jest.MockedFunction<typeof useRouter>;
const mockPush = jest.fn();

describe('ProtectedRoute', () => {
    const TestChild = () => <div data-testid="protected-content">Protected Content</div>;
    const CustomFallback = () => <div data-testid="custom-fallback">Custom Loading...</div>;

    beforeEach(() => {
        jest.clearAllMocks();
        mockUseRouter.mockReturnValue({
            push: mockPush,
            replace: jest.fn(),
            prefetch: jest.fn(),
            back: jest.fn(),
            forward: jest.fn(),
            refresh: jest.fn(),
        });
    });

    it('shows loading state when authentication is loading', () => {
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
            <ProtectedRoute>
                <TestChild />
            </ProtectedRoute>
        );

        expect(screen.getByText('Loading...')).toBeInTheDocument();
        expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument();
    });

    it('shows custom fallback when authentication is loading', () => {
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
            <ProtectedRoute fallback={<CustomFallback />}>
                <TestChild />
            </ProtectedRoute>
        );

        expect(screen.getByTestId('custom-fallback')).toBeInTheDocument();
        expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument();
    });

    it('renders children when user is authenticated', () => {
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
            <ProtectedRoute>
                <TestChild />
            </ProtectedRoute>
        );

        expect(screen.getByTestId('protected-content')).toBeInTheDocument();
        expect(screen.getByText('Protected Content')).toBeInTheDocument();
    });

    it('redirects to signin when user is not authenticated', async () => {
        mockUseAuth.mockReturnValue({
            loading: false,
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
            <ProtectedRoute>
                <TestChild />
            </ProtectedRoute>
        );

        await waitFor(() => {
            expect(mockPush).toHaveBeenCalledWith('/auth/signin');
        });

        expect(screen.queryByTestId('protected-content')).not.toBeInTheDocument();
    });

    it('does not redirect when authentication is still loading', () => {
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
            <ProtectedRoute>
                <TestChild />
            </ProtectedRoute>
        );

        expect(mockPush).not.toHaveBeenCalled();
    });

    it('handles authentication state changes correctly', async () => {
        // Start with loading state
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

        const { rerender } = render(
            <ProtectedRoute>
                <TestChild />
            </ProtectedRoute>
        );

        expect(screen.getByText('Loading...')).toBeInTheDocument();

        // Change to authenticated state
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

        rerender(
            <ProtectedRoute>
                <TestChild />
            </ProtectedRoute>
        );

        expect(screen.getByTestId('protected-content')).toBeInTheDocument();
        expect(screen.queryByText('Loading...')).not.toBeInTheDocument();
    });

    it('renders null when not authenticated (will redirect)', () => {
        mockUseAuth.mockReturnValue({
            loading: false,
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

        const { container } = render(
            <ProtectedRoute>
                <TestChild />
            </ProtectedRoute>
        );

        // Should render null (empty container) when not authenticated
        expect(container.firstChild).toBeNull();
    });
});
