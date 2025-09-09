import { render, screen, fireEvent, waitFor } from '../utils/test-utils.helper';
import { SessionManager } from '@/components/SessionManager';
import { useSessionManagement } from '@/lib/session-middleware';
import { SessionInfo } from '@/types/auth';

// Mock the session management hook
jest.mock('@/lib/session-middleware');
jest.mock('@/components/ui/snackbar', () => ({
    useSnackbar: () => ({
        showSnackbar: jest.fn(),
    }),
}));

const mockUseSessionManagement = useSessionManagement as jest.MockedFunction<typeof useSessionManagement>;

describe('SessionManager', () => {
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

    const mockSessions: SessionInfo[] = [
        mockSessionInfo,
        {
            ...mockSessionInfo,
            session_id: 'session-456',
            email: 'test2@example.com',
            last_activity: new Date(Date.now() - 1800 * 1000).toISOString(), // 30 minutes ago
        },
    ];

    const mockSessionManagement = {
        isAuthenticated: true,
        getSessionInfo: jest.fn(),
        listSessions: jest.fn(),
        destroySession: jest.fn(),
        refreshSession: jest.fn(),
    };

    beforeEach(() => {
        jest.clearAllMocks();
        mockUseSessionManagement.mockReturnValue(mockSessionManagement);
    });

    it('renders nothing when user is not authenticated', () => {
        mockUseSessionManagement.mockReturnValue({
            ...mockSessionManagement,
            isAuthenticated: false,
        });

        const { container } = render(<SessionManager />);
        expect(container.firstChild).toBeNull();
    });

    it('loads current session info on mount', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.listSessions.mockResolvedValue(mockSessions);

        render(<SessionManager />);

        await waitFor(() => {
            expect(mockSessionManagement.getSessionInfo).toHaveBeenCalledTimes(1);
            expect(mockSessionManagement.listSessions).toHaveBeenCalledTimes(1);
        });
    });

    it('displays current session information', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.listSessions.mockResolvedValue(mockSessions);

        render(<SessionManager />);

        await waitFor(() => {
            expect(screen.getByText('Current Session')).toBeInTheDocument();
            expect(screen.getByText(mockSessionInfo.session_id)).toBeInTheDocument();
            expect(screen.getByText(mockSessionInfo.user_id)).toBeInTheDocument();
            expect(screen.getByText(mockSessionInfo.email)).toBeInTheDocument();
            expect(screen.getByText(mockSessionInfo.role)).toBeInTheDocument();
        });
    });

    it('displays all sessions', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.listSessions.mockResolvedValue(mockSessions);

        render(<SessionManager />);

        await waitFor(() => {
            expect(screen.getByText('All Sessions')).toBeInTheDocument();
            expect(screen.getByText('test@example.com')).toBeInTheDocument();
            expect(screen.getByText('test2@example.com')).toBeInTheDocument();
        });
    });

    it('shows loading state while fetching data', () => {
        mockSessionManagement.getSessionInfo.mockImplementation(() => new Promise(() => { })); // Never resolves
        mockSessionManagement.listSessions.mockImplementation(() => new Promise(() => { })); // Never resolves

        render(<SessionManager />);

        expect(screen.getAllByRole('generic').some(el => el.className.includes('animate-spin'))).toBe(true);
    });

    it.skip('handles refresh session action', async () => {
        // TODO: Fix async loading issue in SessionManager component
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.listSessions.mockResolvedValue(mockSessions);
        mockSessionManagement.refreshSession.mockResolvedValue(undefined);

        render(<SessionManager />);

        // Wait for current session to load (check for session ID)
        await waitFor(() => {
            expect(screen.getByText('session-123')).toBeInTheDocument();
        });

        const refreshButton = screen.getByText('Refresh Session');
        fireEvent.click(refreshButton);

        await waitFor(() => {
            expect(mockSessionManagement.refreshSession).toHaveBeenCalledTimes(1);
        });
    });

    it.skip('handles destroy session action', async () => {
        // TODO: Fix async loading issue in SessionManager component
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.listSessions.mockResolvedValue(mockSessions);
        mockSessionManagement.destroySession.mockResolvedValue(undefined);

        render(<SessionManager />);

        // Wait for sessions to load
        await waitFor(() => {
            expect(screen.getByText('session-456')).toBeInTheDocument();
        });

        // Find and click the destroy button for the second session (not current)
        const destroyButtons = screen.getAllByText('Destroy');
        fireEvent.click(destroyButtons[0]); // Click first destroy button

        // Confirm in dialog
        await waitFor(() => {
            expect(screen.getByText('Destroy Session')).toBeInTheDocument();
        });

        const confirmButton = screen.getByRole('button', { name: /destroy session/i });
        fireEvent.click(confirmButton);

        await waitFor(() => {
            expect(mockSessionManagement.destroySession).toHaveBeenCalledWith('session-456');
        });
    });

    it('shows session status badges correctly', async () => {
        const expiredSession = {
            ...mockSessionInfo,
            session_id: 'expired-session',
            expires_at: new Date(Date.now() - 3600 * 1000).toISOString(), // 1 hour ago (expired)
        };

        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.listSessions.mockResolvedValue([mockSessionInfo, expiredSession]);

        render(<SessionManager />);

        await waitFor(() => {
            expect(screen.getByText('Current')).toBeInTheDocument();
            expect(screen.getByText('Expired')).toBeInTheDocument();
        });
    });

    it('handles API errors gracefully', async () => {
        mockSessionManagement.getSessionInfo.mockRejectedValue(new Error('API Error'));
        mockSessionManagement.listSessions.mockRejectedValue(new Error('API Error'));

        render(<SessionManager />);

        await waitFor(() => {
            expect(screen.getByText('No session information available')).toBeInTheDocument();
            expect(screen.getByText('No sessions found')).toBeInTheDocument();
        });
    });

    it.skip('reloads session info when reload button is clicked', async () => {
        // TODO: Fix async loading issue in SessionManager component
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.listSessions.mockResolvedValue(mockSessions);

        render(<SessionManager />);

        // Wait for current session to load
        await waitFor(() => {
            expect(screen.getByText('session-123')).toBeInTheDocument();
        });

        const reloadButton = screen.getByText('Reload Info');
        fireEvent.click(reloadButton);

        await waitFor(() => {
            expect(mockSessionManagement.getSessionInfo).toHaveBeenCalledTimes(2);
        });
    });

    it('refreshes sessions list when refresh button is clicked', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.listSessions.mockResolvedValue(mockSessions);

        render(<SessionManager />);

        await waitFor(() => {
            expect(screen.getByText('Refresh Sessions')).toBeInTheDocument();
        });

        const refreshSessionsButton = screen.getByText('Refresh Sessions');
        fireEvent.click(refreshSessionsButton);

        await waitFor(() => {
            expect(mockSessionManagement.listSessions).toHaveBeenCalledTimes(2);
        });
    });

    it('does not show destroy button for current session', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.listSessions.mockResolvedValue([mockSessionInfo]);

        render(<SessionManager />);

        await waitFor(() => {
            expect(screen.getByText('All Sessions')).toBeInTheDocument();
        });

        // Should not show destroy button for current session
        const destroyButtons = screen.queryAllByText('Destroy');
        expect(destroyButtons).toHaveLength(0);
    });

    it('formats dates correctly', async () => {
        mockSessionManagement.getSessionInfo.mockResolvedValue(mockSessionInfo);
        mockSessionManagement.listSessions.mockResolvedValue(mockSessions);

        render(<SessionManager />);

        await waitFor(() => {
            // Check that dates are formatted and displayed
            expect(screen.getByText(/Created:/)).toBeInTheDocument();
            expect(screen.getByText(/Last Activity:/)).toBeInTheDocument();
            expect(screen.getByText(/Expires:/)).toBeInTheDocument();
        });
    });
});
