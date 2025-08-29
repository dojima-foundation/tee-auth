import { createSlice, PayloadAction, createAsyncThunk } from '@reduxjs/toolkit';
import { gauthApi } from '@/services/gauthApi';

export interface User {
    id: string;
    organization_id: string;
    username: string;
    email: string;
    public_key?: string;
    is_active: boolean;
    created_at: string;
    updated_at: string;
}

export interface UsersState {
    users: User[];
    loading: boolean;
    error: string | null;
    currentPage: number;
    totalPages: number;
    totalUsers: number;
}

const initialState: UsersState = {
    users: [],
    loading: false,
    error: null,
    currentPage: 1,
    totalPages: 1,
    totalUsers: 0,
};

// Async thunk for fetching users
export const fetchUsers = createAsyncThunk(
    'users/fetchUsers',
    async ({ organizationId, page = 1, pageSize = 10 }: {
        organizationId: string;
        page?: number;
        pageSize?: number;
    }, { rejectWithValue }) => {
        try {
            const response = await gauthApi.getUsers(organizationId);

            if (response.success) {
                return {
                    users: response.data.users || [],
                    currentPage: page,
                    totalUsers: response.data.users?.length || 0,
                    totalPages: Math.ceil((response.data.users?.length || 0) / pageSize),
                };
            } else {
                throw new Error('Failed to fetch users');
            }
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to fetch users');
        }
    }
);

// Async thunk for creating a user
export const createUser = createAsyncThunk(
    'users/createUser',
    async ({ organizationId, userData }: {
        organizationId: string;
        userData: Partial<User>;
    }, { rejectWithValue }) => {
        try {
            const response = await gauthApi.createUser({
                organization_id: organizationId,
                username: userData.username || '',
                email: userData.email || '',
                public_key: userData.public_key,
            });

            if (response.success) {
                return response.data.user;
            } else {
                throw new Error('Failed to create user');
            }
        } catch (error) {
            return rejectWithValue(error instanceof Error ? error.message : 'Failed to create user');
        }
    }
);

// Note: updateUser and deleteUser methods are not available in the current API
// These would need to be implemented in the backend API service

const usersSlice = createSlice({
    name: 'users',
    initialState,
    reducers: {
        clearUsers: (state) => {
            state.users = [];
            state.loading = false;
            state.error = null;
            state.currentPage = 1;
            state.totalPages = 1;
            state.totalUsers = 0;
        },
        setUsersLoading: (state, action: PayloadAction<boolean>) => {
            state.loading = action.payload;
        },
        setUsersError: (state, action: PayloadAction<string | null>) => {
            state.error = action.payload;
        },
    },
    extraReducers: (builder) => {
        // Fetch users
        builder
            .addCase(fetchUsers.pending, (state) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(fetchUsers.fulfilled, (state, action) => {
                state.loading = false;
                state.users = action.payload.users;
                state.currentPage = action.payload.currentPage;
                state.totalPages = action.payload.totalPages;
                state.totalUsers = action.payload.totalUsers;
                state.error = null;
            })
            .addCase(fetchUsers.rejected, (state, action) => {
                state.loading = false;
                state.error = action.payload as string;
            });

        // Create user
        builder
            .addCase(createUser.pending, (state) => {
                state.loading = true;
                state.error = null;
            })
            .addCase(createUser.fulfilled, (state, action) => {
                state.loading = false;
                state.users.push(action.payload);
                state.error = null;
            })
            .addCase(createUser.rejected, (state, action) => {
                state.loading = false;
                state.error = action.payload as string;
            });


    },
});

export const {
    clearUsers,
    setUsersLoading,
    setUsersError,
} = usersSlice.actions;

export default usersSlice.reducer;

// Selectors
export const selectUsers = (state: { users: UsersState }) => state.users.users;
export const selectUsersLoading = (state: { users: UsersState }) => state.users.loading;
export const selectUsersError = (state: { users: UsersState }) => state.users.error;
export const selectUsersPagination = (state: { users: UsersState }) => ({
    currentPage: state.users.currentPage,
    totalPages: state.users.totalPages,
    totalUsers: state.users.totalUsers,
});
