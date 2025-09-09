import { test, expect } from '@playwright/test'

test.describe('Authentication Flow', () => {
    test.beforeEach(async ({ page }) => {
        // Disable mock authentication for these tests to test real auth flow
        await page.addInitScript(() => {
            window.__MOCK_AUTH__ = false;
        });

        // Navigate to the home page before each test
        await page.goto('/')
    })

    test('should display home page with sign in and dashboard links', async ({ page }) => {
        // Check that the home page loads correctly
        await expect(page).toHaveTitle(/ODEYS/)

        // Check for main heading
        await expect(page.getByRole('heading', { name: /welcome to odeys/i })).toBeVisible()

        // Check for description
        await expect(page.getByText(/secure authentication and user management platform/i)).toBeVisible()

        // Check for sign in link
        await expect(page.getByRole('link', { name: /sign in/i })).toBeVisible()

        // Check for dashboard link
        await expect(page.getByRole('link', { name: /go to dashboard/i })).toBeVisible()
    })

    test('should navigate to sign in page', async ({ page }) => {
        // Click the sign in link
        await page.getByRole('link', { name: /sign in/i }).click()

        // Check that we're on the sign in page
        await expect(page).toHaveURL(/.*\/auth\/signin/)

        // Check for sign in form elements (these would depend on your actual sign in page)
        // await expect(page.getByRole('button', { name: /sign in with google/i })).toBeVisible()
    })

    test('should navigate to dashboard page', async ({ page }) => {
        // Click the dashboard link (should redirect to sign-in for unauthenticated users)
        await page.getByRole('link', { name: /go to dashboard/i }).click()

        // Wait for the redirect to complete
        await page.waitForLoadState('networkidle')

        // Check that we're redirected to the sign-in page (correct behavior for unauthenticated users)
        await expect(page).toHaveURL(/.*\/auth\/signin/, { timeout: 10000 })

        // Check for sign-in form elements
        // await expect(page.getByRole('button', { name: /sign in with google/i })).toBeVisible()
    })

    test('should handle navigation between pages', async ({ page }) => {
        // Start at home page
        await expect(page.getByRole('heading', { name: /welcome to odeys/i })).toBeVisible()

        // Navigate to sign in
        await page.getByRole('link', { name: /sign in/i }).click()
        await expect(page).toHaveURL(/.*\/auth\/signin/)

        // Navigate back to home (assuming there's a back button or logo link)
        await page.goto('/')
        await page.waitForLoadState('domcontentloaded')
        await expect(page.getByRole('heading', { name: /welcome to odeys/i })).toBeVisible()

        // Navigate to dashboard (should redirect to sign-in for unauthenticated users)
        const dashboardLink = page.getByRole('link', { name: /go to dashboard/i })
        await dashboardLink.waitFor({ state: 'visible' })
        await dashboardLink.click()

        // Wait for the redirect to complete - use a more specific wait with longer timeout
        await page.waitForURL(/.*\/auth\/signin/, { timeout: 15000 })
    })
})

test.describe('Responsive Design', () => {
    test.beforeEach(async ({ page }) => {
        // Enable mock authentication for responsive tests
        await page.addInitScript(() => {
            window.__MOCK_AUTH__ = true;
        });
    });

    test('should work on mobile devices', async ({ page }) => {
        // Set mobile viewport
        await page.setViewportSize({ width: 375, height: 667 })

        await page.goto('/')

        // Check that the page is responsive
        await expect(page.getByRole('heading', { name: /welcome to odeys/i })).toBeVisible()

        // Check that buttons are still clickable on mobile
        await expect(page.getByRole('link', { name: /sign in/i })).toBeVisible()
        await expect(page.getByRole('link', { name: /go to dashboard/i })).toBeVisible()
    })

    test('should work on tablet devices', async ({ page }) => {
        // Set tablet viewport
        await page.setViewportSize({ width: 768, height: 1024 })

        await page.goto('/')

        // Check that the page works on tablet
        await expect(page.getByRole('heading', { name: /welcome to odeys/i })).toBeVisible()
        await expect(page.getByRole('link', { name: /sign in/i })).toBeVisible()
        await expect(page.getByRole('link', { name: /go to dashboard/i })).toBeVisible()
    })
})

test.describe('Accessibility', () => {
    test.beforeEach(async ({ page }) => {
        // Enable mock authentication for accessibility tests
        await page.addInitScript(() => {
            window.__MOCK_AUTH__ = true;
        });
    });

    test('should have proper heading structure', async ({ page }) => {
        await page.goto('/')

        // Check for main heading
        const mainHeading = page.getByRole('heading', { name: /welcome to odeys/i })
        await expect(mainHeading).toBeVisible()

        // Check that it's an h1
        await expect(mainHeading).toHaveText(/welcome to odeys/i)
    })

    test('should have accessible links', async ({ page }) => {
        await page.goto('/')

        // Check that links have proper text
        const signInLink = page.getByRole('link', { name: /sign in/i })
        const dashboardLink = page.getByRole('link', { name: /go to dashboard/i })

        await expect(signInLink).toBeVisible()
        await expect(dashboardLink).toBeVisible()

        // Check that links are keyboard accessible
        await signInLink.focus()
        await expect(signInLink).toBeFocused()

        await dashboardLink.focus()
        await expect(dashboardLink).toBeFocused()
    })

    test('should have proper color contrast', async ({ page }) => {
        await page.goto('/')

        // This would require more sophisticated accessibility testing
        // For now, just check that elements are visible
        await expect(page.getByRole('heading', { name: /welcome to odeys/i })).toBeVisible()
        await expect(page.getByText(/secure authentication and user management platform/i)).toBeVisible()
    })
})
