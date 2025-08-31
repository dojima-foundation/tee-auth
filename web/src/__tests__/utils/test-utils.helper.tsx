import React, { ReactElement } from 'react'
import { render, RenderOptions } from '@testing-library/react'
import { Provider } from 'react-redux'
import { ThemeProvider } from '@/components/ThemeProvider'
import { createMockStore, MockStoreConfig } from './mock-store-factory'

// Custom render function that includes providers
interface CustomRenderOptions extends Omit<RenderOptions, 'wrapper'> {
    storeConfig?: MockStoreConfig
    store?: any
}

export const renderWithProviders = (
    ui: ReactElement,
    {
        storeConfig = {},
        store = createMockStore(storeConfig),
        ...renderOptions
    }: CustomRenderOptions = {}
) => {
    const Wrapper = ({ children }: { children: React.ReactNode }) => (
        <Provider store={store}>
            <ThemeProvider>
                {children}
            </ThemeProvider>
        </Provider>
    )

    return {
        store,
        ...render(ui, { wrapper: Wrapper, ...renderOptions }),
    }
}

// Mock auth context
export const mockAuthContext = {
    loading: false,
    isAuthenticated: true,
    user: {
        id: '1',
        email: 'test@example.com',
        name: 'Test User',
    },
    signIn: jest.fn(),
    signOut: jest.fn(),
}

// Re-export mock data from factory
export {
    mockUsers,
    mockWallets,
    mockPrivateKeys,
    mockAuthUser,
    mockAuthSession,
    mockStoreScenarios
} from './mock-store-factory'

// Mock fetch for API calls
export const mockFetch = (data: any, status = 200) => {
    global.fetch = jest.fn(() =>
        Promise.resolve({
            ok: status >= 200 && status < 300,
            status,
            json: () => Promise.resolve(data),
            text: () => Promise.resolve(JSON.stringify(data)),
        })
    ) as jest.Mock
}

// Mock localStorage
export const mockLocalStorage = () => {
    const store: Record<string, string> = {}

    return {
        getItem: jest.fn((key: string) => store[key] || null),
        setItem: jest.fn((key: string, value: string) => {
            store[key] = value
        }),
        removeItem: jest.fn((key: string) => {
            delete store[key]
        }),
        clear: jest.fn(() => {
            Object.keys(store).forEach(key => delete store[key])
        }),
    }
}

// Mock sessionStorage
export const mockSessionStorage = () => {
    const store: Record<string, string> = {}

    return {
        getItem: jest.fn((key: string) => store[key] || null),
        setItem: jest.fn((key: string, value: string) => {
            store[key] = value
        }),
        removeItem: jest.fn((key: string) => {
            delete store[key]
        }),
        clear: jest.fn(() => {
            Object.keys(store).forEach(key => delete store[key])
        }),
    }
}

// Wait for async operations
export const waitFor = (ms: number) => new Promise(resolve => setTimeout(resolve, ms))

// Mock console methods to avoid noise in tests
export const mockConsole = () => {
    const originalConsole = { ...console }

    beforeEach(() => {
        console.log = jest.fn()
        console.warn = jest.fn()
        console.error = jest.fn()
    })

    afterEach(() => {
        Object.assign(console, originalConsole)
    })
}

// Re-export everything from testing library
export * from '@testing-library/react'
export { default as userEvent } from '@testing-library/user-event'
