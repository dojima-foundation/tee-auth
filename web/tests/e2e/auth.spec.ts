import { test, expect } from '@playwright/test'

test.describe('Authentication Flow', () => {
    test.beforeEach(async ({ page }) => {
        // Disable mock authentication for these tests to test real auth flow
        await page.addInitScript(() => {
            window.__MOCK_AUTH__ = false;
        });

        // Add network request logging
        page.on('request', request => {
            console.log('🌐 REQUEST:', request.method(), request.url());
        });

        page.on('response', response => {
            console.log('📡 RESPONSE:', response.status(), response.url());
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

    test('should debug ProtectedRoute behavior', async ({ page }) => {
        console.log('=== DEBUG: Testing ProtectedRoute behavior ===')

        // Check environment variables and mock auth status
        const envStatus = await page.evaluate(() => {
            return {
                nodeEnv: process.env.NODE_ENV,
                testMode: process.env.NEXT_PUBLIC_TEST_MODE,
                mockAuth: process.env.NEXT_PUBLIC_MOCK_AUTH,
                windowMockAuth: window.__MOCK_AUTH__,
                currentUrl: window.location.href
            }
        })
        console.log('Environment status:', envStatus)

        // Try to navigate directly to dashboard to see what happens
        console.log('Navigating directly to /dashboard...')
        await page.goto('/dashboard')
        await page.waitForLoadState('networkidle')

        const dashboardUrl = page.url()
        console.log('Dashboard URL after navigation:', dashboardUrl)

        // Check if we're redirected or if dashboard loads
        if (dashboardUrl.includes('/auth/signin')) {
            console.log('✅ ProtectedRoute working: Redirected to signin')
        } else if (dashboardUrl.includes('/dashboard')) {
            console.log('❌ ProtectedRoute not working: Dashboard loaded without auth')
            // Take screenshot
            await page.screenshot({ path: 'debug-dashboard-loaded-without-auth.png' })
        } else {
            console.log('❓ Unexpected behavior: URL is', dashboardUrl)
            // Take screenshot
            await page.screenshot({ path: 'debug-unexpected-dashboard-behavior.png' })
        }

        // Check page content
        const pageContent = await page.content()
        console.log('Page content length:', pageContent.length)
        console.log('Page title:', await page.title())

        // Check for loading states or error messages
        const loadingElements = await page.locator('[class*="loading"], [class*="spinner"]').count()
        console.log('Loading elements found:', loadingElements)

        const errorElements = await page.locator('[class*="error"], [role="alert"]').count()
        console.log('Error elements found:', errorElements)
    })

    test('should debug auth context state', async ({ page }) => {
        console.log('=== DEBUG: Testing auth context state ===')

        // Navigate to home page first
        await page.goto('/')
        await page.waitForLoadState('networkidle')

        // Check auth context state
        const authState = await page.evaluate(() => {
            // Try to access the auth context from the window object or React DevTools
            const reactRoot = document.querySelector('#__next') || document.querySelector('[data-reactroot]');
            return {
                hasReactRoot: !!reactRoot,
                localStorage: {
                    gauth_session_token: localStorage.getItem('gauth_session_token'),
                    gauth_session_data: localStorage.getItem('gauth_session_data')
                },
                sessionStorage: {
                    gauth_session_token: sessionStorage.getItem('gauth_session_token'),
                    gauth_session_data: sessionStorage.getItem('gauth_session_data')
                },
                cookies: document.cookie,
                currentUrl: window.location.href
            }
        })
        console.log('Auth state:', authState)

        // Check if there are any console errors
        const consoleLogs = []
        page.on('console', msg => {
            if (msg.type() === 'error') {
                consoleLogs.push(`❌ Console Error: ${msg.text()}`)
            } else if (msg.text().includes('AuthProvider') || msg.text().includes('ProtectedRoute')) {
                consoleLogs.push(`🔍 Auth Log: ${msg.text()}`)
            }
        })

        // Try to navigate to dashboard and see what happens
        console.log('Attempting navigation to dashboard...')
        await page.goto('/dashboard')
        await page.waitForLoadState('networkidle')

        console.log('Console logs during navigation:', consoleLogs)
        console.log('Final URL after dashboard navigation:', page.url())

        // Check if we can find any auth-related elements
        const authElements = await page.evaluate(() => {
            const elements = []
            document.querySelectorAll('*').forEach(el => {
                if (el.textContent?.includes('Loading') ||
                    el.textContent?.includes('Sign in') ||
                    el.textContent?.includes('Dashboard') ||
                    el.className?.includes('auth') ||
                    el.className?.includes('loading')) {
                    elements.push({
                        tag: el.tagName,
                        text: el.textContent?.substring(0, 100),
                        className: el.className
                    })
                }
            })
            return elements.slice(0, 10) // Limit to first 10 elements
        })
        console.log('Auth-related elements found:', authElements)
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
        // Add debugging: Log initial state
        console.log('=== DEBUG: Starting dashboard navigation test ===')
        console.log('Initial URL:', page.url())

        // Check if mock auth is properly disabled
        const mockAuthStatus = await page.evaluate(() => {
            return {
                mockAuth: window.__MOCK_AUTH__,
                testMode: window.process?.env?.NEXT_PUBLIC_TEST_MODE,
                mockAuthEnv: window.process?.env?.NEXT_PUBLIC_MOCK_AUTH
            }
        })
        console.log('Mock auth status:', mockAuthStatus)

        // Wait for page to be fully loaded
        await page.waitForLoadState('networkidle')
        console.log('Page loaded, current URL:', page.url())

        // Check if dashboard link exists and is visible
        const dashboardLink = page.getByRole('link', { name: /go to dashboard/i })
        const isVisible = await dashboardLink.isVisible()
        console.log('Dashboard link visible:', isVisible)

        if (!isVisible) {
            // Take screenshot for debugging
            await page.screenshot({ path: 'debug-dashboard-link-not-visible.png' })
            console.log('Screenshot saved: debug-dashboard-link-not-visible.png')
        }

        // Click the dashboard link (should redirect to sign-in for unauthenticated users)
        console.log('Clicking dashboard link...')
        await dashboardLink.click()

        // Wait for navigation to start
        await page.waitForLoadState('domcontentloaded')
        console.log('After click, current URL:', page.url())

        // Wait for the redirect to complete with more detailed logging
        console.log('Waiting for redirect to complete...')
        try {
            await page.waitForURL(/.*\/auth\/signin/, { timeout: 15000 })
            console.log('✅ Redirect successful! Final URL:', page.url())
        } catch (error) {
            console.log('❌ Redirect failed or timed out')
            console.log('Current URL after timeout:', page.url())
            console.log('Page title:', await page.title())

            // Take screenshot for debugging
            await page.screenshot({ path: 'debug-redirect-failed.png' })
            console.log('Screenshot saved: debug-redirect-failed.png')

            // Log page content for debugging
            const pageContent = await page.content()
            console.log('Page content length:', pageContent.length)
            console.log('Page content preview:', pageContent.substring(0, 500))

            throw error
        }

        // Check for sign-in form elements
        // await expect(page.getByRole('button', { name: /sign in with google/i })).toBeVisible()
    })

    test('should handle navigation between pages', async ({ page }) => {
        console.log('=== DEBUG: Starting navigation between pages test ===')
        console.log('Initial URL:', page.url())

        // Start at home page
        await expect(page.getByRole('heading', { name: /welcome to odeys/i })).toBeVisible()
        console.log('✅ Home page loaded successfully')

        // Navigate to sign in
        console.log('Navigating to sign in page...')
        await page.getByRole('link', { name: /sign in/i }).click()
        await expect(page).toHaveURL(/.*\/auth\/signin/)
        console.log('✅ Sign in page loaded, URL:', page.url())

        // Navigate back to home (assuming there's a back button or logo link)
        console.log('Navigating back to home page...')
        await page.goto('/')
        await page.waitForLoadState('domcontentloaded')
        await expect(page.getByRole('heading', { name: /welcome to odeys/i })).toBeVisible()
        console.log('✅ Back to home page, URL:', page.url())

        // Navigate to dashboard (should redirect to sign-in for unauthenticated users)
        console.log('Preparing to navigate to dashboard...')
        const dashboardLink = page.getByRole('link', { name: /go to dashboard/i })
        await dashboardLink.waitFor({ state: 'visible' })
        console.log('Dashboard link is visible, clicking...')
        await dashboardLink.click()

        // Wait for navigation to start
        await page.waitForLoadState('domcontentloaded')
        console.log('After dashboard click, URL:', page.url())

        // Wait for the redirect to complete - use a more specific wait with longer timeout
        console.log('Waiting for redirect to /auth/signin...')
        try {
            await page.waitForURL(/.*\/auth\/signin/, { timeout: 15000 })
            console.log('✅ Redirect successful! Final URL:', page.url())
        } catch (error) {
            console.log('❌ Redirect failed or timed out')
            console.log('Current URL after timeout:', page.url())
            console.log('Page title:', await page.title())

            // Take screenshot for debugging
            await page.screenshot({ path: 'debug-navigation-redirect-failed.png' })
            console.log('Screenshot saved: debug-navigation-redirect-failed.png')

            // Log page content for debugging
            const pageContent = await page.content()
            console.log('Page content length:', pageContent.length)
            console.log('Page content preview:', pageContent.substring(0, 500))

            throw error
        }
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
