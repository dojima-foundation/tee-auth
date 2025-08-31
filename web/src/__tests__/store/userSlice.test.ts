import userReducer, {
    setCurrentUser,
    setUsers,
    addUser,
    updateUser,
    removeUser,
    setLoading,
    setError,
    clearError,
} from '@/store/slices/userSlice'
import type { User } from '@/store/slices/userSlice'

describe('userSlice', () => {
    const initialState = {
        currentUser: null,
        users: [],
        loading: false,
        error: null,
    }

    const mockUser: User = {
        id: '1',
        name: 'John Doe',
        email: 'john@example.com',
        avatar: 'avatar.jpg',
        role: 'user',
        createdAt: '2024-01-01T00:00:00Z',
    }

    const mockUsers: User[] = [
        mockUser,
        {
            id: '2',
            name: 'Jane Smith',
            email: 'jane@example.com',
            role: 'admin',
            createdAt: '2024-01-02T00:00:00Z',
        },
    ]

    it('should return the initial state', () => {
        expect(userReducer(undefined, { type: 'unknown' })).toEqual(initialState)
    })

    describe('setCurrentUser', () => {
        it('should set current user', () => {
            const result = userReducer(initialState, setCurrentUser(mockUser))
            expect(result.currentUser).toEqual(mockUser)
            expect(result.error).toBeNull()
        })

        it('should clear current user when set to null', () => {
            const stateWithUser = { ...initialState, currentUser: mockUser }
            const result = userReducer(stateWithUser, setCurrentUser(null))
            expect(result.currentUser).toBeNull()
        })
    })

    describe('setUsers', () => {
        it('should set users array', () => {
            const result = userReducer(initialState, setUsers(mockUsers))
            expect(result.users).toEqual(mockUsers)
            expect(result.error).toBeNull()
        })

        it('should clear error when setting users', () => {
            const stateWithError = { ...initialState, error: 'Some error' }
            const result = userReducer(stateWithError, setUsers(mockUsers))
            expect(result.error).toBeNull()
        })
    })

    describe('addUser', () => {
        it('should add user to users array', () => {
            const result = userReducer(initialState, addUser(mockUser))
            expect(result.users).toHaveLength(1)
            expect(result.users[0]).toEqual(mockUser)
        })

        it('should add user to existing users array', () => {
            const stateWithUsers = { ...initialState, users: [mockUsers[0]] }
            const result = userReducer(stateWithUsers, addUser(mockUsers[1]))
            expect(result.users).toHaveLength(2)
            expect(result.users[1]).toEqual(mockUsers[1])
        })
    })

    describe('updateUser', () => {
        it('should update user in users array', () => {
            const stateWithUsers = { ...initialState, users: mockUsers }
            const updatedUser = { ...mockUser, name: 'Updated Name' }

            const result = userReducer(stateWithUsers, updateUser(updatedUser))
            expect(result.users[0]).toEqual(updatedUser)
        })

        it('should update current user if it matches the updated user', () => {
            const stateWithCurrentUser = { ...initialState, currentUser: mockUser }
            const updatedUser = { ...mockUser, name: 'Updated Name' }

            const result = userReducer(stateWithCurrentUser, updateUser(updatedUser))
            expect(result.currentUser).toEqual(updatedUser)
        })

        it('should not update if user is not found', () => {
            const stateWithUsers = { ...initialState, users: mockUsers }
            const nonExistentUser = { ...mockUser, id: '999' }

            const result = userReducer(stateWithUsers, updateUser(nonExistentUser))
            expect(result.users).toEqual(mockUsers)
        })
    })

    describe('removeUser', () => {
        it('should remove user from users array', () => {
            const stateWithUsers = { ...initialState, users: mockUsers }
            const result = userReducer(stateWithUsers, removeUser('1'))

            expect(result.users).toHaveLength(1)
            expect(result.users[0].id).toBe('2')
        })

        it('should not remove if user is not found', () => {
            const stateWithUsers = { ...initialState, users: mockUsers }
            const result = userReducer(stateWithUsers, removeUser('999'))

            expect(result.users).toEqual(mockUsers)
        })
    })

    describe('setLoading', () => {
        it('should set loading to true', () => {
            const result = userReducer(initialState, setLoading(true))
            expect(result.loading).toBe(true)
        })

        it('should set loading to false', () => {
            const stateWithLoading = { ...initialState, loading: true }
            const result = userReducer(stateWithLoading, setLoading(false))
            expect(result.loading).toBe(false)
        })
    })

    describe('setError', () => {
        it('should set error and clear loading', () => {
            const stateWithLoading = { ...initialState, loading: true }
            const result = userReducer(stateWithLoading, setError('Something went wrong'))

            expect(result.error).toBe('Something went wrong')
            expect(result.loading).toBe(false)
        })

        it('should clear error when set to null', () => {
            const stateWithError = { ...initialState, error: 'Some error' }
            const result = userReducer(stateWithError, setError(null))

            expect(result.error).toBeNull()
        })
    })

    describe('clearError', () => {
        it('should clear error', () => {
            const stateWithError = { ...initialState, error: 'Some error' }
            const result = userReducer(stateWithError, clearError())

            expect(result.error).toBeNull()
        })
    })

    describe('complex scenarios', () => {
        it('should handle multiple operations in sequence', () => {
            let state = initialState

            // Set loading
            state = userReducer(state, setLoading(true))
            expect(state.loading).toBe(true)

            // Set users
            state = userReducer(state, setUsers(mockUsers))
            expect(state.users).toEqual(mockUsers)
            expect(state.loading).toBe(true) // Should not be affected

            // Add new user
            const newUser: User = {
                id: '3',
                name: 'New User',
                email: 'new@example.com',
                role: 'user',
                createdAt: '2024-01-03T00:00:00Z',
            }
            state = userReducer(state, addUser(newUser))
            expect(state.users).toHaveLength(3)

            // Update existing user
            const updatedUser = { ...mockUser, name: 'Updated John' }
            state = userReducer(state, updateUser(updatedUser))
            expect(state.users[0].name).toBe('Updated John')

            // Remove user
            state = userReducer(state, removeUser('2'))
            expect(state.users).toHaveLength(2)

            // Set error
            state = userReducer(state, setError('Operation failed'))
            expect(state.error).toBe('Operation failed')
            expect(state.loading).toBe(false)

            // Clear error
            state = userReducer(state, clearError())
            expect(state.error).toBeNull()
        })
    })
})
