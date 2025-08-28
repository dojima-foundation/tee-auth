import { call, put, takeLatest, select, delay } from 'redux-saga/effects'
import { PayloadAction } from '@reduxjs/toolkit'
import { RootState } from '../index'
import {
    setTheme,
    setSystemTheme,
    setLoading,
    setError,
    clearError,
    Theme,
} from '../slices/themeSlice'

// Mock API functions (replace with real API calls)
const api = {
    saveThemePreference: async (theme: Theme): Promise<void> => {
        // Simulate API call to save theme preference
        await delay(300)
        localStorage.setItem('theme-preference', theme)
    },

    getThemePreference: async (): Promise<Theme> => {
        // Simulate API call to get theme preference
        await delay(200)
        return (localStorage.getItem('theme-preference') as Theme) || 'system'
    },
}

// Action types for saga
export const THEME_ACTIONS = {
    INITIALIZE_THEME: 'theme/initializeTheme',
    SAVE_THEME: 'theme/saveTheme',
    DETECT_SYSTEM_THEME: 'theme/detectSystemTheme',
} as const

// Action creators for saga
export const themeActions = {
    initializeTheme: () => ({ type: THEME_ACTIONS.INITIALIZE_THEME }),
    saveTheme: (theme: Theme) => ({
        type: THEME_ACTIONS.SAVE_THEME,
        payload: theme
    }),
    detectSystemTheme: () => ({ type: THEME_ACTIONS.DETECT_SYSTEM_THEME }),
}

// Helper function to detect system theme
function detectSystemTheme(): 'light' | 'dark' {
    if (typeof window !== 'undefined') {
        return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'
    }
    return 'light'
}

// Helper function to apply theme to DOM
function applyThemeToDOM(theme: Theme) {
    if (typeof window !== 'undefined') {
        const root = window.document.documentElement
        root.classList.remove('light', 'dark')

        if (theme === 'system') {
            const systemTheme = detectSystemTheme()
            root.classList.add(systemTheme)
        } else {
            root.classList.add(theme)
        }
    }
}

// Saga functions
function* initializeThemeSaga() {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        // Get saved theme preference
        const savedTheme: Theme = yield call(api.getThemePreference)
        yield put(setTheme(savedTheme))

        // Detect system theme
        const systemTheme = detectSystemTheme()
        yield put(setSystemTheme(systemTheme))

        // Apply theme to DOM
        applyThemeToDOM(savedTheme)

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to initialize theme'))
    } finally {
        yield put(setLoading(false))
    }
}

function* saveThemeSaga(action: PayloadAction<Theme>) {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const theme = action.payload

        // Save theme preference
        yield call(api.saveThemePreference, theme)

        // Update state
        yield put(setTheme(theme))

        // Apply theme to DOM
        applyThemeToDOM(theme)

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to save theme'))
    } finally {
        yield put(setLoading(false))
    }
}

function* detectSystemThemeSaga() {
    try {
        const systemTheme = detectSystemTheme()
        yield put(setSystemTheme(systemTheme))

        // Get current theme from state
        const currentTheme: Theme = yield select((state: RootState) => state.theme.currentTheme)

        // If current theme is 'system', apply the detected theme
        if (currentTheme === 'system') {
            applyThemeToDOM('system')
        }

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to detect system theme'))
    }
}

// Root theme saga
export function* themeSaga() {
    yield takeLatest(THEME_ACTIONS.INITIALIZE_THEME, initializeThemeSaga)
    yield takeLatest(THEME_ACTIONS.SAVE_THEME, saveThemeSaga)
    yield takeLatest(THEME_ACTIONS.DETECT_SYSTEM_THEME, detectSystemThemeSaga)
}
