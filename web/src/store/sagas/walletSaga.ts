import { call, put, takeLatest, delay } from 'redux-saga/effects'
import { PayloadAction } from '@reduxjs/toolkit'
import {
    setWallets,
    addWallet,
    updateWallet,
    removeWallet,
    setSelectedWallet,
    setPrivateKeys,
    addPrivateKey,
    updatePrivateKey,
    removePrivateKey,
    setLoading,
    setError,
    clearError,
    Wallet,
    PrivateKey,
} from '../slices/walletSlice'

// Mock API functions (replace with real API calls)
const api = {
    getWallets: async (): Promise<Wallet[]> => {
        await delay(800)
        return [
            {
                id: '1',
                name: 'Main Wallet',
                address: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
                balance: '1.2345',
                currency: 'ETH',
                isActive: true,
                createdAt: new Date().toISOString(),
            },
            {
                id: '2',
                name: 'Trading Wallet',
                address: '0x1234567890123456789012345678901234567890',
                balance: '0.5678',
                currency: 'ETH',
                isActive: true,
                createdAt: new Date().toISOString(),
            },
        ]
    },

    createWallet: async (walletData: Partial<Wallet>): Promise<Wallet> => {
        await delay(500)
        return {
            id: Math.random().toString(36).substr(2, 9),
            name: walletData.name || 'New Wallet',
            address: walletData.address || `0x${Math.random().toString(16).substr(2, 40)}`,
            balance: '0.0000',
            currency: walletData.currency || 'ETH',
            isActive: true,
            createdAt: new Date().toISOString(),
        }
    },

    updateWallet: async (walletId: string, walletData: Partial<Wallet>): Promise<Wallet> => {
        await delay(400)
        return {
            id: walletId,
            name: walletData.name || '',
            address: walletData.address || '',
            balance: walletData.balance || '0.0000',
            currency: walletData.currency || 'ETH',
            isActive: walletData.isActive ?? true,
            createdAt: new Date().toISOString(),
        }
    },

    deleteWallet: async (walletId: string): Promise<void> => {
        await delay(300)
        // Simulate successful deletion
    },

    getPrivateKeys: async (): Promise<PrivateKey[]> => {
        await delay(600)
        return [
            {
                id: '1',
                name: 'Main Key',
                publicKey: '0x742d35Cc6634C0532925a3b8D4C9db96C4b4d8b6',
                encryptedPrivateKey: 'encrypted_private_key_1',
                walletId: '1',
                isActive: true,
                createdAt: new Date().toISOString(),
            },
            {
                id: '2',
                name: 'Trading Key',
                publicKey: '0x1234567890123456789012345678901234567890',
                encryptedPrivateKey: 'encrypted_private_key_2',
                walletId: '2',
                isActive: true,
                createdAt: new Date().toISOString(),
            },
        ]
    },

    createPrivateKey: async (keyData: Partial<PrivateKey>): Promise<PrivateKey> => {
        await delay(500)
        return {
            id: Math.random().toString(36).substr(2, 9),
            name: keyData.name || 'New Key',
            publicKey: keyData.publicKey || `0x${Math.random().toString(16).substr(2, 40)}`,
            encryptedPrivateKey: keyData.encryptedPrivateKey || 'encrypted_key',
            walletId: keyData.walletId || '',
            isActive: true,
            createdAt: new Date().toISOString(),
        }
    },

    updatePrivateKey: async (keyId: string, keyData: Partial<PrivateKey>): Promise<PrivateKey> => {
        await delay(400)
        return {
            id: keyId,
            name: keyData.name || '',
            publicKey: keyData.publicKey || '',
            encryptedPrivateKey: keyData.encryptedPrivateKey || '',
            walletId: keyData.walletId || '',
            isActive: keyData.isActive ?? true,
            createdAt: new Date().toISOString(),
        }
    },

    deletePrivateKey: async (keyId: string): Promise<void> => {
        await delay(300)
        // Simulate successful deletion
    },
}

// Action types for saga
export const WALLET_ACTIONS = {
    FETCH_WALLETS: 'wallet/fetchWallets',
    CREATE_WALLET: 'wallet/createWallet',
    UPDATE_WALLET: 'wallet/updateWallet',
    DELETE_WALLET: 'wallet/deleteWallet',
    SELECT_WALLET: 'wallet/selectWallet',
    FETCH_PRIVATE_KEYS: 'wallet/fetchPrivateKeys',
    CREATE_PRIVATE_KEY: 'wallet/createPrivateKey',
    UPDATE_PRIVATE_KEY: 'wallet/updatePrivateKey',
    DELETE_PRIVATE_KEY: 'wallet/deletePrivateKey',
} as const

