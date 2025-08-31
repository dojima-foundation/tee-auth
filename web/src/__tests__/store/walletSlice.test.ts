import walletReducer, {
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
} from '@/store/slices/walletSlice'
import type { Wallet, PrivateKey } from '@/store/slices/walletSlice'

describe('walletSlice', () => {
    const initialState = {
        wallets: [],
        privateKeys: [],
        selectedWallet: null,
        loading: false,
        error: null,
    }

    const mockWallet: Wallet = {
        id: '1',
        name: 'Test Wallet',
        address: '0x1234567890abcdef',
        balance: '1.5',
        currency: 'ETH',
        isActive: true,
        createdAt: '2024-01-01T00:00:00Z',
    }

    const mockWallets: Wallet[] = [
        mockWallet,
        {
            id: '2',
            name: 'Another Wallet',
            address: '0xabcdef1234567890',
            balance: '2.0',
            currency: 'BTC',
            isActive: false,
            createdAt: '2024-01-02T00:00:00Z',
        },
    ]

    const mockPrivateKey: PrivateKey = {
        id: '1',
        name: 'Test Private Key',
        publicKey: '0xpublickey123',
        encryptedPrivateKey: 'encrypted_data',
        walletId: '1',
        isActive: true,
        createdAt: '2024-01-01T00:00:00Z',
    }

    const mockPrivateKeys: PrivateKey[] = [
        mockPrivateKey,
        {
            id: '2',
            name: 'Another Private Key',
            publicKey: '0xpublickey456',
            encryptedPrivateKey: 'encrypted_data_2',
            walletId: '2',
            isActive: false,
            createdAt: '2024-01-02T00:00:00Z',
        },
    ]

    it('should return the initial state', () => {
        expect(walletReducer(undefined, { type: 'unknown' })).toEqual(initialState)
    })

    describe('Wallet actions', () => {
        describe('setWallets', () => {
            it('should set wallets array', () => {
                const result = walletReducer(initialState, setWallets(mockWallets))
                expect(result.wallets).toEqual(mockWallets)
                expect(result.error).toBeNull()
            })

            it('should clear error when setting wallets', () => {
                const stateWithError = { ...initialState, error: 'Some error' }
                const result = walletReducer(stateWithError, setWallets(mockWallets))
                expect(result.error).toBeNull()
            })
        })

        describe('addWallet', () => {
            it('should add wallet to wallets array', () => {
                const result = walletReducer(initialState, addWallet(mockWallet))
                expect(result.wallets).toHaveLength(1)
                expect(result.wallets[0]).toEqual(mockWallet)
            })

            it('should add wallet to existing wallets array', () => {
                const stateWithWallets = { ...initialState, wallets: [mockWallets[0]] }
                const result = walletReducer(stateWithWallets, addWallet(mockWallets[1]))
                expect(result.wallets).toHaveLength(2)
                expect(result.wallets[1]).toEqual(mockWallets[1])
            })
        })

        describe('updateWallet', () => {
            it('should update wallet in wallets array', () => {
                const stateWithWallets = { ...initialState, wallets: mockWallets }
                const updatedWallet = { ...mockWallet, name: 'Updated Wallet' }

                const result = walletReducer(stateWithWallets, updateWallet(updatedWallet))
                expect(result.wallets[0]).toEqual(updatedWallet)
            })

            it('should update selected wallet if it matches the updated wallet', () => {
                const stateWithSelectedWallet = { ...initialState, selectedWallet: mockWallet }
                const updatedWallet = { ...mockWallet, name: 'Updated Wallet' }

                const result = walletReducer(stateWithSelectedWallet, updateWallet(updatedWallet))
                expect(result.selectedWallet).toEqual(updatedWallet)
            })

            it('should not update if wallet is not found', () => {
                const stateWithWallets = { ...initialState, wallets: mockWallets }
                const nonExistentWallet = { ...mockWallet, id: '999' }

                const result = walletReducer(stateWithWallets, updateWallet(nonExistentWallet))
                expect(result.wallets).toEqual(mockWallets)
            })
        })

        describe('removeWallet', () => {
            it('should remove wallet from wallets array', () => {
                const stateWithWallets = { ...initialState, wallets: mockWallets }
                const result = walletReducer(stateWithWallets, removeWallet('1'))

                expect(result.wallets).toHaveLength(1)
                expect(result.wallets[0].id).toBe('2')
            })

            it('should clear selected wallet if it matches the removed wallet', () => {
                const stateWithSelectedWallet = { ...initialState, selectedWallet: mockWallet }
                const result = walletReducer(stateWithSelectedWallet, removeWallet('1'))

                expect(result.selectedWallet).toBeNull()
            })

            it('should not remove if wallet is not found', () => {
                const stateWithWallets = { ...initialState, wallets: mockWallets }
                const result = walletReducer(stateWithWallets, removeWallet('999'))

                expect(result.wallets).toEqual(mockWallets)
            })
        })

        describe('setSelectedWallet', () => {
            it('should set selected wallet', () => {
                const result = walletReducer(initialState, setSelectedWallet(mockWallet))
                expect(result.selectedWallet).toEqual(mockWallet)
            })

            it('should clear selected wallet when set to null', () => {
                const stateWithSelectedWallet = { ...initialState, selectedWallet: mockWallet }
                const result = walletReducer(stateWithSelectedWallet, setSelectedWallet(null))
                expect(result.selectedWallet).toBeNull()
            })
        })
    })

    describe('Private key actions', () => {
        describe('setPrivateKeys', () => {
            it('should set private keys array', () => {
                const result = walletReducer(initialState, setPrivateKeys(mockPrivateKeys))
                expect(result.privateKeys).toEqual(mockPrivateKeys)
                expect(result.error).toBeNull()
            })

            it('should clear error when setting private keys', () => {
                const stateWithError = { ...initialState, error: 'Some error' }
                const result = walletReducer(stateWithError, setPrivateKeys(mockPrivateKeys))
                expect(result.error).toBeNull()
            })
        })

        describe('addPrivateKey', () => {
            it('should add private key to private keys array', () => {
                const result = walletReducer(initialState, addPrivateKey(mockPrivateKey))
                expect(result.privateKeys).toHaveLength(1)
                expect(result.privateKeys[0]).toEqual(mockPrivateKey)
            })

            it('should add private key to existing private keys array', () => {
                const stateWithPrivateKeys = { ...initialState, privateKeys: [mockPrivateKeys[0]] }
                const result = walletReducer(stateWithPrivateKeys, addPrivateKey(mockPrivateKeys[1]))
                expect(result.privateKeys).toHaveLength(2)
                expect(result.privateKeys[1]).toEqual(mockPrivateKeys[1])
            })
        })

        describe('updatePrivateKey', () => {
            it('should update private key in private keys array', () => {
                const stateWithPrivateKeys = { ...initialState, privateKeys: mockPrivateKeys }
                const updatedPrivateKey = { ...mockPrivateKey, name: 'Updated Private Key' }

                const result = walletReducer(stateWithPrivateKeys, updatePrivateKey(updatedPrivateKey))
                expect(result.privateKeys[0]).toEqual(updatedPrivateKey)
            })

            it('should not update if private key is not found', () => {
                const stateWithPrivateKeys = { ...initialState, privateKeys: mockPrivateKeys }
                const nonExistentPrivateKey = { ...mockPrivateKey, id: '999' }

                const result = walletReducer(stateWithPrivateKeys, updatePrivateKey(nonExistentPrivateKey))
                expect(result.privateKeys).toEqual(mockPrivateKeys)
            })
        })

        describe('removePrivateKey', () => {
            it('should remove private key from private keys array', () => {
                const stateWithPrivateKeys = { ...initialState, privateKeys: mockPrivateKeys }
                const result = walletReducer(stateWithPrivateKeys, removePrivateKey('1'))

                expect(result.privateKeys).toHaveLength(1)
                expect(result.privateKeys[0].id).toBe('2')
            })

            it('should not remove if private key is not found', () => {
                const stateWithPrivateKeys = { ...initialState, privateKeys: mockPrivateKeys }
                const result = walletReducer(stateWithPrivateKeys, removePrivateKey('999'))

                expect(result.privateKeys).toEqual(mockPrivateKeys)
            })
        })
    })

    describe('Common actions', () => {
        describe('setLoading', () => {
            it('should set loading to true', () => {
                const result = walletReducer(initialState, setLoading(true))
                expect(result.loading).toBe(true)
            })

            it('should set loading to false', () => {
                const stateWithLoading = { ...initialState, loading: true }
                const result = walletReducer(stateWithLoading, setLoading(false))
                expect(result.loading).toBe(false)
            })
        })

        describe('setError', () => {
            it('should set error and clear loading', () => {
                const stateWithLoading = { ...initialState, loading: true }
                const result = walletReducer(stateWithLoading, setError('Something went wrong'))

                expect(result.error).toBe('Something went wrong')
                expect(result.loading).toBe(false)
            })

            it('should clear error when set to null', () => {
                const stateWithError = { ...initialState, error: 'Some error' }
                const result = walletReducer(stateWithError, setError(null))

                expect(result.error).toBeNull()
            })
        })

        describe('clearError', () => {
            it('should clear error', () => {
                const stateWithError = { ...initialState, error: 'Some error' }
                const result = walletReducer(stateWithError, clearError())

                expect(result.error).toBeNull()
            })
        })
    })

    describe('complex scenarios', () => {
        it('should handle wallet and private key operations together', () => {
            let state = initialState

            // Set wallets
            state = walletReducer(state, setWallets(mockWallets))
            expect(state.wallets).toEqual(mockWallets)

            // Set private keys
            state = walletReducer(state, setPrivateKeys(mockPrivateKeys))
            expect(state.privateKeys).toEqual(mockPrivateKeys)

            // Select a wallet
            state = walletReducer(state, setSelectedWallet(mockWallet))
            expect(state.selectedWallet).toEqual(mockWallet)

            // Update the selected wallet
            const updatedWallet = { ...mockWallet, balance: '3.0' }
            state = walletReducer(state, updateWallet(updatedWallet))
            expect(state.selectedWallet).toEqual(updatedWallet)
            expect(state.wallets[0]).toEqual(updatedWallet)

            // Remove a private key
            state = walletReducer(state, removePrivateKey('1'))
            expect(state.privateKeys).toHaveLength(1)

            // Remove the selected wallet
            state = walletReducer(state, removeWallet('1'))
            expect(state.wallets).toHaveLength(1)
            expect(state.selectedWallet).toBeNull()
        })
    })
})
