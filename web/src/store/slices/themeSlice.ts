import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export type Theme = 'light' | 'dark' | 'system'

interface ThemeState {
    currentTheme: Theme
    systemTheme: 'light' | 'dark'
    loading: boolean
    error: string | null
}

const initialState: ThemeState = {
    currentTheme: 'system',
    systemTheme: 'light',
    loading: false,
    error: null,
}

const themeSlice = createSlice({
    name: 'theme',
    initialState,
    reducers: {
        setTheme: (state, action: PayloadAction<Theme>) => {
            state.currentTheme = action.payload
            state.error = null
        },
        setSystemTheme: (state, action: PayloadAction<'light' | 'dark'>) => {
            state.systemTheme = action.payload
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
    setTheme,
    setSystemTheme,
    setLoading,
    setError,
    clearError,
} = themeSlice.actions

export default themeSlice.reducer
