import { render, screen } from '@testing-library/react'
import {
    Card,
    CardHeader,
    CardTitle,
    CardDescription,
    CardContent,
    CardFooter,
    CardAction,
} from '@/components/ui/card'

describe('Card Components', () => {
    describe('Card', () => {
        it('renders with default props', () => {
            render(<Card>Card content</Card>)

            const card = screen.getByText('Card content')
            expect(card).toBeInTheDocument()
            expect(card).toHaveAttribute('data-slot', 'card')
            expect(card).toHaveClass('bg-card', 'text-card-foreground', 'rounded-xl', 'border')
        })

        it('applies custom className', () => {
            render(<Card className="custom-card">Custom card</Card>)

            const card = screen.getByText('Custom card')
            expect(card).toHaveClass('custom-card')
        })

        it('forwards additional props', () => {
            render(<Card data-testid="test-card">Test card</Card>)

            const card = screen.getByTestId('test-card')
            expect(card).toBeInTheDocument()
        })
    })

    describe('CardHeader', () => {
        it('renders with default props', () => {
            render(
                <Card>
                    <CardHeader>Header content</CardHeader>
                </Card>
            )

            const header = screen.getByText('Header content')
            expect(header).toBeInTheDocument()
            expect(header).toHaveAttribute('data-slot', 'card-header')
            expect(header).toHaveClass('px-6')
        })

        it('applies custom className', () => {
            render(
                <Card>
                    <CardHeader className="custom-header">Custom header</CardHeader>
                </Card>
            )

            const header = screen.getByText('Custom header')
            expect(header).toHaveClass('custom-header')
        })
    })

    describe('CardTitle', () => {
        it('renders with default props', () => {
            render(
                <Card>
                    <CardHeader>
                        <CardTitle>Card Title</CardTitle>
                    </CardHeader>
                </Card>
            )

            const title = screen.getByText('Card Title')
            expect(title).toBeInTheDocument()
            expect(title).toHaveAttribute('data-slot', 'card-title')
            expect(title).toHaveClass('font-semibold', 'leading-none')
        })

        it('applies custom className', () => {
            render(
                <Card>
                    <CardHeader>
                        <CardTitle className="custom-title">Custom Title</CardTitle>
                    </CardHeader>
                </Card>
            )

            const title = screen.getByText('Custom Title')
            expect(title).toHaveClass('custom-title')
        })
    })

    describe('CardDescription', () => {
        it('renders with default props', () => {
            render(
                <Card>
                    <CardHeader>
                        <CardDescription>Card description</CardDescription>
                    </CardHeader>
                </Card>
            )

            const description = screen.getByText('Card description')
            expect(description).toBeInTheDocument()
            expect(description).toHaveAttribute('data-slot', 'card-description')
            expect(description).toHaveClass('text-muted-foreground', 'text-sm')
        })

        it('applies custom className', () => {
            render(
                <Card>
                    <CardHeader>
                        <CardDescription className="custom-description">Custom description</CardDescription>
                    </CardHeader>
                </Card>
            )

            const description = screen.getByText('Custom description')
            expect(description).toHaveClass('custom-description')
        })
    })

    describe('CardContent', () => {
        it('renders with default props', () => {
            render(
                <Card>
                    <CardContent>Card content</CardContent>
                </Card>
            )

            const content = screen.getByText('Card content')
            expect(content).toBeInTheDocument()
            expect(content).toHaveAttribute('data-slot', 'card-content')
            expect(content).toHaveClass('px-6')
        })

        it('applies custom className', () => {
            render(
                <Card>
                    <CardContent className="custom-content">Custom content</CardContent>
                </Card>
            )

            const content = screen.getByText('Custom content')
            expect(content).toHaveClass('custom-content')
        })
    })

    describe('CardFooter', () => {
        it('renders with default props', () => {
            render(
                <Card>
                    <CardFooter>Footer content</CardFooter>
                </Card>
            )

            const footer = screen.getByText('Footer content')
            expect(footer).toBeInTheDocument()
            expect(footer).toHaveAttribute('data-slot', 'card-footer')
            expect(footer).toHaveClass('flex', 'items-center', 'px-6')
        })

        it('applies custom className', () => {
            render(
                <Card>
                    <CardFooter className="custom-footer">Custom footer</CardFooter>
                </Card>
            )

            const footer = screen.getByText('Custom footer')
            expect(footer).toHaveClass('custom-footer')
        })
    })

    describe('CardAction', () => {
        it('renders with default props', () => {
            render(
                <Card>
                    <CardHeader>
                        <CardAction>Action content</CardAction>
                    </CardHeader>
                </Card>
            )

            const action = screen.getByText('Action content')
            expect(action).toBeInTheDocument()
            expect(action).toHaveAttribute('data-slot', 'card-action')
            expect(action).toHaveClass('col-start-2', 'row-span-2', 'row-start-1')
        })

        it('applies custom className', () => {
            render(
                <Card>
                    <CardHeader>
                        <CardAction className="custom-action">Custom action</CardAction>
                    </CardHeader>
                </Card>
            )

            const action = screen.getByText('Custom action')
            expect(action).toHaveClass('custom-action')
        })
    })

    describe('Complete Card Structure', () => {
        it('renders a complete card with all components', () => {
            render(
                <Card>
                    <CardHeader>
                        <CardTitle>Test Card</CardTitle>
                        <CardDescription>This is a test card description</CardDescription>
                        <CardAction>
                            <button>Action</button>
                        </CardAction>
                    </CardHeader>
                    <CardContent>
                        <p>This is the main content of the card.</p>
                    </CardContent>
                    <CardFooter>
                        <button>Footer Action</button>
                    </CardFooter>
                </Card>
            )

            expect(screen.getByText('Test Card')).toBeInTheDocument()
            expect(screen.getByText('This is a test card description')).toBeInTheDocument()
            expect(screen.getByText('This is the main content of the card.')).toBeInTheDocument()
            expect(screen.getByRole('button', { name: 'Action' })).toBeInTheDocument()
            expect(screen.getByRole('button', { name: 'Footer Action' })).toBeInTheDocument()
        })
    })
})
