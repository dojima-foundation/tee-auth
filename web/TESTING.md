# Testing Documentation

This document provides comprehensive information about the testing strategy, patterns, and best practices used in this project.

## Table of Contents

1. [Testing Strategy](#testing-strategy)
2. [Test Types](#test-types)
3. [Testing Patterns](#testing-patterns)
4. [Configuration](#configuration)
5. [Best Practices](#best-practices)
6. [Troubleshooting](#troubleshooting)

## Testing Strategy

Our testing strategy follows the testing pyramid approach:

```
    /\
   /  \     E2E Tests (Few)
  /____\    Integration Tests (Some)
 /______\   Unit Tests (Many)
```

### Testing Layers

1. **Unit Tests (70%)**: Test individual functions, components, and utilities
2. **Integration Tests (20%)**: Test component interactions and API integrations
3. **E2E Tests (10%)**: Test complete user workflows

### Testing Principles

- **Fast**: Unit tests should run in milliseconds
- **Reliable**: Tests should be deterministic and not flaky
- **Isolated**: Each test should be independent
- **Maintainable**: Tests should be easy to understand and modify
- **Comprehensive**: Cover critical user paths and edge cases

## Test Types

### 1. Unit Tests

Unit tests verify that individual functions, components, or utilities work correctly in isolation.

**Location**: `src/__tests__/`

**Examples**:
- Component rendering and behavior
- Hook logic and state management
- Utility function calculations
- Form validation logic

**Tools**: Jest + React Testing Library

### 2. Integration Tests

Integration tests verify that multiple components work together correctly.

**Location**: `src/__tests__/integration/`

**Examples**:
- Component with API calls
- Form submission with validation
- State management across components
- Router navigation

**Tools**: Jest + React Testing Library + MSW

### 3. E2E Tests

E2E tests verify complete user workflows from start to finish.

**Location**: `e2e/specs/`

**Examples**:
- User registration and login
- Complete form workflows
- Navigation between pages
- Error handling scenarios

**Tools**: Playwright

### 4. Visual Regression Tests

Visual regression tests ensure UI consistency across changes.

**Location**: `e2e/specs/visual-regression.spec.ts`

**Examples**:
- Screenshot comparisons
- Responsive design verification
- Dark/light mode consistency

**Tools**: Playwright + Percy

### 5. Performance Tests

Performance tests monitor Core Web Vitals and performance metrics.

**Location**: `e2e/specs/performance.spec.ts`

**Examples**:
- Page load times
- Core Web Vitals measurement
- Resource loading optimization
- Memory usage monitoring

**Tools**: Playwright + Lighthouse CI

### 6. Accessibility Tests

Accessibility tests ensure the application is accessible to all users.

**Examples**:
- ARIA attribute verification
- Keyboard navigation
- Screen reader compatibility
- Color contrast validation

**Tools**: Playwright + axe-core

## Testing Patterns

### Component Testing Pattern

```typescript
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { Component } from './Component';

describe('Component', () => {
  // Arrange
  const defaultProps = {
    // Define default props
  };

  // Test rendering
  it('renders correctly', () => {
    render(<Component {...defaultProps} />);
    expect(screen.getByRole('button')).toBeInTheDocument();
  });

  // Test user interactions
  it('handles user interactions', async () => {
    const user = userEvent.setup();
    const handleClick = jest.fn();
    
    render(<Component {...defaultProps} onClick={handleClick} />);
    await user.click(screen.getByRole('button'));
    
    expect(handleClick).toHaveBeenCalledTimes(1);
  });

  // Test different states
  it('displays loading state', () => {
    render(<Component {...defaultProps} loading={true} />);
    expect(screen.getByTestId('loading-spinner')).toBeInTheDocument();
  });

  // Test error states
  it('displays error message', () => {
    render(<Component {...defaultProps} error="Something went wrong" />);
    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
  });
});
```

### Hook Testing Pattern

```typescript
import { renderHook, act } from '@testing-library/react';
import { useCustomHook } from './useCustomHook';

describe('useCustomHook', () => {
  it('returns initial state', () => {
    const { result } = renderHook(() => useCustomHook());
    
    expect(result.current.value).toBe(initialValue);
    expect(result.current.isLoading).toBe(false);
  });

  it('updates state when action is called', async () => {
    const { result } = renderHook(() => useCustomHook());
    
    await act(async () => {
      await result.current.updateValue('new value');
    });
    
    expect(result.current.value).toBe('new value');
  });
});
```

### API Integration Testing Pattern

```typescript
import { render, screen, waitFor } from '@testing-library/react';
import { server } from '@/tests/setup';
import { http, HttpResponse } from 'msw';
import { ComponentWithAPI } from './ComponentWithAPI';

describe('ComponentWithAPI', () => {
  it('fetches and displays data', async () => {
    // Mock API response
    server.use(
      http.get('/api/data', () => {
        return HttpResponse.json({ items: [{ id: 1, name: 'Test' }] });
      })
    );

    render(<ComponentWithAPI />);
    
    // Wait for loading to finish
    await waitFor(() => {
      expect(screen.getByText('Test')).toBeInTheDocument();
    });
  });

  it('handles API errors gracefully', async () => {
    // Mock API error
    server.use(
      http.get('/api/data', () => {
        return HttpResponse.json({ error: 'Not found' }, { status: 404 });
      })
    );

    render(<ComponentWithAPI />);
    
    await waitFor(() => {
      expect(screen.getByText('Error loading data')).toBeInTheDocument();
    });
  });
});
```

### E2E Testing Pattern

```typescript
import { test, expect } from '@playwright/test';

test.describe('User Authentication', () => {
  test('user can login successfully', async ({ page }) => {
    // Navigate to login page
    await page.goto('/login');
    
    // Fill login form
    await page.fill('[data-testid="email-input"]', 'user@example.com');
    await page.fill('[data-testid="password-input"]', 'password123');
    
    // Submit form
    await page.click('[data-testid="login-button"]');
    
    // Verify successful login
    await expect(page).toHaveURL('/dashboard');
    await expect(page.locator('[data-testid="user-name"]')).toContainText('Test User');
  });

  test('shows error for invalid credentials', async ({ page }) => {
    await page.goto('/login');
    
    await page.fill('[data-testid="email-input"]', 'invalid@example.com');
    await page.fill('[data-testid="password-input"]', 'wrongpassword');
    await page.click('[data-testid="login-button"]');
    
    await expect(page.locator('[data-testid="error-message"]')).toContainText('Invalid credentials');
  });
});
```

## Configuration

### Jest Configuration

The Jest configuration is defined in `jest.config.js` and includes:

- Next.js integration
- TypeScript support
- Coverage collection
- Test environment setup
- Module name mapping

### Playwright Configuration

The Playwright configuration is defined in `playwright.config.ts` and includes:

- Multiple browser support
- Device emulation
- Parallel test execution
- Screenshot and video capture
- Custom test fixtures

### ESLint Configuration

The ESLint configuration includes testing-specific rules:

- Jest plugin rules
- Testing Library rules
- Jest DOM rules
- Custom overrides for test files

## Best Practices

### 1. Test Organization

- Group related tests using `describe` blocks
- Use descriptive test names that explain the expected behavior
- Follow the Arrange-Act-Assert pattern
- Keep tests focused and test one thing at a time

### 2. Test Data Management

- Use fixtures for consistent test data
- Create factory functions for complex objects
- Avoid hardcoded values in tests
- Use meaningful test data that represents real scenarios

### 3. Mocking Strategy

- Mock external dependencies (APIs, timers, etc.)
- Use MSW for API mocking
- Avoid mocking implementation details
- Mock at the right level (unit vs integration)

### 4. Assertions

- Use specific assertions that test behavior, not implementation
- Prefer user-centric queries (getByRole, getByText)
- Avoid testing implementation details
- Use meaningful error messages

### 5. Performance

- Keep tests fast and focused
- Use `beforeEach` and `afterEach` for setup/cleanup
- Avoid unnecessary async operations
- Use appropriate timeouts

### 6. Accessibility

- Include accessibility testing in your test suite
- Test keyboard navigation
- Verify ARIA attributes
- Test with screen readers when possible

## Troubleshooting

### Common Issues

#### 1. Tests Failing Intermittently

**Symptoms**: Tests pass locally but fail in CI or intermittently

**Solutions**:
- Add proper waiting mechanisms
- Use `waitFor` instead of `setTimeout`
- Ensure proper cleanup in `afterEach`
- Check for race conditions

#### 2. E2E Tests Timing Out

**Symptoms**: E2E tests fail with timeout errors

**Solutions**:
- Increase timeout values in Playwright config
- Add proper waiting for network idle
- Check for slow network conditions
- Verify element selectors are correct

#### 3. Visual Regression Tests Failing

**Symptoms**: Percy tests fail due to minor UI changes

**Solutions**:
- Review changes for intentional UI updates
- Update baseline images if changes are expected
- Check for dynamic content that changes between runs
- Verify viewport sizes are consistent

#### 4. Performance Tests Failing

**Symptoms**: Lighthouse tests fail due to performance regressions

**Solutions**:
- Review recent changes for performance impact
- Check for new dependencies or assets
- Verify network conditions in CI
- Update performance budgets if needed

### Debug Commands

```bash
# Debug Jest tests with verbose output
npm test -- --verbose --detectOpenHandles

# Debug specific test file
npm test -- ComponentName.test.tsx

# Debug E2E tests with headed mode
npm run test:e2e:headed

# Run tests with coverage
npm run test:coverage

# Check test configuration
npm test -- --showConfig
```

### Getting Help

1. Check the test logs for specific error messages
2. Review the testing documentation
3. Check for similar issues in the project history
4. Consult the tool documentation (Jest, Playwright, etc.)
5. Ask for help in the team chat or create an issue

## Resources

- [Jest Documentation](https://jestjs.io/docs/getting-started)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
- [Playwright Documentation](https://playwright.dev/docs/intro)
- [MSW Documentation](https://mswjs.io/docs/)
- [Percy Documentation](https://docs.percy.io/)
- [Lighthouse CI](https://github.com/GoogleChrome/lighthouse-ci)
