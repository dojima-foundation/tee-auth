import { defineConfig, devices } from '@playwright/test'

/**
 * Playwright configuration specifically for accessibility tests
 * Optimized for CI/CD environments with reduced browser matrix
 */
export default defineConfig({
    testDir: './tests/e2e',
    /* Run tests in files in parallel */
    fullyParallel: true,
    /* Fail the build on CI if you accidentally left test.only in the source code. */
    forbidOnly: !!process.env.CI,
    /* Retry on CI only */
    retries: process.env.CI ? 2 : 0,
    /* Single worker for CI stability */
    workers: 1,
    /* Increased timeout for CI */
    timeout: process.env.CI ? 45000 : 15000,
    /* Reporter to use. See https://playwright.dev/docs/test-reporters */
    reporter: [
        ['html'],
        ['json', { outputFile: 'test-results/accessibility-results.json' }],
        ['junit', { outputFile: 'test-results/accessibility-results.xml' }],
    ],
    /* Shared settings for all the projects below. See https://playwright.dev/docs/api/class-testoptions. */
    use: {
        /* Base URL to use in actions like `await page.goto('/')`. */
        baseURL: 'http://localhost:3000',

        /* Run in headless mode for CI, headed for local development */
        headless: process.env.CI ? true : false,

        /* Collect trace when retrying the failed test. See https://playwright.dev/docs/trace-viewer */
        trace: 'on-first-retry',

        /* Take screenshot on failure */
        screenshot: 'only-on-failure',

        /* Record video on failure */
        video: 'retain-on-failure',
    },

    /* Reduced browser matrix for accessibility tests - focus on most important browsers */
    projects: [
        {
            name: 'chromium',
            use: {
                ...devices['Desktop Chrome'],
                launchOptions: {
                    args: process.env.CI ? [
                        '--no-sandbox',
                        '--disable-setuid-sandbox',
                        '--disable-dev-shm-usage',
                        '--disable-gpu',
                        '--disable-web-security'
                    ] : []
                }
            },
        },
        // Only include WebKit if not in CI or if explicitly enabled
        ...(process.env.CI && process.env.ENABLE_WEBKIT !== 'true' ? [] : [{
            name: 'webkit',
            use: {
                ...devices['Desktop Safari'],
                // WebKit doesn't support --no-sandbox flag
                launchOptions: {
                    args: process.env.CI ? [] : []
                }
            },
        }]),
        /* Test against mobile viewports. */
        {
            name: 'Mobile Chrome',
            use: {
                ...devices['Pixel 5'],
                launchOptions: {
                    args: process.env.CI ? [
                        '--no-sandbox',
                        '--disable-setuid-sandbox',
                        '--disable-dev-shm-usage',
                        '--disable-gpu'
                    ] : []
                }
            },
        },
    ],

    /* Run your local dev server before starting the tests */
    webServer: {
        command: 'npm run dev',
        url: 'http://localhost:3000',
        reuseExistingServer: !process.env.CI,
        timeout: process.env.CI ? 180 * 1000 : 120 * 1000,
        env: {
            NEXT_PUBLIC_API_URL: process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080',
            NEXT_PUBLIC_GAUTH_API_URL: process.env.NEXT_PUBLIC_GAUTH_API_URL || 'http://localhost:8080',
            NEXT_PUBLIC_GRPC_URL: process.env.NEXT_PUBLIC_GRPC_URL || 'localhost:9090',
            NEXT_PUBLIC_TEST_MODE: 'true',
            NEXT_PUBLIC_MOCK_AUTH: 'true',
        },
    },
})
