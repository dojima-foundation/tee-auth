import { renderWithProviders, render, screen, fireEvent, waitFor, act } from '../utils/test-utils.helper';
import { SessionStatus } from '@/components/SessionStatus';
import { useSessionManagement } from '@/lib/session-middleware';
import { SessionInfo } from '@/types/auth';

// Mock the session management hook
jest.mock('@/lib/session-middleware');
jest.mock('@/components/ui/snackbar', () => ({
    useSnackbar: () => ({
        showSnackbar: jest.fn(),
    }),
}));

// Mock the auth context
jest.mock('@/lib/auth-context', () => ({
    useAuth: () => ({
        isAuthenticated: true,
        validateSession: jest.fn().mockResolvedValue(true),
        refreshSession: jest.fn().mockResolvedValue(undefined),
        getSessionInfo: jest.fn().mockResolvedValue({
            session_id: 'session-123',
            user_id: 'user-123',
            organization_id: 'org-456',
            email: 'test@example.com',
            role: 'admin',
            oauth_provider: 'google',
            created_at: new Date(Date.now() - 3600 * 1000).toISOString(),
            last_activity: new Date(Date.now() - 300 * 1000).toISOString(),
            expires_at: new Date(Date.now() + 3600 * 1000).toISOString(),
        }),
        listSessions: jest.fn().mockResolvedValue([]),
        destroySession: jest.fn().mockResolvedValue(undefined),
        logout: jest.fn().mockResolvedValue(undefined),
    }),
}));

const mockUseSessionManagement = useSessionManagement as jest.MockedFunction<typeof useSessionManagement>;

