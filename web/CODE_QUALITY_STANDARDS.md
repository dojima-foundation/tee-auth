# Code Quality Standards

## Overview
This document outlines the code quality standards and best practices for the TEE Auth web application.

## ESLint Configuration

### Current Rules
- **Next.js Core Web Vitals**: Ensures optimal performance and user experience
- **TypeScript**: Strict type checking and modern TypeScript practices
- **Jest Overrides**: Allows CommonJS imports in Jest configuration files

### Key Rules Enforced
- `@typescript-eslint/no-explicit-any`: Prevents use of `any` type
- `@typescript-eslint/no-unused-vars`: Ensures all variables are used
- `react-hooks/exhaustive-deps`: Enforces proper React hooks dependencies
- `@typescript-eslint/ban-ts-comment`: Prefers `@ts-expect-error` over `@ts-ignore`

## Type Safety Standards

### Type Definitions
- Use `unknown` instead of `any` for unknown data types
- Define proper interfaces for component props
- Use type assertions only when necessary with proper comments

### React Hooks
- Wrap async functions in `useCallback` to prevent infinite re-renders
- Include all dependencies in useEffect and useCallback dependency arrays
- Use proper TypeScript types for hook parameters

## Testing Standards

### Mock Components
- Define proper TypeScript interfaces for mock component props
- Add `displayName` to mock components for better debugging
- Use `unknown` type for mock data instead of `any`

### Test Utilities
- Use typed mock factories for consistent test data
- Properly type store utilities and test helpers
- Comment out unused mock data instead of deleting (for future reference)

## Code Organization

### File Structure
- Components in `src/components/`
- Tests in `src/__tests__/`
- Utilities in `src/__tests__/utils/`
- Types in `src/types/`

### Import Management
- Remove unused imports to keep code clean
- Use specific imports instead of wildcard imports
- Group imports logically (React, third-party, local)

## Performance Considerations

### React Optimization
- Use `useCallback` for functions passed as props
- Use `useMemo` for expensive calculations
- Proper dependency arrays to prevent unnecessary re-renders

### Bundle Optimization
- Tree-shake unused code
- Lazy load components when appropriate
- Optimize imports to reduce bundle size

## Maintenance Guidelines

### Regular Tasks
- Run `npm run lint` before committing
- Fix ESLint warnings and errors promptly
- Update type definitions when APIs change
- Review and update test mocks regularly

### Code Review Checklist
- [ ] No `any` types used
- [ ] All React hooks have proper dependencies
- [ ] No unused variables or imports
- [ ] Proper TypeScript interfaces defined
- [ ] Tests are properly typed
- [ ] Mock components have display names

## Tools and Commands

### Development
```bash
npm run lint          # Check code quality
npm run lint:fix      # Auto-fix linting issues
npm test              # Run unit tests
npm run test:coverage # Run tests with coverage
npm run test:e2e      # Run end-to-end tests
```

### Quality Assurance
```bash
npm run build         # Build for production
npm run type-check    # TypeScript type checking
```

## Future Improvements

### Planned Enhancements
- Add more specific ESLint rules for accessibility
- Implement stricter TypeScript configuration
- Add performance monitoring rules
- Enhance test coverage requirements

### Monitoring
- Track ESLint error trends
- Monitor bundle size changes
- Review test coverage reports
- Analyze performance metrics

---

*Last updated: $(date)*
*Maintained by: Development Team*
