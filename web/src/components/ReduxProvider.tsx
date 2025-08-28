'use client';

import { Provider } from 'react-redux';
import { getStore } from '@/store';

interface ReduxProviderProps {
    children: React.ReactNode;
}

export function ReduxProvider({ children }: ReduxProviderProps) {
    const store = getStore();

    return (
        <Provider store={store}>
            {children}
        </Provider>
    );
}
