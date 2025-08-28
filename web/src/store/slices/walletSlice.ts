import { createSlice, PayloadAction } from '@reduxjs/toolkit'

export interface Wallet {
    id: string
    name: string
    address: string
    balance: string
    currency: string
    isActive: boolean
    createdAt: string
}

export interface PrivateKey {
    id: string
    name: string
    publicKey: string
    encryptedPrivateKey: string
    walletId: string
    isActive: boolean
    createdAt: string
}

interface WalletState {
    wallets: Wallet[]
    privateKeys: PrivateKey[]
    selectedWallet: Wallet | null
    loading: boolean
    error: string | null
}

const initialState: WalletState = {
    wallets: [],
    privateKeys: [],
    selectedWallet: null,
    loading: false,
    error: null,
}

const walletSlice = createSlice({
    name: 'wallet',
    initialState,
    reducers: {
        // Wallet actions
        setWallets: (state, action: PayloadAction<Wallet[]>) => {
            state.wallets = action.payload
            state.error = null
        },
        addWallet: (state, action: PayloadAction<Wallet>) => {
            state.wallets.push(action.payload)
        },
        updateWallet: (state, action: PayloadAction<Wallet>) => {
            const index = state.wallets.findIndex(wallet => wallet.id === action.payload.id)
            if (index !== -1) {
                state.wallets[index] = action.payload
            }
            if (state.selectedWallet?.id === action.payload.id) {
                state.selectedWallet = action.payload
            }
        },
        removeWallet: (state, action: PayloadAction<string>) => {
            state.wallets = state.wallets.filter(wallet => wallet.id !== action.payload)
            if (state.selectedWallet?.id === action.payload) {
                state.selectedWallet = null
            }
        },
        setSelectedWallet: (state, action: PayloadAction<Wallet | null>) => {
            state.selectedWallet = action.payload
        },

        // Private key actions
        setPrivateKeys: (state, action: PayloadAction<PrivateKey[]>) => {
            state.privateKeys = action.payload
            state.error = null
        },
        addPrivateKey: (state, action: PayloadAction<PrivateKey>) => {
            state.privateKeys.push(action.payload)
        },
        updatePrivateKey: (state, action: PayloadAction<PrivateKey>) => {
            const index = state.privateKeys.findIndex(key => key.id === action.payload.id)
            if (index !== -1) {
                state.privateKeys[index] = action.payload
            }
        },
        removePrivateKey: (state, action: PayloadAction<string>) => {
            state.privateKeys = state.privateKeys.filter(key => key.id !== action.payload)
        },

        // Common actions
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
} = walletSlice.actions

export default walletSlice.reducer
