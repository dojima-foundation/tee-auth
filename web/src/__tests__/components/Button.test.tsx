import { render, screen, fireEvent } from '@testing-library/react'
import { Button } from '@/components/Button'

describe('Button Component', () => {
    it('renders with default props', () => {
        render(<Button>Click me</Button>)

        const button = screen.getByRole('button', { name: /click me/i })
        expect(button).toBeInTheDocument()
        expect(button).toHaveClass('bg-blue-600', 'text-white')
    })

    it('renders with different variants', () => {
        const { rerender } = render(<Button variant="primary">Primary</Button>)
        expect(screen.getByRole('button')).toHaveClass('bg-blue-600')

        rerender(<Button variant="secondary">Secondary</Button>)
        expect(screen.getByRole('button')).toHaveClass('bg-gray-200')

        rerender(<Button variant="danger">Danger</Button>)
        expect(screen.getByRole('button')).toHaveClass('bg-red-600')
    })

    it('renders with different sizes', () => {
        const { rerender } = render(<Button size="small">Small</Button>)
        expect(screen.getByRole('button')).toHaveClass('px-3', 'py-1.5', 'text-sm')

        rerender(<Button size="medium">Medium</Button>)
        expect(screen.getByRole('button')).toHaveClass('px-4', 'py-2', 'text-base')

        rerender(<Button size="large">Large</Button>)
        expect(screen.getByRole('button')).toHaveClass('px-6', 'py-3', 'text-lg')
    })

    it('handles click events', () => {
        const handleClick = jest.fn()
        render(<Button onClick={handleClick}>Click me</Button>)

        fireEvent.click(screen.getByRole('button'))
        expect(handleClick).toHaveBeenCalledTimes(1)
    })

    it('shows loading state', () => {
        render(<Button loading>Loading</Button>)

        const button = screen.getByRole('button')
        expect(button).toBeDisabled()
        expect(screen.getByTestId('loading-spinner')).toBeInTheDocument()
    })

    it('can be disabled', () => {
        render(<Button disabled>Disabled</Button>)

        const button = screen.getByRole('button')
        expect(button).toBeDisabled()
        expect(button).toHaveClass('opacity-50', 'cursor-not-allowed')
    })

    it('is disabled when loading', () => {
        render(<Button loading>Loading</Button>)

        const button = screen.getByRole('button')
        expect(button).toBeDisabled()
    })

    it('renders with custom test id', () => {
        render(<Button data-testid="custom-button">Custom</Button>)

        expect(screen.getByTestId('custom-button')).toBeInTheDocument()
    })

    it('renders with different button types', () => {
        const { rerender } = render(<Button type="submit">Submit</Button>)
        expect(screen.getByRole('button')).toHaveAttribute('type', 'submit')

        rerender(<Button type="reset">Reset</Button>)
        expect(screen.getByRole('button')).toHaveAttribute('type', 'reset')

        rerender(<Button type="button">Button</Button>)
        expect(screen.getByRole('button')).toHaveAttribute('type', 'button')
    })

    it('does not call onClick when disabled', () => {
        const handleClick = jest.fn()
        render(<Button disabled onClick={handleClick}>Disabled</Button>)

        fireEvent.click(screen.getByRole('button'))
        expect(handleClick).not.toHaveBeenCalled()
    })

    it('does not call onClick when loading', () => {
        const handleClick = jest.fn()
        render(<Button loading onClick={handleClick}>Loading</Button>)

        fireEvent.click(screen.getByRole('button'))
        expect(handleClick).not.toHaveBeenCalled()
    })
})
