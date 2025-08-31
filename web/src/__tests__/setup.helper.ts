/**
 * @jest-environment jsdom
 */

// Global test setup file
import '@testing-library/jest-dom'

// Mock console methods to reduce noise in tests
const originalConsole = { ...console }

beforeAll(() => {
    // Suppress console warnings and errors in tests unless explicitly needed
    console.warn = jest.fn()
    console.error = jest.fn()
    console.log = jest.fn()
})

afterAll(() => {
    // Restore original console methods
    Object.assign(console, originalConsole)
})

// Global test utilities
global.testUtils = {
    // Helper to wait for async operations
    waitFor: (ms: number) => new Promise(resolve => setTimeout(resolve, ms)),

    // Helper to create mock user data
    createMockUser: (overrides = {}) => ({
        id: '1',
        email: 'test@example.com',
        name: 'Test User',
        organization_id: 'org-1',
        username: 'testuser',
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
        ...overrides,
    }),

    // Helper to create mock wallet data
    createMockWallet: (overrides = {}) => ({
        id: '1',
        name: 'Test Wallet',
        address: '0x1234567890abcdef',
        balance: '1.5',
        currency: 'ETH',
        isActive: true,
        createdAt: '2024-01-01T00:00:00Z',
        ...overrides,
    }),

    // Helper to create mock private key data
    createMockPrivateKey: (overrides = {}) => ({
        id: '1',
        name: 'Test Private Key',
        publicKey: '0xpublickey123',
        encryptedPrivateKey: 'encrypted_data',
        walletId: '1',
        isActive: true,
        createdAt: '2024-01-01T00:00:00Z',
        ...overrides,
    }),
}

// Extend Jest matchers
declare global {
    namespace jest {
        interface Matchers<R> {
            toHaveFocus(): R
            toBeInTheDocument(): R
            toHaveClass(...classNames: string[]): R
            toHaveAttribute(attr: string, value?: string): R
        }
    }

    var testUtils: {
        waitFor: (ms: number) => Promise<void>
        createMockUser: (overrides?: any) => any
        createMockWallet: (overrides?: any) => any
        createMockPrivateKey: (overrides?: any) => any
    }
}