describe('SessionStatus', () => {
    const mockSessionInfo: SessionInfo = {
        session_id: 'session-123',
        user_id: 'user-123',
        organization_id: 'org-456',
        email: 'test@example.com',
        role: 'admin',
        oauth_provider: 'google',
        created_at: new Date(Date.now() - 3600 * 1000).toISOString(), // 1 hour ago
        last_activity: new Date(Date.now() - 300 * 1000).toISOString(), // 5 minutes ago
        expires_at: new Date(Date.now() + 3600 * 1000).toISOString(), // 1 hour from now
    };

    const mockSessionManagement = {
        isAuthenticated: true,
        getSessionInfo: jest.fn().mockResolvedValue(mockSessionInfo),
        refreshSession: jest.fn().mockResolvedValue(undefined),
    };

    beforeEach(() => {
        jest.clearAllMocks();
        mockUseSessionManagement.mockReturnValue(mockSessionManagement);
        // Ensure the mock returns the session info
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
    });

    it('renders nothing when user is not authenticated', () => {
        mockUseSessionManagement.mockReturnValue({
            ...mockSessionManagement,
            isAuthenticated: false,
        });

        const { container } = renderWithProviders(<SessionStatus />);
        expect(container.firstChild).toBeNull();
    });

    it('renders nothing when session info is not available', () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(null);

        const { container } = renderWithProviders(<SessionStatus />);
        expect(container.firstChild).toBeNull();
    });

    it('loads session info on mount when authenticated', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);

        renderWithProviders(<SessionStatus />);

        await waitFor(() => {
            expect(mockSessionManagement.getSessionInfo).toHaveBeenCalledTimes(1);
        });
    });

    it('displays session status badge with time until expiry', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);

        renderWithProviders(<SessionStatus />);

        await waitFor(() => {
            expect(screen.getByText(/Expires in/)).toBeInTheDocument();
        });
    });

    it('shows expired status when session is expired', async () => {
        const expiredSession = {
            ...mockSessionInfo,
            expires_at: new Date(Date.now() - 3600 * 1000).toISOString(), // 1 hour ago (expired)
        };

        mockSessionManagement.getSessionInfo.mockResolvedValue(expiredSession);

        renderWithProviders(<SessionStatus />);

        await waitFor(() => {
            expect(screen.getByText('Expired')).toBeInTheDocument();
        });
    });

    it('shows refresh button when session is not expired', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);

        renderWithProviders(<SessionStatus />);

        await waitFor(() => {
            expect(screen.getByText('↻')).toBeInTheDocument();
        });
    });

    it('does not show refresh button when session is expired', async () => {
        const expiredSession = {
            ...mockSessionInfo,
            expires_at: new Date(Date.now() - 3600 * 1000).toISOString(), // 1 hour ago (expired)
        };

        mockSessionManagement.getSessionInfo.mockResolvedValue(expiredSession);

        renderWithProviders(<SessionStatus />);

        await waitFor(() => {
            expect(screen.queryByText('↻')).not.toBeInTheDocument();
        });
    });

    it('handles refresh button click', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.refreshSession.mockResolvedValue(undefined);

        await act(async () => {
            renderWithProviders(<SessionStatus />);
        });

        await waitFor(() => {
            expect(screen.getByText('↻')).toBeInTheDocument();
        });

        const refreshButton = screen.getByText('↻');
        fireEvent.click(refreshButton);

        await waitFor(() => {
            expect(mockSessionManagement.refreshSession).toHaveBeenCalledTimes(1);
        });
    }, 10000);

    it('shows loading state while refreshing', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.refreshSession.mockImplementation(() => new Promise(() => { })); // Never resolves

        renderWithProviders(<SessionStatus />);

        await waitFor(() => {
            expect(screen.getByText('↻')).toBeInTheDocument();
        });

        const refreshButton = screen.getByText('↻');
        fireEvent.click(refreshButton);

        await waitFor(() => {
            expect(screen.getByRole('generic').className).toContain('animate-spin');
        });
    });

    it('displays detailed view when showDetails is true', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);

        render(<SessionStatus showDetails={true} />);

        await waitFor(() => {
            expect(screen.getByText(/Session:/)).toBeInTheDocument();
            expect(screen.getByText(/Last Activity:/)).toBeInTheDocument();
            expect(screen.getByText(/Expires:/)).toBeInTheDocument();
        });
    });

    it('shows OAuth provider badge in detailed view', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);

        render(<SessionStatus showDetails={true} />);

        await waitFor(() => {
            expect(screen.getByText('google')).toBeInTheDocument();
        });
    });

    it('does not show OAuth provider badge when not available', async () => {
        const sessionWithoutProvider = {
            ...mockSessionInfo,
            oauth_provider: undefined,
        };

        mockSessionManagement.getSessionInfo.mockResolvedValue(sessionWithoutProvider);

        render(<SessionStatus showDetails={true} />);

        await waitFor(() => {
            expect(screen.queryByText('google')).not.toBeInTheDocument();
        });
    });

    it('applies custom className', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);

        const { container } = render(<SessionStatus className="custom-class" />);

        await waitFor(() => {
            expect(container.firstChild).toHaveClass('custom-class');
        });
    });

    it('handles API errors gracefully', async () => {
        mockSessionManagement.getSessionInfo.mockRejectedValue(new Error('API Error'));

        const { container } = renderWithProviders(<SessionStatus />);

        await waitFor(() => {
            expect(container.firstChild).toBeNull();
        });
    });

    it('auto-refreshes session info every minute', async () => {
        // Mock setInterval to call the callback immediately
        const originalSetInterval = global.setInterval;
        const mockSetInterval = jest.fn((callback, _delay) => {
            // Call the callback immediately for testing
            setTimeout(callback, 10);
            return 1; // Return a mock interval ID
        });
        global.setInterval = mockSetInterval;

        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);

        await act(async () => {
            renderWithProviders(<SessionStatus />);
        });

        await waitFor(() => {
            expect(mockSessionManagement.getSessionInfo).toHaveBeenCalledTimes(1);
        });

        // Verify setInterval was called
        expect(mockSetInterval).toHaveBeenCalledWith(expect.any(Function), 60000);

        // Restore original setInterval
        global.setInterval = originalSetInterval;
    });

    it('clears session info when authentication status changes to false', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);

        const { rerender } = renderWithProviders(<SessionStatus />);

        await waitFor(() => {
            expect(screen.getByText(/Expires in/)).toBeInTheDocument();
        });

        // Change authentication status to false
        mockUseSessionManagement.mockReturnValue({
            ...mockSessionManagement,
            isAuthenticated: false,
        });

        await act(async () => {
            rerender(<SessionStatus />);
        });

        await waitFor(() => {
            expect(screen.queryByText(/Expires in/)).not.toBeInTheDocument();
        });
    });

    it('calculates time until expiry correctly for different time ranges', async () => {
        // Test with 2 hours remaining
        const session2Hours = {
            ...mockSessionInfo,
            expires_at: new Date(Date.now() + 2 * 60 * 60 * 1000).toISOString(),
        };

        mockSessionManagement.getSessionInfo.mockResolvedValue(session2Hours);

        await act(async () => {
            renderWithProviders(<SessionStatus />);
        });

        await waitFor(() => {
            expect(screen.getByText(/Expires in 2h/)).toBeInTheDocument();
        });
    });

    it('shows warning color for sessions expiring soon', async () => {
        // Test with 15 minutes remaining (should show warning)
        const session15Minutes = {
            ...mockSessionInfo,
            expires_at: new Date(Date.now() + 15 * 60 * 1000).toISOString(),
        };

        mockSessionManagement.getSessionInfo.mockResolvedValue(session15Minutes);

        await act(async () => {
            renderWithProviders(<SessionStatus />);
        });

        await waitFor(() => {
            const badge = screen.getByText(/Expires in 15m/);
            expect(badge).toBeInTheDocument();
        });
    });
});
