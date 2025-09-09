# Renclave-v2 GitHub Actions Dependency Flow

## Corrected Job Dependencies

```
┌─────────────────┐    ┌─────────────────┐
│  quick-checks   │    │   unit-tests    │
│  (fmt, clippy)  │    │  (local tests)  │
└─────────┬───────┘    └─────────┬───────┘
          │                      │
          └──────────┬───────────┘
                     │
          ┌──────────▼───────────┐
          │  integration-tests   │
          │  (Docker tests)      │
          └──────────┬───────────┘
                     │
          ┌──────────▼───────────┐
          │     e2e-tests        │
          │  (End-to-end tests)  │
          └──────────────────────┘

┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│     build       │    │    security     │    │    coverage     │
│ (Release build) │    │   (Security     │    │  (Code coverage)│
│                 │    │    scanning)    │    │                 │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         └───────────────────────┼───────────────────────┘
                                 │
                    ┌────────────▼────────────┐
                    │       summary           │
                    │   (Test Summary)        │
                    └─────────────────────────┘
```

## Dependency Logic

### Level 1: Fast Parallel Jobs
- **quick-checks**: Code formatting and linting (fast)
- **unit-tests**: Local unit tests (fast)

### Level 2: Integration Tests
- **integration-tests**: Depends on quick-checks (ensures code quality first)

### Level 3: End-to-End Tests
- **e2e-tests**: Depends on integration-tests (integration before e2e)

### Level 4: Advanced Analysis (Parallel)
- **build**: Depends on quick-checks + unit-tests (basic validation first)
- **security**: Depends on quick-checks + unit-tests (basic validation first)
- **coverage**: Depends on quick-checks + unit-tests (basic validation first)

### Level 5: Summary
- **summary**: Depends on all jobs (always runs, shows final status)

## Benefits of This Flow

1. **Fast Feedback**: Quick checks run first and in parallel
2. **Logical Progression**: Code quality → Unit tests → Integration → E2E
3. **Efficient Resource Usage**: Parallel execution where possible
4. **Early Failure Detection**: Basic validation before expensive operations
5. **Comprehensive Coverage**: All test types run with proper dependencies

## Job Execution Order

1. **Parallel**: `quick-checks` + `unit-tests`
2. **Sequential**: `integration-tests` (after quick-checks)
3. **Sequential**: `e2e-tests` (after integration-tests)
4. **Parallel**: `build` + `security` + `coverage` (after quick-checks + unit-tests)
5. **Final**: `summary` (after all jobs complete)
