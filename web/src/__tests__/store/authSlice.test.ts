import authReducer, {
    setAuthLoading,
    setAuthSession,
    setAuthError,
    clearAuth,
    updateUser,
    selectAuth,
    selectIsAuthenticated,
    selectAuthSession,
    selectAuthUser,
    selectOrganizationId,
    selectAuthLoading,
    selectAuthError,
} from '@/store/authSlice'
import type { AuthSession, AuthUser, AuthMethod } from '@/store/authSlice'

describe('authSlice', () => {
    const initialState = {
        isAuthenticated: false,
        session: null,
        loading: false,
        error: null,
    }

    const mockUser: AuthUser = {
        id: '1',
        organization_id: 'org-1',
        username: 'testuser',
        email: 'test@example.com',
        public_key: 'public-key-123',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
    }

    const mockAuthMethod: AuthMethod = {
        id: 'auth-1',
        type: 'google',
        name: 'Google OAuth',
    }

    const mockSession: AuthSession = {
        session_token: 'session-token-123',
        expires_at: '2024-12-31T23:59:59Z',
        user: mockUser,
        auth_method: mockAuthMethod,
    }

    it('should return the initial state', () => {
        expect(authReducer(undefined, { type: 'unknown' })).toEqual(initialState)
    })

    it('should handle setAuthLoading', () => {
        const previousState = { ...initialState }

        expect(authReducer(previousState, setAuthLoading(true))).toEqual({
            ...previousState,
            loading: true,
        })

        expect(authReducer(previousState, setAuthLoading(false))).toEqual({
            ...previousState,
            loading: false,
        })
    })

    it('should handle setAuthSession', () => {
        const previousState = { ...initialState, loading: true }

        expect(authReducer(previousState, setAuthSession(mockSession))).toEqual({
            isAuthenticated: true,
            session: mockSession,
            loading: false,
            error: null,
        })
    })

    it('should handle setAuthError', () => {
        const previousState = { ...initialState, loading: true }
        const errorMessage = 'Authentication failed'

        expect(authReducer(previousState, setAuthError(errorMessage))).toEqual({
            ...previousState,
            error: errorMessage,
            loading: false,
        })
    })

    it('should handle clearAuth', () => {
        const previousState = {
            isAuthenticated: true,
            session: mockSession,
            loading: false,
            error: null,
        }

        expect(authReducer(previousState, clearAuth())).toEqual(initialState)
    })

    it('should handle updateUser', () => {
        const previousState = {
            isAuthenticated: true,
            session: mockSession,
            loading: false,
            error: null,
        }

        const userUpdates = {
            username: 'updateduser',
            email: 'updated@example.com',
        }

        const expectedState = {
            ...previousState,
            session: {
                ...mockSession,
                user: {
                    ...mockUser,
                    ...userUpdates,
                },
            },
        }

        expect(authReducer(previousState, updateUser(userUpdates))).toEqual(expectedState)
    })

    it('should not update user when session is null', () => {
        const previousState = { ...initialState }
        const userUpdates = { username: 'updateduser' }

        expect(authReducer(previousState, updateUser(userUpdates))).toEqual(previousState)
    })

    it('should handle multiple actions in sequence', () => {
        let state = initialState

        // Set loading
        state = authReducer(state, setAuthLoading(true))
        expect(state.loading).toBe(true)

        // Set session
        state = authReducer(state, setAuthSession(mockSession))
        expect(state.isAuthenticated).toBe(true)
        expect(state.session).toEqual(mockSession)
        expect(state.loading).toBe(false)

        // Update user
        state = authReducer(state, updateUser({ username: 'newusername' }))
        expect(state.session?.user.username).toBe('newusername')

        // Clear auth
        state = authReducer(state, clearAuth())
        expect(state).toEqual(initialState)
    })
})

describe('authSlice selectors', () => {
    const mockState = {
        auth: {
            isAuthenticated: true,
            session: {
                session_token: 'token',
                expires_at: '2024-12-31T23:59:59Z',
                user: {
                    id: '1',
                    organization_id: 'org-1',
                    username: 'testuser',
                    email: 'test@example.com',
                    is_active: true,
                    created_at: '2024-01-01T00:00:00Z',
                    updated_at: '2024-01-01T00:00:00Z',
                },
                auth_method: {
                    id: 'auth-1',
                    type: 'google',
                    name: 'Google OAuth',
                },
            },
            loading: false,
            error: null,
        },
    }

    it('should select auth state', () => {
        expect(selectAuth(mockState)).toEqual(mockState.auth)
    })

    it('should select isAuthenticated', () => {
        expect(selectIsAuthenticated(mockState)).toBe(true)
    })

    it('should select auth session', () => {
        expect(selectAuthSession(mockState)).toEqual(mockState.auth.session)
    })

    it('should select auth user', () => {
        expect(selectAuthUser(mockState)).toEqual(mockState.auth.session?.user)
    })

    it('should select organization id', () => {
        expect(selectOrganizationId(mockState)).toBe('org-1')
    })

    it('should select auth loading', () => {
        expect(selectAuthLoading(mockState)).toBe(false)
    })

    it('should select auth error', () => {
        expect(selectAuthError(mockState)).toBe(null)
    })
})
