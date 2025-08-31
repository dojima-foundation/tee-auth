import { render, screen } from '@testing-library/react'
import Home from '@/app/page'

// Mock Next.js Link component
jest.mock('next/link', () => {
    return ({ children, href, ...props }: any) => (
        <a href={href} {...props}>
            {children}
        </a>
    )
})

describe('Home Page', () => {
    it('renders the main heading', () => {
        render(<Home />)

        const heading = screen.getByRole('heading', { name: /welcome to odeys/i })
        expect(heading).toBeInTheDocument()
        expect(heading.tagName).toBe('H1')
    })

    it('renders the description text', () => {
        render(<Home />)

        const description = screen.getByText(/secure authentication and user management platform/i)
        expect(description).toBeInTheDocument()
    })

    it('renders the sign in link', () => {
        render(<Home />)

        const signInLink = screen.getByRole('link', { name: /sign in/i })
        expect(signInLink).toBeInTheDocument()
        expect(signInLink).toHaveAttribute('href', '/auth/signin')
    })

    it('renders the dashboard link', () => {
        render(<Home />)

        const dashboardLink = screen.getByRole('link', { name: /go to dashboard/i })
        expect(dashboardLink).toBeInTheDocument()
        expect(dashboardLink).toHaveAttribute('href', '/dashboard')
    })

    it('has proper styling classes', () => {
        render(<Home />)

        const main = screen.getByRole('main')
        expect(main).toHaveClass('min-h-screen', 'bg-background', 'text-foreground')

        const container = main.querySelector('.container')
        expect(container).toHaveClass('mx-auto', 'px-4', 'py-8')
    })

    it('has responsive layout classes', () => {
        render(<Home />)

        const signInLink = screen.getByRole('link', { name: /sign in/i })
        const buttonContainer = signInLink.parentElement
        expect(buttonContainer).toHaveClass('flex', 'flex-col', 'sm:flex-row', 'gap-4')
    })

    it('has proper button styling', () => {
        render(<Home />)

        const signInButton = screen.getByRole('link', { name: /sign in/i })
        expect(signInButton).toHaveClass(
            'inline-block',
            'bg-primary',
            'hover:bg-primary/90',
            'text-primary-foreground',
            'font-semibold',
            'py-3',
            'px-6',
            'rounded-lg',
            'transition-colors',
            'duration-200',
            'text-center'
        )

        const dashboardButton = screen.getByRole('link', { name: /go to dashboard/i })
        expect(dashboardButton).toHaveClass(
            'inline-block',
            'bg-secondary',
            'hover:bg-secondary/80',
            'text-secondary-foreground',
            'font-semibold',
            'py-3',
            'px-6',
            'rounded-lg',
            'transition-colors',
            'duration-200',
            'text-center'
        )
    })

    it('has proper heading hierarchy', () => {
        render(<Home />)

        const heading = screen.getByRole('heading', { name: /welcome to odeys/i })
        expect(heading).toHaveClass('text-4xl', 'font-bold', 'mb-8')
    })

    it('has proper description styling', () => {
        render(<Home />)

        const description = screen.getByText(/secure authentication and user management platform/i)
        expect(description).toHaveClass('text-muted-foreground', 'text-lg', 'mb-8')
    })

    it('renders all elements in correct order', () => {
        render(<Home />)

        const main = screen.getByRole('main')
        const elements = Array.from(main.children[0].children)

        // Check order: heading, description, button container
        expect(elements[0]).toHaveTextContent(/welcome to odeys/i)
        expect(elements[1]).toHaveTextContent(/secure authentication/i)
        expect(elements[2]).toHaveClass('flex', 'flex-col', 'sm:flex-row')
    })
})