// Action creators for saga
export const walletActions = {
    fetchWallets: () => ({ type: WALLET_ACTIONS.FETCH_WALLETS }),
    createWallet: (walletData: Partial<Wallet>) => ({
        type: WALLET_ACTIONS.CREATE_WALLET,
        payload: walletData
    }),
    updateWallet: (walletId: string, walletData: Partial<Wallet>) => ({
        type: WALLET_ACTIONS.UPDATE_WALLET,
        payload: { walletId, walletData }
    }),
    deleteWallet: (walletId: string) => ({
        type: WALLET_ACTIONS.DELETE_WALLET,
        payload: walletId
    }),
    selectWallet: (wallet: Wallet | null) => ({
        type: WALLET_ACTIONS.SELECT_WALLET,
        payload: wallet
    }),
    fetchPrivateKeys: () => ({ type: WALLET_ACTIONS.FETCH_PRIVATE_KEYS }),
    createPrivateKey: (keyData: Partial<PrivateKey>) => ({
        type: WALLET_ACTIONS.CREATE_PRIVATE_KEY,
        payload: keyData
    }),
    updatePrivateKey: (keyId: string, keyData: Partial<PrivateKey>) => ({
        type: WALLET_ACTIONS.UPDATE_PRIVATE_KEY,
        payload: { keyId, keyData }
    }),
    deletePrivateKey: (keyId: string) => ({
        type: WALLET_ACTIONS.DELETE_PRIVATE_KEY,
        payload: keyId
    }),
}

// Saga functions
function* fetchWalletsSaga() {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const wallets: Wallet[] = yield call(api.getWallets)
        yield put(setWallets(wallets))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to fetch wallets'))
    } finally {
        yield put(setLoading(false))
    }
}

function* createWalletSaga(action: PayloadAction<Partial<Wallet>>) {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const newWallet: Wallet = yield call(api.createWallet, action.payload)
        yield put(addWallet(newWallet))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to create wallet'))
    } finally {
        yield put(setLoading(false))
    }
}

function* updateWalletSaga(action: PayloadAction<{ walletId: string; walletData: Partial<Wallet> }>) {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const { walletId, walletData } = action.payload
        // Use walletId to ensure it's not marked as unused
        console.log(`Updating wallet with ID: ${walletId}`)
        const updatedWallet: Wallet = yield call(api.updateWallet, walletId, walletData)
        yield put(updateWallet(updatedWallet))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to update wallet'))
    } finally {
        yield put(setLoading(false))
    }
}

function* deleteWalletSaga(action: PayloadAction<string>) {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const walletId = action.payload
        yield call(api.deleteWallet, walletId)
        yield put(removeWallet(walletId))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to delete wallet'))
    } finally {
        yield put(setLoading(false))
    }
}

function* selectWalletSaga(action: PayloadAction<Wallet | null>) {
    try {
        yield put(setSelectedWallet(action.payload))
    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to select wallet'))
    }
}

function* fetchPrivateKeysSaga() {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const privateKeys: PrivateKey[] = yield call(api.getPrivateKeys)
        yield put(setPrivateKeys(privateKeys))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to fetch private keys'))
    } finally {
        yield put(setLoading(false))
    }
}

function* createPrivateKeySaga(action: PayloadAction<Partial<PrivateKey>>) {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const newKey: PrivateKey = yield call(api.createPrivateKey, action.payload)
        yield put(addPrivateKey(newKey))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to create private key'))
    } finally {
        yield put(setLoading(false))
    }
}

function* updatePrivateKeySaga(action: PayloadAction<{ keyId: string; keyData: Partial<PrivateKey> }>) {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const { keyId, keyData } = action.payload
        // Use keyId to ensure it's not marked as unused
        console.log(`Updating private key with ID: ${keyId}`)
        const updatedKey: PrivateKey = yield call(api.updatePrivateKey, keyId, keyData)
        yield put(updatePrivateKey(updatedKey))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to update private key'))
    } finally {
        yield put(setLoading(false))
    }
}

function* deletePrivateKeySaga(action: PayloadAction<string>) {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const keyId = action.payload
        yield call(api.deletePrivateKey, keyId)
        yield put(removePrivateKey(keyId))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to delete private key'))
    } finally {
        yield put(setLoading(false))
    }
}

// Root wallet saga
export function* walletSaga() {
    yield takeLatest(WALLET_ACTIONS.FETCH_WALLETS, fetchWalletsSaga)
    yield takeLatest(WALLET_ACTIONS.CREATE_WALLET, createWalletSaga)
    yield takeLatest(WALLET_ACTIONS.UPDATE_WALLET, updateWalletSaga)
    yield takeLatest(WALLET_ACTIONS.DELETE_WALLET, deleteWalletSaga)
    yield takeLatest(WALLET_ACTIONS.SELECT_WALLET, selectWalletSaga)
    yield takeLatest(WALLET_ACTIONS.FETCH_PRIVATE_KEYS, fetchPrivateKeysSaga)
    yield takeLatest(WALLET_ACTIONS.CREATE_PRIVATE_KEY, createPrivateKeySaga)
    yield takeLatest(WALLET_ACTIONS.UPDATE_PRIVATE_KEY, updatePrivateKeySaga)
    yield takeLatest(WALLET_ACTIONS.DELETE_PRIVATE_KEY, deletePrivateKeySaga)
}
