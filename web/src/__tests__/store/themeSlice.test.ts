import themeReducer, { setTheme, setSystemTheme, setLoading, setError, clearError } from '@/store/slices/themeSlice'

describe('themeSlice', () => {
    const initialState = {
        currentTheme: 'system' as const,
        systemTheme: 'light' as const,
        loading: false,
        error: null,
    }

    it('should return the initial state', () => {
        expect(themeReducer(undefined, { type: 'unknown' })).toEqual(initialState)
    })

    it('should handle setTheme', () => {
        const previousState = { ...initialState }
        const newTheme = 'dark' as const

        expect(themeReducer(previousState, setTheme(newTheme))).toEqual({
            ...previousState,
            currentTheme: newTheme,
            error: null,
        })
    })

    it('should handle setSystemTheme', () => {
        const previousState = { ...initialState }
        const newSystemTheme = 'dark' as const

        expect(themeReducer(previousState, setSystemTheme(newSystemTheme))).toEqual({
            ...previousState,
            systemTheme: newSystemTheme,
        })
    })

    it('should handle setLoading', () => {
        const previousState = { ...initialState }

        expect(themeReducer(previousState, setLoading(true))).toEqual({
            ...previousState,
            loading: true,
        })

        expect(themeReducer(previousState, setLoading(false))).toEqual({
            ...previousState,
            loading: false,
        })
    })

    it('should handle setError', () => {
        const previousState = { ...initialState, loading: true }
        const errorMessage = 'Theme loading failed'

        expect(themeReducer(previousState, setError(errorMessage))).toEqual({
            ...previousState,
            error: errorMessage,
            loading: false,
        })
    })

    it('should handle clearError', () => {
        const previousState = { ...initialState, error: 'Some error' }

        expect(themeReducer(previousState, clearError())).toEqual({
            ...previousState,
            error: null,
        })
    })

    it('should handle setError with null', () => {
        const previousState = { ...initialState, loading: true }

        expect(themeReducer(previousState, setError(null))).toEqual({
            ...previousState,
            error: null,
            loading: false,
        })
    })

    it('should handle multiple actions in sequence', () => {
        let state = initialState

        // Set loading
        state = themeReducer(state, setLoading(true))
        expect(state.loading).toBe(true)

        // Set theme
        state = themeReducer(state, setTheme('dark'))
        expect(state.currentTheme).toBe('dark')
        expect(state.loading).toBe(true) // Should not be affected by setTheme

        // Set error
        state = themeReducer(state, setError('Failed to load theme'))
        expect(state.error).toBe('Failed to load theme')
        expect(state.loading).toBe(false) // Should be set to false by setError

        // Clear error
        state = themeReducer(state, clearError())
        expect(state.error).toBe(null)
    })
})
