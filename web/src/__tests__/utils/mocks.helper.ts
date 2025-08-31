// Mock Next.js modules
jest.mock('next/navigation', () => ({
    useRouter: () => ({
        push: jest.fn(),
        replace: jest.fn(),
        prefetch: jest.fn(),
        back: jest.fn(),
        forward: jest.fn(),
        refresh: jest.fn(),
    }),
    useSearchParams: () => new URLSearchParams(),
    usePathname: () => '/',
}))

jest.mock('next/router', () => ({
    useRouter: () => ({
        route: '/',
        pathname: '/',
        query: {},
        asPath: '/',
        push: jest.fn(),
        pop: jest.fn(),
        reload: jest.fn(),
        back: jest.fn(),
        prefetch: jest.fn().mockResolvedValue(undefined),
        beforePopState: jest.fn(),
        events: {
            on: jest.fn(),
            off: jest.fn(),
            emit: jest.fn(),
        },
        isFallback: false,
    }),
}))

// Mock auth context
jest.mock('@/lib/auth-context', () => ({
    useAuth: () => ({
        loading: false,
        isAuthenticated: true,
        user: {
            id: '1',
            email: 'test@example.com',
            name: 'Test User',
        },
        signIn: jest.fn(),
        signOut: jest.fn(),
    }),
}))

// Mock API service
jest.mock('@/services/gauthApi', () => ({
    gauthApi: {
        getUsers: jest.fn(),
        createUser: jest.fn(),
        deleteUser: jest.fn(),
        getWallets: jest.fn(),
        createWallet: jest.fn(),
        deleteWallet: jest.fn(),
        getPrivateKeys: jest.fn(),
        createPrivateKey: jest.fn(),
        deletePrivateKey: jest.fn(),
    },
}))

// Mock Redux store
jest.mock('@/store', () => ({
    getStore: () => ({
        getState: () => ({
            user: { user: null, loading: false, error: null },
            theme: { theme: 'light' },
            wallet: { wallets: [], loading: false, error: null },
            auth: { isAuthenticated: false, loading: false, error: null },
            users: { users: [], loading: false, error: null },
            wallets: { wallets: [], loading: false, error: null },
            privateKeys: { privateKeys: [], loading: false, error: null },
        }),
        dispatch: jest.fn(),
        subscribe: jest.fn(),
    }),
}))

// Mock window.matchMedia
Object.defineProperty(window, 'matchMedia', {
    writable: true,
    value: jest.fn().mockImplementation(query => ({
        matches: false,
        media: query,
        onchange: null,
        addListener: jest.fn(),
        removeListener: jest.fn(),
        addEventListener: jest.fn(),
        removeEventListener: jest.fn(),
        dispatchEvent: jest.fn(),
    })),
})

// Mock IntersectionObserver
global.IntersectionObserver = class IntersectionObserver {
    constructor() { }
    disconnect() { }
    observe() { }
    unobserve() { }
}

// Mock ResizeObserver
global.ResizeObserver = class ResizeObserver {
    constructor() { }
    disconnect() { }
    observe() { }
    unobserve() { }
}

// Mock fetch
global.fetch = jest.fn()

// Mock localStorage
const localStorageMock = {
    getItem: jest.fn(),
    setItem: jest.fn(),
    removeItem: jest.fn(),
    clear: jest.fn(),
}
Object.defineProperty(window, 'localStorage', {
    value: localStorageMock,
})

// Mock sessionStorage
const sessionStorageMock = {
    getItem: jest.fn(),
    setItem: jest.fn(),
    removeItem: jest.fn(),
    clear: jest.fn(),
}
Object.defineProperty(window, 'sessionStorage', {
    value: sessionStorageMock,
})




