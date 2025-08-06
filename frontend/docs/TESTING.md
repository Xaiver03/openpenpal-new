# Testing Guide

## Overview

This document provides comprehensive information about the testing strategy, setup, and execution for the OpenPenPal project.

## Testing Strategy

Our testing approach follows a pyramid structure:

```
        /\
       /  \
      /E2E \    <- End-to-End Tests (Playwright)
     /______\
    /        \
   /Integration\ <- API Integration Tests (Jest)
  /__________\
 /            \
/  Unit Tests  \ <- Component & Utility Tests (Jest + RTL)
/______________\
```

## Test Types

### 1. Unit Tests
- **Location**: `src/**/__tests__/*.test.ts`
- **Framework**: Jest + React Testing Library
- **Coverage**: Individual functions, components, and utilities
- **Run**: `npm test`

### 2. Integration Tests
- **Location**: `src/app/api/__tests__/*.integration.test.ts`
- **Framework**: Jest with API mocking
- **Coverage**: API endpoints, middleware, and service interactions
- **Run**: `npm test -- --testPathPattern=integration`

### 3. End-to-End Tests
- **Location**: `tests/e2e/tests/*.spec.ts`
- **Framework**: Playwright
- **Coverage**: Complete user workflows and browser interactions
- **Run**: `npm run test:e2e`

## Setup Instructions

### Prerequisites
```bash
# Install dependencies
npm install

# Install Playwright browsers
npm run test:e2e:install
```

### Environment Setup
```bash
# Copy environment template
cp .env.example .env.test

# Configure test database
export TEST_DATABASE_URL="postgresql://testuser:testpass@localhost:5432/openpenpal_test"
export TEST_REDIS_URL="redis://localhost:6379/1"
```

## Running Tests

### Local Development
```bash
# Run all unit tests
npm test

# Run tests in watch mode
npm run test:watch

# Run tests with coverage
npm run test:coverage

# Run E2E tests
npm run test:e2e

# Run E2E tests with UI
npm run test:e2e:ui

# Run E2E tests in headed mode
npm run test:e2e:headed
```

### CI/CD Pipeline
Tests are automatically executed in GitHub Actions:

1. **Unit Tests**: On every push and PR
2. **Integration Tests**: On every push and PR
3. **E2E Tests**: On every push to main
4. **Security Tests**: Daily scheduled scans
5. **Performance Tests**: On production deployments

## Test Configuration

### Jest Configuration
```javascript
// jest.config.js
module.exports = {
  testEnvironment: 'jsdom',
  setupFilesAfterEnv: ['<rootDir>/jest.setup.js'],
  moduleNameMapping: {
    '^@/(.*)$': '<rootDir>/src/$1'
  },
  collectCoverageFrom: [
    'src/**/*.{js,jsx,ts,tsx}',
    '!src/**/*.d.ts',
    '!src/**/index.ts'
  ],
  coverageThreshold: {
    global: {
      branches: 70,
      functions: 70,
      lines: 70,
      statements: 70
    }
  }
}
```

### Playwright Configuration
```typescript
// playwright.config.ts
export default defineConfig({
  testDir: './tests/e2e',
  fullyParallel: true,
  retries: process.env.CI ? 2 : 0,
  use: {
    baseURL: 'http://localhost:3000',
    trace: 'on-first-retry',
    screenshot: 'only-on-failure'
  },
  projects: [
    { name: 'chromium', use: devices['Desktop Chrome'] },
    { name: 'firefox', use: devices['Desktop Firefox'] },
    { name: 'webkit', use: devices['Desktop Safari'] }
  ]
})
```

## Writing Tests

### Unit Test Example
```typescript
// src/components/__tests__/Button.test.tsx
import { render, screen, fireEvent } from '@testing-library/react'
import { Button } from '../Button'

describe('Button Component', () => {
  test('renders button with text', () => {
    render(<Button>Click me</Button>)
    expect(screen.getByRole('button')).toHaveTextContent('Click me')
  })

  test('calls onClick handler when clicked', () => {
    const handleClick = jest.fn()
    render(<Button onClick={handleClick}>Click me</Button>)
    
    fireEvent.click(screen.getByRole('button'))
    expect(handleClick).toHaveBeenCalledTimes(1)
  })
})
```

### Integration Test Example
```typescript
// src/app/api/__tests__/auth.integration.test.ts
import { POST as loginHandler } from '../auth/login/route'
import { NextRequest } from 'next/server'

describe('Authentication API', () => {
  test('successful login returns user data', async () => {
    const request = new NextRequest('http://localhost/api/auth/login', {
      method: 'POST',
      body: JSON.stringify({
        username: 'testuser',
        password: 'validpassword'
      })
    })

    const response = await loginHandler(request)
    const data = await response.json()

    expect(response.status).toBe(200)
    expect(data.success).toBe(true)
    expect(data.data).toHaveProperty('user')
    expect(data.data).toHaveProperty('accessToken')
  })
})
```

