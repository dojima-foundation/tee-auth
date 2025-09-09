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
    createMockUser: (overrides: Partial<MockUser> = {}): MockUser => ({
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
    createMockWallet: (overrides: Partial<MockWallet> = {}): MockWallet => ({
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
    createMockPrivateKey: (overrides: Partial<MockPrivateKey> = {}): MockPrivateKey => ({
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

// Type definitions for mock data
interface MockUser {
    id: string;
    email: string;
    name: string;
    organization_id: string;
    username: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

interface MockWallet {
    id: string;
    name: string;
    address: string;
    balance: string;
    currency: string;
    isActive: boolean;
    createdAt: string;
}

interface MockPrivateKey {
    id: string;
    name: string;
    publicKey: string;
    encryptedPrivateKey: string;
    walletId: string;
    isActive: boolean;
    createdAt: string;
}

// Extend Jest matchers using module augmentation
declare module '@jest/expect' {
    interface Matchers<R> {
        toHaveFocus(): R;
        toBeInTheDocument(): R;
        toHaveClass(...classNames: string[]): R;
        toHaveAttribute(attr: string, value?: string): R;
    }
}

// Global test utilities interface
declare global {
    var testUtils: {
        waitFor: (ms: number) => Promise<void>;
        createMockUser: (overrides?: Partial<MockUser>) => MockUser;
        createMockWallet: (overrides?: Partial<MockWallet>) => MockWallet;
        createMockPrivateKey: (overrides?: Partial<MockPrivateKey>) => MockPrivateKey;
    };
}
