import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import { gauthApi, type PrivateKey } from '@/services/gauthApi';

export interface PrivateKeysState {
    privateKeys: PrivateKey[];
    loading: boolean;
    error: string | null;
    currentPage: number;
    totalPages: number;
    totalPrivateKeys: number;
}

const initialState: PrivateKeysState = {
    privateKeys: [],
    loading: false,
    error: null,
    currentPage: 1,
    totalPages: 1,
    totalPrivateKeys: 0,
};

// Async thunk for fetching private keys
export const fetchPrivateKeys = createAsyncThunk(
    'privateKeys/fetchPrivateKeys',
    async ({ organizationId, page = 1, pageSize = 10 }: {
        organizationId: string;
        page?: number;
        pageSize?: number;
    }, { rejectWithValue }) => {
        try {
            const response = await gauthApi.getPrivateKeys(organizationId);

            if (response.success) {
                return {
                    privateKeys: response.data.private_keys || [],
                    currentPage: page,
                    totalPrivateKeys: response.data.private_keys?.length || 0,
                    totalPages: Math.ceil((response.data.private_keys?.length || 0) / pageSize),
                };
            } else {
                throw new Error('Failed to fetch private keys');
            }
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to fetch private keys');
        }
    }
);

// Async thunk for creating a private key
export const createPrivateKey = createAsyncThunk(
    'privateKeys/createPrivateKey',
    async ({ organizationId, privateKeyData }: {
        organizationId: string;
        privateKeyData: { wallet_id: string; name: string; curve: string; tags?: string[] };
    }, { rejectWithValue }) => {
        try {
            const response = await gauthApi.createPrivateKey({
                organization_id: organizationId,
                wallet_id: privateKeyData.wallet_id,
                name: privateKeyData.name,
                curve: privateKeyData.curve,
                tags: privateKeyData.tags,
            });

            if (response.success) {
                return response.data.private_key;
            } else {
                throw new Error('Failed to create private key');
            }
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to create private key');
        }
    }
);

const privateKeysSlice = createSlice({
    name: 'privateKeys',
    initialState,
    reducers: {
        clearPrivateKeys: (state) => {
            state.privateKeys = [];
            state.loading = false;
            state.error = null;
            state.currentPage = 1;
            state.totalPages = 1;
            state.totalPrivateKeys = 0;
        },
        setPrivateKeysLoading: (state, action: PayloadAction<boolean>) => {
            state.loading = action.payload;
        },
        setPrivateKeysError: (state, action: PayloadAction<string | null>) => {
            state.error = action.payload;
        },
    },
    extraReducers: (builder) => {
        // Fetch private keys
        builder
            .addCase(fetchPrivateKeys.pending, (state) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(fetchPrivateKeys.fulfilled, (state, action) => {
                state.loading = false;
                state.privateKeys = action.payload.privateKeys;
                state.currentPage = action.payload.currentPage;
                state.totalPages = action.payload.totalPages;
                state.totalPrivateKeys = action.payload.totalPrivateKeys;
                state.error = null;
            })
            .addCase(fetchPrivateKeys.rejected, (state, action) => {
                state.loading = false;
                state.error = action.payload as string;
            });

        // Create private key
        builder
            .addCase(createPrivateKey.pending, (state) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(createPrivateKey.fulfilled, (state, action) => {
                state.loading = false;
                state.privateKeys.push(action.payload);
                state.error = null;
            })
            .addCase(createPrivateKey.rejected, (state, action) => {
                state.loading = false;
                state.error = action.payload as string;
            });
    },
});

export const {
    clearPrivateKeys,
    setPrivateKeysLoading,
    setPrivateKeysError,
} = privateKeysSlice.actions;

export default privateKeysSlice.reducer;

// Selectors
export const selectPrivateKeys = (state: { privateKeys: PrivateKeysState }) => state.privateKeys.privateKeys;
export const selectPrivateKeysLoading = (state: { privateKeys: PrivateKeysState }) => state.privateKeys.loading;
export const selectPrivateKeysError = (state: { privateKeys: PrivateKeysState }) => state.privateKeys.error;
export const selectPrivateKeysPagination = (state: { privateKeys: PrivateKeysState }) => ({
    currentPage: state.privateKeys.currentPage,
    totalPages: state.privateKeys.totalPages,
    totalPrivateKeys: state.privateKeys.totalPrivateKeys,
});
