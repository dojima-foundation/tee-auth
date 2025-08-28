import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export interface User {
    id: string
    name: string
    email: string
    avatar?: string
    role: 'admin' | 'user'
    createdAt: string
}

interface UserState {
    currentUser: User | null
    users: User[]
    loading: boolean
    error: string | null
}

const initialState: UserState = {
    currentUser: null,
    users: [],
    loading: false,
    error: null,
}

const userSlice = createSlice({
    name: 'user',
    initialState,
    reducers: {
        // Synchronous actions
        setCurrentUser: (state, action: PayloadAction<User | null>) => {
            state.currentUser = action.payload
            state.error = null
        },
        setUsers: (state, action: PayloadAction<User[]>) => {
            state.users = action.payload
            state.error = null
        },
        addUser: (state, action: PayloadAction<User>) => {
            state.users.push(action.payload)
        },
        updateUser: (state, action: PayloadAction<User>) => {
            const index = state.users.findIndex(user => user.id === action.payload.id)
            if (index !== -1) {
                state.users[index] = action.payload
            }
            if (state.currentUser?.id === action.payload.id) {
                state.currentUser = action.payload
            }
        },
        removeUser: (state, action: PayloadAction<string>) => {
            state.users = state.users.filter(user => user.id !== action.payload)
        },
        setLoading: (state, action: PayloadAction<boolean>) => {
            state.loading = action.payload
        },
        setError: (state, action: PayloadAction<string | null>) => {
            state.error = action.payload
            state.loading = false
        },
        clearError: (state) => {
            state.error = null
        },
    },
})

export const {
    setCurrentUser,
    setUsers,
    addUser,
    updateUser,
    removeUser,
    setLoading,
    setError,
    clearError,
} = userSlice.actions

export default userSlice.reducer