### E2E Test Example
```typescript
// tests/e2e/tests/auth.spec.ts
import { test, expect } from '@playwright/test'
import { LoginPage } from '../pages/login.page'

test.describe('Authentication Flow', () => {
  test('user can login with valid credentials', async ({ page }) => {
    const loginPage = new LoginPage(page)
    
    await loginPage.goto()
    await loginPage.login('testuser', 'validpassword')
    
    await expect(page).toHaveURL('/dashboard')
    await expect(page.getByTestId('user-avatar')).toBeVisible()
  })
})
```

## Test Data Management

### Test Users
```typescript
// tests/e2e/fixtures/test-users.ts
export const TEST_USERS = {
  regularUser: {
    username: 'testuser',
    password: 'TestPass123!',
    email: 'testuser@example.com'
  },
  courier: {
    username: 'testcourier',
    password: 'CourierPass123!',
    email: 'courier@example.com'
  }
}
```

### Database Seeding
```typescript
// src/lib/auth/test-data-manager.ts
export class TestDataManager {
  static async initializeTestUsers() {
    // Create test users in database
    // This runs before E2E tests
  }
  
  static async cleanupTestData() {
    // Clean up test data after tests
  }
}
```

## Mocking Strategies

### API Mocking
```typescript
// Unit tests - Mock external services
jest.mock('@/lib/services/auth-service', () => ({
  AuthService: {
    login: jest.fn(),
    logout: jest.fn()
  }
}))
```

### Network Mocking (E2E)
```typescript
// E2E tests - Intercept network requests
await page.route('**/api/auth/login', route => {
  route.fulfill({
    status: 200,
    body: JSON.stringify({ success: true })
  })
})
```

## Coverage Reports

### Generating Coverage
```bash
# Generate coverage report
npm run test:coverage

# View HTML coverage report
open coverage/lcov-report/index.html
```

### Coverage Thresholds
- **Branches**: 70%
- **Functions**: 70%  
- **Lines**: 70%
- **Statements**: 70%

## Debugging Tests

### Unit Tests
```bash
# Debug specific test
npm test -- --testNamePattern="should login successfully" --verbose

# Debug with Node debugger
node --inspect-brk node_modules/.bin/jest --runInBand
```

### E2E Tests
```bash
# Debug with Playwright Inspector
npm run test:e2e:debug

# Run in headed mode
npm run test:e2e:headed

# Generate trace files
npm run test:e2e -- --trace on
```

## Best Practices

### Test Organization
1. **Arrange-Act-Assert**: Structure tests clearly
2. **One assertion per test**: Keep tests focused
3. **Descriptive names**: Use clear test descriptions
4. **Page Object Model**: For E2E tests, use page objects

### Test Data
1. **Isolated tests**: Each test should be independent
2. **Clean setup/teardown**: Reset state between tests
3. **Realistic data**: Use data that matches production
4. **Minimal fixtures**: Only create necessary test data

### Assertions
1. **Specific assertions**: Use precise matchers
2. **User-centric**: Test from user's perspective
3. **Accessibility**: Include accessibility tests
4. **Error states**: Test error conditions

### Performance
1. **Parallel execution**: Run tests concurrently when possible
2. **Smart waiting**: Use proper wait strategies in E2E tests
3. **Test categorization**: Group tests by type and execution time
4. **Selective testing**: Run relevant tests based on changes

## Continuous Integration

### GitHub Actions Workflows
- **test.yml**: Main test suite workflow
- **security.yml**: Security scanning workflow
- **deploy.yml**: Deployment with testing

### Test Execution Matrix
```yaml
strategy:
  matrix:
    node-version: [18.x, 20.x]
    browser: [chromium, firefox, webkit]
```

### Reporting
- **Test Results**: JUnit XML format
- **Coverage**: LCOV format uploaded to Codecov
- **E2E Results**: HTML report with screenshots and videos
- **Security Scans**: SARIF format for GitHub Security tab

## Troubleshooting

### Common Issues

#### Jest Issues
```bash
# Clear Jest cache
npm test -- --clearCache

# Update snapshots
npm test -- --updateSnapshot
```

#### Playwright Issues
```bash
# Update browsers
npx playwright install

# Clear browser data
rm -rf ~/.cache/ms-playwright
```

#### CI/CD Issues
- Check environment variables are set
- Verify database connections
- Review artifact uploads
- Check test timeouts

## Resources

- [Jest Documentation](https://jestjs.io/docs/getting-started)
- [React Testing Library](https://testing-library.com/docs/react-testing-library/intro/)
- [Playwright Documentation](https://playwright.dev/docs/intro)
- [Testing Best Practices](https://kentcdodds.com/blog/common-mistakes-with-react-testing-library)

## Support

For testing-related questions:
1. Check this documentation
2. Review existing test examples
3. Create an issue in the project repository
4. Contact the development team