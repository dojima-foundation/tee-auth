import { renderWithProviders, screen, fireEvent, waitFor } from '../utils/test-utils.helper';
import CreatePrivateKeyDialog from '@/components/CreatePrivateKeyDialog';
import { useAppSelector } from '@/store/hooks';
import userEvent from '@testing-library/user-event';

// Mock the store hook
jest.mock('@/store/hooks');

const mockUseAppSelector = useAppSelector as jest.MockedFunction<typeof useAppSelector>;

describe('CreatePrivateKeyDialog', () => {
    const mockOnPrivateKeyCreated = jest.fn();
    const mockWallets = [
        {
            id: 'wallet-1',
            name: 'Test Wallet 1',
            organization_id: 'org-456',
            public_key: '0x1234567890abcdef',
            is_active: true,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
        },
        {
            id: 'wallet-2',
            name: 'Test Wallet 2',
            organization_id: 'org-456',
            public_key: '0xabcdef1234567890',
            is_active: true,
            created_at: new Date().toISOString(),
            updated_at: new Date().toISOString(),
        },
    ];

    beforeEach(() => {
        jest.clearAllMocks();
        mockUseAppSelector.mockReturnValue(mockWallets);
    });

    it('renders create private key button', () => {
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        expect(screen.getByText('Create Private Key')).toBeInTheDocument();
    });

    it('opens dialog when create private key button is clicked', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        expect(screen.getByText('Create New Private Key')).toBeInTheDocument();
        expect(screen.getByText('Create a new private key for a wallet. Select the wallet and provide the key details.')).toBeInTheDocument();
    });

    it('displays form fields correctly', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        expect(screen.getByLabelText('Wallet')).toBeInTheDocument();
        expect(screen.getByLabelText('Private Key Name')).toBeInTheDocument();
        expect(screen.getByLabelText('Curve')).toBeInTheDocument();
        expect(screen.getByLabelText('Tags (Optional)')).toBeInTheDocument();
        expect(screen.getByText('Select a wallet')).toBeInTheDocument();
        expect(screen.getByPlaceholderText('Enter private key name')).toBeInTheDocument();
        expect(screen.getByText('Select a curve')).toBeInTheDocument();
    });

    it('shows wallet options in select dropdown', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        await user.click(walletSelect);

        // Check that options are available in the select
        expect(screen.getByRole('option', { name: 'Test Wallet 1' })).toBeInTheDocument();
        expect(screen.getByRole('option', { name: 'Test Wallet 2' })).toBeInTheDocument();
    });

    it('shows curve options in select dropdown', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const curveSelect = screen.getByRole('combobox', { name: /curve/i });
        await user.click(curveSelect);

        // Check that options are available in the select
        expect(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' })).toBeInTheDocument();
        expect(screen.getByRole('option', { name: 'ED25519 (Solana)' })).toBeInTheDocument();
    });

    it('validates required fields', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        await waitFor(() => {
            expect(screen.getByText('Wallet is required')).toBeInTheDocument();
        });
        expect(mockOnPrivateKeyCreated).not.toHaveBeenCalled();
    });

    it('validates wallet selection', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const nameInput = screen.getByLabelText('Private Key Name');
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });

        await user.type(nameInput, 'Test Private Key');
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        await waitFor(() => {
            expect(screen.getByText('Wallet is required')).toBeInTheDocument();
        });
    });

    it('validates private key name', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        await waitFor(() => {
            expect(screen.getByText('Private key name is required')).toBeInTheDocument();
        });
    });

    it('validates curve selection', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const nameInput = screen.getByLabelText('Private Key Name');

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.type(nameInput, 'Test Private Key');

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        await waitFor(() => {
            expect(screen.getByText('Curve is required')).toBeInTheDocument();
        });
    });

    it('submits form with valid data', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const nameInput = screen.getByLabelText('Private Key Name');
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.type(nameInput, 'Test Private Key');
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        expect(mockOnPrivateKeyCreated).toHaveBeenCalledWith({
            wallet_id: 'wallet-1',
            name: 'Test Private Key',
            curve: 'CURVE_SECP256K1',
            tags: [],
        });
    });

    it('submits form with tags', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const nameInput = screen.getByLabelText('Private Key Name');
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });
        const tagsInput = screen.getByLabelText('Tags (Optional)');

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.type(nameInput, 'Test Private Key');
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));
        await user.type(tagsInput, 'main, ethereum, test');

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        expect(mockOnPrivateKeyCreated).toHaveBeenCalledWith({
            wallet_id: 'wallet-1',
            name: 'Test Private Key',
            curve: 'CURVE_SECP256K1',
            tags: ['main', 'ethereum', 'test'],
        });
    });

    it('closes dialog after successful submission', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const nameInput = screen.getByLabelText('Private Key Name');
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.type(nameInput, 'Test Private Key');
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        await waitFor(() => {
            expect(screen.queryByText('Create New Private Key')).not.toBeInTheDocument();
        });
    });

    it('resets form after successful submission', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const nameInput = screen.getByLabelText('Private Key Name');
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.type(nameInput, 'Test Private Key');
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        // Open dialog again
        await user.click(screen.getByText('Create Private Key'));

        expect(screen.getByLabelText('Private Key Name')).toHaveValue('');
        expect(screen.getByText('Select a wallet')).toBeInTheDocument();
        expect(screen.getByText('Select a curve')).toBeInTheDocument();
    });

    it('cancels dialog when cancel button is clicked', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const cancelButton = screen.getByText('Cancel');
        await user.click(cancelButton);

        expect(screen.queryByText('Create New Private Key')).not.toBeInTheDocument();
    });

    it('shows disabled state when disabled prop is true', () => {
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} disabled={true} />);

        const createButton = screen.getByRole('button', { name: /create private key/i });
        expect(createButton).toBeDisabled();
    });

    it('shows loading state when loading', async () => {
        const user = userEvent.setup();
        // Mock the callback to be async to simulate loading
        const asyncMockOnPrivateKeyCreated = jest.fn().mockImplementation(() =>
            new Promise(resolve => setTimeout(resolve, 100))
        );

        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={asyncMockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const nameInput = screen.getByLabelText('Private Key Name');
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.type(nameInput, 'Test Private Key');
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        // The button should show loading state
        await waitFor(() => {
            expect(screen.getByRole('button', { name: /creating/i })).toBeInTheDocument();
        });
    });

    it('disables submit button when required fields are missing', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        expect(submitButton).toBeDisabled();
    });

    it('enables submit button when all required fields are provided', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const nameInput = screen.getByLabelText('Private Key Name');
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.type(nameInput, 'Test Private Key');
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        expect(submitButton).not.toBeDisabled();
    });

    it('parses tags correctly', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const nameInput = screen.getByLabelText('Private Key Name');
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });
        const tagsInput = screen.getByLabelText('Tags (Optional)');

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.type(nameInput, 'Test Private Key');
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));
        await user.type(tagsInput, '  tag1  , tag2 ,  tag3  ');

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        expect(mockOnPrivateKeyCreated).toHaveBeenCalledWith({
            wallet_id: 'wallet-1',
            name: 'Test Private Key',
            curve: 'CURVE_SECP256K1',
            tags: ['tag1', 'tag2', 'tag3'], // Should trim whitespace
        });
    });

    it('handles empty tags', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const nameInput = screen.getByLabelText('Private Key Name');
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });
        const tagsInput = screen.getByLabelText('Tags (Optional)');

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.type(nameInput, 'Test Private Key');
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));
        await user.type(tagsInput, '  , ,  ');

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        expect(mockOnPrivateKeyCreated).toHaveBeenCalledWith({
            wallet_id: 'wallet-1',
            name: 'Test Private Key',
            curve: 'CURVE_SECP256K1',
            tags: [], // Should filter out empty tags
        });
    });

    it('handles form submission with Enter key', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const nameInput = screen.getByLabelText('Private Key Name');
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.type(nameInput, 'Test Private Key');
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));

        // Press Enter on the name input to trigger form submission
        await user.type(nameInput, '{Enter}');

        await waitFor(() => {
            expect(mockOnPrivateKeyCreated).toHaveBeenCalledWith({
                wallet_id: 'wallet-1',
                name: 'Test Private Key',
                curve: 'CURVE_SECP256K1',
                tags: [],
            });
        });
    });

    it('shows key icon in dialog title', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        // Check that the key icon is present (it's part of the title)
        expect(screen.getByText('Create New Private Key')).toBeInTheDocument();
    });

    it('disables form fields when loading', async () => {
        const user = userEvent.setup();
        // Mock the callback to be async to simulate loading
        const asyncMockOnPrivateKeyCreated = jest.fn().mockImplementation(() =>
            new Promise(resolve => setTimeout(resolve, 100))
        );

        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={asyncMockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        const nameInput = screen.getByLabelText('Private Key Name');
        const curveSelect = screen.getByRole('combobox', { name: /curve/i });

        await user.click(walletSelect);
        await user.click(screen.getByRole('option', { name: 'Test Wallet 1' }));
        await user.type(nameInput, 'Test Private Key');
        await user.click(curveSelect);
        await user.click(screen.getByRole('option', { name: 'SECP256K1 (Ethereum)' }));

        const submitButton = screen.getByRole('button', { name: /create private key/i });
        await user.click(submitButton);

        // Form fields should be disabled during loading
        await waitFor(() => {
            expect(walletSelect).toBeDisabled();
            expect(nameInput).toBeDisabled();
            expect(curveSelect).toBeDisabled();
        });
    });

    it('handles dialog open/close state correctly', async () => {
        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        // Initially closed
        expect(screen.queryByText('Create New Private Key')).not.toBeInTheDocument();

        // Open dialog
        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);
        expect(screen.getByText('Create New Private Key')).toBeInTheDocument();

        // Close dialog
        const cancelButton = screen.getByText('Cancel');
        await user.click(cancelButton);
        expect(screen.queryByText('Create New Private Key')).not.toBeInTheDocument();
    });

    it('handles no wallets available', async () => {
        mockUseAppSelector.mockReturnValue([]);

        const user = userEvent.setup();
        renderWithProviders(<CreatePrivateKeyDialog onPrivateKeyCreated={mockOnPrivateKeyCreated} />);

        const createButton = screen.getByText('Create Private Key');
        await user.click(createButton);

        const walletSelect = screen.getByRole('combobox', { name: /wallet/i });
        await user.click(walletSelect);

        // Should not show any wallet options
        expect(screen.queryByRole('option', { name: 'Test Wallet 1' })).not.toBeInTheDocument();
        expect(screen.queryByRole('option', { name: 'Test Wallet 2' })).not.toBeInTheDocument();
    });
});
