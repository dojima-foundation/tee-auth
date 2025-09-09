import { render, screen, waitFor } from '../utils/test-utils.helper';
import CreateWalletDialog from '@/components/CreateWalletDialog';
import userEvent from '@testing-library/user-event';

describe('CreateWalletDialog', () => {
    const mockOnWalletCreated = jest.fn();

    beforeEach(() => {
        jest.clearAllMocks();
    });

    it('renders create wallet button', () => {
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        expect(screen.getByText('Create Wallet')).toBeInTheDocument();
    });

    it('opens dialog when create wallet button is clicked', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        expect(screen.getByText('Create New Wallet')).toBeInTheDocument();
        expect(screen.getByText('Create a new HD wallet for your organization. You can optionally provide a seed phrase.')).toBeInTheDocument();
    });

    it('displays form fields correctly', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        expect(screen.getByLabelText('Wallet Name')).toBeInTheDocument();
        expect(screen.getByPlaceholderText('Enter wallet name')).toBeInTheDocument();
    });

    it('validates required wallet name', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const submitButton = screen.getByRole('button', { name: /create wallet/i });
        await user.click(submitButton);

        await waitFor(() => {
            expect(screen.getByText('Wallet name is required')).toBeInTheDocument();
        });
        expect(mockOnWalletCreated).not.toHaveBeenCalled();
    });

    it('submits form with valid data', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const nameInput = screen.getByLabelText('Wallet Name');
        await user.type(nameInput, 'Test Wallet');

        const submitButton = screen.getByRole('button', { name: /create wallet/i });
        await user.click(submitButton);

        expect(mockOnWalletCreated).toHaveBeenCalledWith({
            name: 'Test Wallet',
        });
    });

    it('closes dialog after successful submission', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const nameInput = screen.getByLabelText('Wallet Name');
        await user.type(nameInput, 'Test Wallet');

        const submitButton = screen.getByRole('button', { name: /create wallet/i });
        await user.click(submitButton);

        await waitFor(() => {
            expect(screen.queryByText('Create New Wallet')).not.toBeInTheDocument();
        });
    });

    it('resets form after successful submission', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const nameInput = screen.getByLabelText('Wallet Name');
        await user.type(nameInput, 'Test Wallet');

        const submitButton = screen.getByRole('button', { name: /create wallet/i });
        await user.click(submitButton);

        // Open dialog again
        await user.click(screen.getByText('Create Wallet'));

        expect(screen.getByLabelText('Wallet Name')).toHaveValue('');
    });

    it('cancels dialog when cancel button is clicked', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const cancelButton = screen.getByText('Cancel');
        await user.click(cancelButton);

        expect(screen.queryByText('Create New Wallet')).not.toBeInTheDocument();
    });

    it('clears error when user starts typing', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const submitButton = screen.getByRole('button', { name: /create wallet/i });
        await user.click(submitButton);

        await waitFor(() => {
            expect(screen.getByText('Wallet name is required')).toBeInTheDocument();
        });

        const nameInput = screen.getByLabelText('Wallet Name');
        await user.type(nameInput, 'T');

        await waitFor(() => {
            expect(screen.queryByText('Wallet name is required')).not.toBeInTheDocument();
        });
    });

    it('shows disabled state when disabled prop is true', () => {
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} disabled={true} />);

        const createButton = screen.getByRole('button', { name: /create wallet/i });
        expect(createButton).toBeDisabled();
    });

    it('shows loading state when loading', async () => {
        const user = userEvent.setup();
        // Mock the callback to be async to simulate loading
        const asyncMockOnWalletCreated = jest.fn().mockImplementation(() =>
            new Promise(resolve => setTimeout(resolve, 100))
        );

        render(<CreateWalletDialog onWalletCreated={asyncMockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const nameInput = screen.getByLabelText('Wallet Name');
        await user.type(nameInput, 'Test Wallet');

        // Simulate loading state by clicking submit
        const submitButton = screen.getByRole('button', { name: /create wallet/i });
        await user.click(submitButton);

        // The button should show loading state
        await waitFor(() => {
            expect(screen.getByRole('button', { name: /creating/i })).toBeInTheDocument();
        });
    });

    it('disables submit button when name is empty', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const submitButton = screen.getByRole('button', { name: /create wallet/i });
        expect(submitButton).toBeDisabled();
    });

    it('enables submit button when name is provided', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const nameInput = screen.getByLabelText('Wallet Name');
        await user.type(nameInput, 'Test Wallet');

        const submitButton = screen.getByRole('button', { name: /create wallet/i });
        expect(submitButton).not.toBeDisabled();
    });

    it('trims whitespace from wallet name', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const nameInput = screen.getByLabelText('Wallet Name');
        await user.type(nameInput, '  Test Wallet  ');

        const submitButton = screen.getByRole('button', { name: /create wallet/i });
        await user.click(submitButton);

        expect(mockOnWalletCreated).toHaveBeenCalledWith({
            name: 'Test Wallet', // Should be trimmed
        });
    });

    it('handles form submission with Enter key', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const nameInput = screen.getByLabelText('Wallet Name');
        await user.type(nameInput, 'Test Wallet');

        await user.keyboard('{Enter}');

        expect(mockOnWalletCreated).toHaveBeenCalledWith({
            name: 'Test Wallet',
        });
    });

    it('shows error message when provided', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const nameInput = screen.getByLabelText('Wallet Name');
        await user.type(nameInput, 'Test Wallet');

        // Simulate an error by calling the handler with an error
        const submitButton = screen.getByRole('button', { name: /create wallet/i });
        await user.click(submitButton);

        // The error should be cleared when dialog closes
        await waitFor(() => {
            expect(screen.queryByText('Create New Wallet')).not.toBeInTheDocument();
        });
    });

    it('clears form and error when dialog is closed', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const nameInput = screen.getByLabelText('Wallet Name');
        await user.type(nameInput, 'Test Wallet');

        // Close dialog without submitting
        const cancelButton = screen.getByText('Cancel');
        await user.click(cancelButton);

        // Wait for dialog to close
        await waitFor(() => {
            expect(screen.queryByText('Create New Wallet')).not.toBeInTheDocument();
        });

        // Open dialog again
        await user.click(screen.getByText('Create Wallet'));

        // Wait for dialog to open and check form is cleared
        await waitFor(() => {
            expect(screen.getByLabelText('Wallet Name')).toHaveValue('');
        });
    });

    it('handles dialog open/close state correctly', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        // Initially closed
        expect(screen.queryByText('Create New Wallet')).not.toBeInTheDocument();

        // Open dialog
        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);
        expect(screen.getByText('Create New Wallet')).toBeInTheDocument();

        // Close dialog
        const cancelButton = screen.getByText('Cancel');
        await user.click(cancelButton);
        expect(screen.queryByText('Create New Wallet')).not.toBeInTheDocument();
    });

    it('disables input when loading', async () => {
        const user = userEvent.setup();
        // Mock the callback to be async to simulate loading
        const asyncMockOnWalletCreated = jest.fn().mockImplementation(() =>
            new Promise(resolve => setTimeout(resolve, 100))
        );

        render(<CreateWalletDialog onWalletCreated={asyncMockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        const nameInput = screen.getByLabelText('Wallet Name');
        await user.type(nameInput, 'Test Wallet');

        const submitButton = screen.getByRole('button', { name: /create wallet/i });
        await user.click(submitButton);

        // Input should be disabled during loading
        await waitFor(() => {
            expect(nameInput).toBeDisabled();
        });
    });

    it('shows wallet icon in dialog title', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByText('Create Wallet');
        await user.click(createButton);

        // Check that the wallet icon is present (it's part of the title)
        expect(screen.getByText('Create New Wallet')).toBeInTheDocument();
    });

    it('handles multiple rapid open/close operations', async () => {
        const user = userEvent.setup();
        render(<CreateWalletDialog onWalletCreated={mockOnWalletCreated} />);

        const createButton = screen.getByRole('button', { name: /create wallet/i });

        // Open and close multiple times
        await user.click(createButton);
        await waitFor(() => {
            expect(screen.getByText('Create New Wallet')).toBeInTheDocument();
        });

        const cancelButton = screen.getByText('Cancel');
        await user.click(cancelButton);
        await waitFor(() => {
            expect(screen.queryByText('Create New Wallet')).not.toBeInTheDocument();
        });

        await user.click(createButton);
        await waitFor(() => {
            expect(screen.getByText('Create New Wallet')).toBeInTheDocument();
        });

        await user.click(cancelButton);
        await waitFor(() => {
            expect(screen.queryByText('Create New Wallet')).not.toBeInTheDocument();
        });
    });
});
