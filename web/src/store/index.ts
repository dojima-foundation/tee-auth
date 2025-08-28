import { configureStore } from '@reduxjs/toolkit'
import createSagaMiddleware from 'redux-saga'
import { all } from 'redux-saga/effects'

// Import reducers
import userReducer from './slices/userSlice'
import themeReducer from './slices/themeSlice'
import walletReducer from './slices/walletSlice'

// Import sagas
import { userSaga } from './sagas/userSaga'
import { themeSaga } from './sagas/themeSaga'
import { walletSaga } from './sagas/walletSaga'

// Root saga
function* rootSaga() {
    yield all([
        userSaga(),
        themeSaga(),
        walletSaga(),
    ])
}

// Create store function (to be called on client side only)
let store: ReturnType<typeof createStore> | undefined

function createStore() {
    // Create saga middleware
    const sagaMiddleware = createSagaMiddleware()

    // Configure store
    const newStore = configureStore({
        reducer: {
            user: userReducer,
            theme: themeReducer,
            wallet: walletReducer,
        },
        middleware: (getDefaultMiddleware) =>
            getDefaultMiddleware({
                thunk: false, // Disable thunk since we're using saga
                serializableCheck: {
                    ignoredActions: ['persist/PERSIST', 'persist/REHYDRATE'],
                },
            }).concat(sagaMiddleware),
        devTools: process.env.NODE_ENV !== 'production',
    })

    // Run saga
    sagaMiddleware.run(rootSaga)

    return newStore
}

// Get store function (creates store on first call)
export function getStore() {
    if (!store) {
        store = createStore()
    }
    return store
}

// Export types
export type RootState = ReturnType<ReturnType<typeof createStore>['getState']>
export type AppDispatch = ReturnType<typeof createStore>['dispatch']
