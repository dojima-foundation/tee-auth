import { call, put, takeLatest, delay } from 'redux-saga/effects'
import { PayloadAction } from '@reduxjs/toolkit'
import {
    setCurrentUser,
    setUsers,
    addUser,
    updateUser,
    removeUser,
    setLoading,
    setError,
    clearError,
    User,
} from '../slices/userSlice'

// Mock API functions (replace with real API calls)
const api = {
    getCurrentUser: async (): Promise<User> => {
        // Simulate API call
        await delay(1000)
        return {
            id: '1',
            name: 'John Doe',
            email: 'john@example.com',
            role: 'admin',
            createdAt: new Date().toISOString(),
        }
    },

    getUsers: async (): Promise<User[]> => {
        await delay(800)
        return [
            {
                id: '1',
                name: 'John Doe',
                email: 'john@example.com',
                role: 'admin',
                createdAt: new Date().toISOString(),
            },
            {
                id: '2',
                name: 'Jane Smith',
                email: 'jane@example.com',
                role: 'user',
                createdAt: new Date().toISOString(),
            },
        ]
    },

    createUser: async (userData: Partial<User>): Promise<User> => {
        await delay(500)
        return {
            id: Math.random().toString(36).substr(2, 9),
            name: userData.name || '',
            email: userData.email || '',
            role: userData.role || 'user',
            createdAt: new Date().toISOString(),
        }
    },

    updateUser: async (userId: string, userData: Partial<User>): Promise<User> => {
        await delay(500)
        return {
            id: userId,
            name: userData.name || '',
            email: userData.email || '',
            role: userData.role || 'user',
            createdAt: new Date().toISOString(),
        }
    },

    deleteUser: async (userId: string): Promise<void> => {
        await delay(300)
        // Simulate successful deletion
    },
}

// Action types for saga
export const USER_ACTIONS = {
    FETCH_CURRENT_USER: 'user/fetchCurrentUser',
    FETCH_USERS: 'user/fetchUsers',
    CREATE_USER: 'user/createUser',
    UPDATE_USER: 'user/updateUser',
    DELETE_USER: 'user/deleteUser',
    LOGOUT: 'user/logout',
} as const

// Action creators for saga
export const userActions = {
    fetchCurrentUser: () => ({ type: USER_ACTIONS.FETCH_CURRENT_USER }),
    fetchUsers: () => ({ type: USER_ACTIONS.FETCH_USERS }),
    createUser: (userData: Partial<User>) => ({
        type: USER_ACTIONS.CREATE_USER,
        payload: userData
    }),
    updateUser: (userId: string, userData: Partial<User>) => ({
        type: USER_ACTIONS.UPDATE_USER,
        payload: { userId, userData }
    }),
    deleteUser: (userId: string) => ({
        type: USER_ACTIONS.DELETE_USER,
        payload: userId
    }),
    logout: () => ({ type: USER_ACTIONS.LOGOUT }),
}

// Saga functions
function* fetchCurrentUserSaga() {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const user: User = yield call(api.getCurrentUser)
        yield put(setCurrentUser(user))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to fetch user'))
    } finally {
        yield put(setLoading(false))
    }
}

function* fetchUsersSaga() {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const users: User[] = yield call(api.getUsers)
        yield put(setUsers(users))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to fetch users'))
    } finally {
        yield put(setLoading(false))
    }
}

function* createUserSaga(action: PayloadAction<Partial<User>>) {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const newUser: User = yield call(api.createUser, action.payload)
        yield put(addUser(newUser))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to create user'))
    } finally {
        yield put(setLoading(false))
    }
}

function* updateUserSaga(action: PayloadAction<{ userId: string; userData: Partial<User> }>) {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const { userId, userData } = action.payload
        // Use userId to ensure it's not marked as unused
        console.log(`Updating user with ID: ${userId}`)
        const updatedUser: User = yield call(api.updateUser, userId, userData)
        yield put(updateUser(updatedUser))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to update user'))
    } finally {
        yield put(setLoading(false))
    }
}

function* deleteUserSaga(action: PayloadAction<string>) {
    try {
        yield put(setLoading(true))
        yield put(clearError())

        const userId = action.payload
        yield call(api.deleteUser, userId)
        yield put(removeUser(userId))

    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to delete user'))
    } finally {
        yield put(setLoading(false))
    }
}

function* logoutSaga() {
    try {
        yield put(setCurrentUser(null))
        yield put(setUsers([]))
        yield put(clearError())
    } catch (error) {
        yield put(setError(error instanceof Error ? error.message : 'Failed to logout'))
    }
}

// Root user saga
export function* userSaga() {
    yield takeLatest(USER_ACTIONS.FETCH_CURRENT_USER, fetchCurrentUserSaga)
    yield takeLatest(USER_ACTIONS.FETCH_USERS, fetchUsersSaga)
    yield takeLatest(USER_ACTIONS.CREATE_USER, createUserSaga)
    yield takeLatest(USER_ACTIONS.UPDATE_USER, updateUserSaga)
    yield takeLatest(USER_ACTIONS.DELETE_USER, deleteUserSaga)
    yield takeLatest(USER_ACTIONS.LOGOUT, logoutSaga)
}
