# Unit Testing Guide

This document provides guidelines for writing and running unit tests for the TaiPhanVan Blog Backend API.

## Test Organization

Tests follow the Go standard convention of being placed in the same package as the code they test, with a `_test.go` suffix. We use the `testing` package together with the `stretchr/testify` suite for writing tests.

## Running Tests

Run all tests:

```bash
go test ./...
```

Run tests with verbose output:

```bash
go test ./... -v
```

Run tests with coverage information:

```bash
go test ./... -cover
```

Generate a detailed HTML coverage report:

```bash
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

Run tests for a specific package:

```bash
go test github.com/phanvantai/taiphanvan_backend/pkg/utils
```

## Test Structure

We use two approaches to structuring tests:

1. **Table-driven tests**: These are used for simple unit tests of functions with clear inputs and outputs, like utility functions.

2. **Test suites**: These are used for more complex tests that require setup and teardown, like API endpoints or database interactions.

### Table-Driven Test Example

```go
func TestFormatDate(t *testing.T) {
    testCases := []struct {
        name     string
        input    time.Time
        expected string
    }{
        {
            name:     "Basic date format",
            input:    time.Date(2025, 5, 15, 10, 0, 0, 0, time.UTC),
            expected: "May 15, 2025",
        },
        // More test cases...
    }

    for _, tc := range testCases {
        t.Run(tc.name, func(t *testing.T) {
            result := utils.FormatDate(tc.input)
            assert.Equal(t, tc.expected, result)
        })
    }
}
```

### Test Suite Example

```go
type AuthTestSuite struct {
    suite.Suite
    router *gin.Engine
    db     *gorm.DB
}

func (s *AuthTestSuite) SetupSuite() {
    // Setup code runs once before all tests
}

func (s *AuthTestSuite) TearDownSuite() {
    // Teardown code runs once after all tests
}

func (s *AuthTestSuite) SetupTest() {
    // Runs before each test
}

func (s *AuthTestSuite) TestLogin() {
    // Test code
}

func TestAuthSuite(t *testing.T) {
    suite.Run(t, new(AuthTestSuite))
}
```

## Testing Best Practices

1. **Isolation**: Each test should be independent and not rely on the state from other tests.

2. **In-memory database**: Use SQLite in memory for database tests to ensure tests are fast and don't affect real data.

3. **Mock external dependencies**: Use mocks for external services (like Cloudinary) to avoid making real API calls during tests.

4. **Test both success and failure paths**: Ensure you test both the happy path and error conditions.

5. **Clear assertions**: Use descriptive assertions so that test failures provide clear information about what went wrong.

6. **Descriptive test names**: Use clear, descriptive names for your tests to make it obvious what functionality is being tested.

7. **Don't test the framework**: Focus on testing your application logic, not the behavior of external libraries.

## Test Coverage

Aim for high test coverage, especially for critical business logic, but remember that 100% coverage doesn't guarantee bug-free code. Focus on testing:

- Business logic
- Error handling
- Edge cases
- Security-critical code

## Adding New Tests

When adding new features, consider writing tests first (TDD approach) or at least in parallel with the implementation. For existing code without tests, prioritize adding tests for:

1. Critical business logic
2. Areas with known bugs
3. Complex algorithms
4. Public APIs

## Continuous Integration

Tests are automatically run as part of the CI pipeline. Ensure all tests pass before submitting a pull request.
