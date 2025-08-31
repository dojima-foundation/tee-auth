import { render, screen, fireEvent } from '@testing-library/react'
import { Button } from '@/components/ui/button'

describe('UI Button Component', () => {
    it('renders with default props', () => {
        render(<Button>Click me</Button>)

        const button = screen.getByRole('button', { name: /click me/i })
        expect(button).toBeInTheDocument()
        expect(button).toHaveAttribute('data-slot', 'button')
    })

    it('renders with different variants', () => {
        const { rerender } = render(<Button variant="default">Default</Button>)
        expect(screen.getByRole('button')).toHaveClass('bg-primary', 'text-primary-foreground')

        rerender(<Button variant="destructive">Destructive</Button>)
        expect(screen.getByRole('button')).toHaveClass('bg-destructive', 'text-white')

        rerender(<Button variant="outline">Outline</Button>)
        expect(screen.getByRole('button')).toHaveClass('border', 'bg-background')

        rerender(<Button variant="secondary">Secondary</Button>)
        expect(screen.getByRole('button')).toHaveClass('bg-secondary', 'text-secondary-foreground')

        rerender(<Button variant="ghost">Ghost</Button>)
        expect(screen.getByRole('button')).toHaveClass('hover:bg-accent')

        rerender(<Button variant="link">Link</Button>)
        expect(screen.getByRole('button')).toHaveClass('text-primary', 'underline-offset-4')
    })

    it('renders with different sizes', () => {
        const { rerender } = render(<Button size="default">Default</Button>)
        expect(screen.getByRole('button')).toHaveClass('h-9', 'px-4', 'py-2')

        rerender(<Button size="sm">Small</Button>)
        expect(screen.getByRole('button')).toHaveClass('h-8', 'px-3')

        rerender(<Button size="lg">Large</Button>)
        expect(screen.getByRole('button')).toHaveClass('h-10', 'px-6')

        rerender(<Button size="icon">Icon</Button>)
        expect(screen.getByRole('button')).toHaveClass('size-9')
    })

    it('handles click events', () => {
        const handleClick = jest.fn()
        render(<Button onClick={handleClick}>Click me</Button>)

        fireEvent.click(screen.getByRole('button'))
        expect(handleClick).toHaveBeenCalledTimes(1)
    })

    it('can be disabled', () => {
        render(<Button disabled>Disabled</Button>)

        const button = screen.getByRole('button')
        expect(button).toBeDisabled()
        expect(button).toHaveClass('disabled:pointer-events-none', 'disabled:opacity-50')
    })

    it('renders as child component when asChild is true', () => {
        render(
            <Button asChild>
                <a href="/test">Link Button</a>
            </Button>
        )

        const link = screen.getByRole('link', { name: /link button/i })
        expect(link).toBeInTheDocument()
        expect(link).toHaveAttribute('href', '/test')
        expect(link).toHaveAttribute('data-slot', 'button')
    })

    it('applies custom className', () => {
        render(<Button className="custom-class">Custom</Button>)

        const button = screen.getByRole('button')
        expect(button).toHaveClass('custom-class')
    })

    it('forwards additional props', () => {
        render(<Button data-testid="custom-button" type="submit">Submit</Button>)

        const button = screen.getByTestId('custom-button')
        expect(button).toHaveAttribute('type', 'submit')
    })

    it('handles focus and keyboard events', () => {
        render(<Button>Focusable</Button>)

        const button = screen.getByRole('button')
        button.focus()
        expect(button).toHaveFocus()

        // Test keyboard interaction
        fireEvent.keyDown(button, { key: 'Enter', code: 'Enter' })
        fireEvent.keyDown(button, { key: ' ', code: 'Space' })
    })

    it('has proper accessibility attributes', () => {
        render(<Button aria-label="Custom label">Button</Button>)

        const button = screen.getByRole('button', { name: /custom label/i })
        expect(button).toHaveAttribute('aria-label', 'Custom label')
    })
})
