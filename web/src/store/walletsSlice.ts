import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import { gauthApi, type Wallet } from '@/services/gauthApi';

export interface WalletsState {
    wallets: Wallet[];
    loading: boolean;
    error: string | null;
    currentPage: number;
    totalPages: number;
    totalWallets: number;
}

const initialState: WalletsState = {
    wallets: [],
    loading: false,
    error: null,
    currentPage: 1,
    totalPages: 1,
    totalWallets: 0,
};

// Async thunk for fetching wallets
export const fetchWallets = createAsyncThunk(
    'wallets/fetchWallets',
    async ({ page = 1, pageSize = 10 }: {
        page?: number;
        pageSize?: number;
    }, { rejectWithValue }) => {
        console.log('📡 [WalletsSlice] fetchWallets called with:', { page, pageSize });
        try {
            console.log('🌐 [WalletsSlice] Making API call to getWallets...');
            const response = await gauthApi.getWallets();
            console.log('📡 [WalletsSlice] getWallets response:', {
                success: response.success,
                hasData: !!response.data,
                walletsCount: response.data?.wallets?.length || 0
            });


            if (response.success) {
                // Sanitize wallet data to ensure it has the expected structure
                const sanitizedWallets = (response.data.wallets || []).map(wallet => ({
                    ...wallet,
                    accounts: Array.isArray(wallet.accounts) ? wallet.accounts.map(account => ({
                        ...account,
                        public_key: typeof account.public_key === 'string' ? account.public_key : ''
                    })) : []
                }));


                return {
                    wallets: sanitizedWallets,
                    currentPage: page,
                    totalWallets: response.data.wallets?.length || 0,
                    totalPages: Math.ceil((response.data.wallets?.length || 0) / pageSize),
                };
            } else {
                throw new Error('Failed to fetch wallets');
            }
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to fetch wallets');
        }
    }
);

// Async thunk for creating a wallet
export const createWallet = createAsyncThunk(
    'wallets/createWallet',
    async (walletData: { name: string; seed_phrase?: string }, { rejectWithValue }) => {
        try {
            const response = await gauthApi.createWallet({
                name: walletData.name,
                accounts: [{
                    curve: 'CURVE_SECP256K1',
                    path_format: 'PATH_FORMAT_BIP32',
                    path: 'm/44\'/60\'/0\'/0/0',
                    address_format: 'ADDRESS_FORMAT_ETHEREUM'
                }],
                mnemonic_length: 12,
                tags: []
            });

            if (response.success) {
                return response.data.wallet;
            } else {
                throw new Error('Failed to create wallet');
            }
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to create wallet');
        }
    }
);

const walletsSlice = createSlice({
    name: 'wallets',
    initialState,
    reducers: {
        clearWallets: (state) => {
            state.wallets = [];
            state.loading = false;
            state.error = null;
            state.currentPage = 1;
            state.totalPages = 1;
            state.totalWallets = 0;
        },
        setWalletsLoading: (state, action: PayloadAction<boolean>) => {
            state.loading = action.payload;
        },
        setWalletsError: (state, action: PayloadAction<string | null>) => {
            state.error = action.payload;
        },
    },
    extraReducers: (builder) => {
        // Fetch wallets
        builder
            .addCase(fetchWallets.pending, (state) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(fetchWallets.fulfilled, (state, action) => {
                state.loading = false;
                state.wallets = action.payload.wallets;
                state.currentPage = action.payload.currentPage;
                state.totalPages = action.payload.totalPages;
                state.totalWallets = action.payload.totalWallets;
                state.error = null;
            })
            .addCase(fetchWallets.rejected, (state, action) => {
                state.loading = false;
                state.error = action.payload as string;
            });

        // Create wallet
        builder
            .addCase(createWallet.pending, (state) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(createWallet.fulfilled, (state, action) => {
                state.loading = false;
                state.wallets.push(action.payload);
                state.error = null;
            })
            .addCase(createWallet.rejected, (state, action) => {
                state.loading = false;
                state.error = action.payload as string;
            });
    },
});

export const {
    clearWallets,
    setWalletsLoading,
    setWalletsError,
} = walletsSlice.actions;

export default walletsSlice.reducer;

// Selectors
export const selectWallets = (state: { wallets: WalletsState }) => state.wallets.wallets;
export const selectWalletsLoading = (state: { wallets: WalletsState }) => state.wallets.loading;
export const selectWalletsError = (state: { wallets: WalletsState }) => state.wallets.error;
export const selectWalletsPagination = (state: { wallets: WalletsState }) => ({
    currentPage: state.wallets.currentPage,
    totalPages: state.wallets.totalPages,
    totalWallets: state.wallets.totalWallets,
});
