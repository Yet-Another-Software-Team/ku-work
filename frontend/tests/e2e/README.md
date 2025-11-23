# E2E Tests with Google OAuth Test Sessions

This directory contains E2E tests for the KU-Work application using Playwright with authenticated test sessions.

## Overview

The tests are organized using Playwright's project feature with setup dependencies:

- **Setup Projects**: Handle authentication and save session state
  - `company-auth.setup.ts` - Authenticates as a company user
  - `student-auth.setup.ts` - Authenticates as a student user

- **Test Projects**: Use authenticated sessions to test functionality
  - `company.spec.ts` - Company user workflows
  - `student.spec.ts` - Student user workflows

## Prerequisites

1. **Test Accounts**: Create Google accounts for testing
   - One for company user testing
   - One for student user testing

2. **Environment Variables**: Copy `.env.test.example` to `.env.test` and fill in credentials:
   ```bash
   cp sample.env.test .env.test
   ```

   Then edit `.env.test` with your test account credentials:
   ```env
   TEST_COMPANY_EMAIL=your-company-test@gmail.com
   TEST_COMPANY_PASSWORD=YourPassword123
   TEST_STUDENT_EMAIL=your-student-test@gmail.com
   TEST_STUDENT_PASSWORD=YourPassword123
   ```

3. **Dependencies**: Ensure all dependencies are installed:
   ```bash
   bun install
   ```

## Running Tests

### Run all tests
```bash
bun run test:e2e
```

### Run tests in UI mode (interactive)
```bash
bun run test:e2e:ui
```

### Run specific project tests
```bash
# Only company tests
npx playwright test --project=company

# Only student tests
npx playwright test --project=student

# Only setup tests (to refresh auth state)
npx playwright test --project=setup-company
npx playwright test --project=setup-student
```

### Run in headed mode (see browser)
```bash
npx playwright test --headed
```

### Debug tests
```bash
npx playwright test --debug
```

## How It Works

### Authentication Flow

1. **Setup Phase**: 
   - `company-auth.setup.ts` and `student-auth.setup.ts` run first
   - They perform Google OAuth login via popup
   - Authentication state (cookies, localStorage) is saved to `playwright/.auth/`

2. **Test Phase**:
   - Test files (`company.spec.ts`, `student.spec.ts`) load the saved auth state
   - Tests run with authenticated sessions
   - No need to login again for each test

### Test Session Files

Authentication states are stored in:
- `playwright/.auth/company.json` - Company user session
- `playwright/.auth/student.json` - Student user session

These files are created during setup and reused across tests. They are gitignored.

## Test Structure

### Company Tests (`company.spec.ts`)
- ✅ Dashboard access and job postings view
- ✅ Job status tabs interaction (Accepted, Pending, Rejected)
- ✅ Profile page access
- ✅ Job search/filtering
- ✅ Navigation menu functionality
- ✅ User authentication verification
- ✅ Logout functionality

### Student Tests (`student.spec.ts`)
- ✅ Dashboard access and applications view
- ✅ Application status tabs interaction
- ✅ Profile page access
- ✅ Job browsing and viewing
- ✅ Job search functionality
- ✅ Company profile viewing
- ✅ Navigation menu functionality
- ✅ User authentication verification
- ✅ Application history and details
- ✅ Application filtering/sorting
- ✅ Logout functionality

## Troubleshooting

### Authentication fails
- Verify your test account credentials in `.env.test`
- Check if Google requires 2FA (disable for test accounts)
- Ensure test accounts are pre-registered in the application

### Timeout errors
- Google OAuth popup may take time to load
- Increase timeout in setup files if needed
- Check network connectivity

### Tests fail after setup
- Delete `playwright/.auth/` directory and re-run setup
- Verify the application is running on `http://localhost:3000`

### Re-authenticate
If you need to refresh authentication state:
```bash
# Delete existing auth state
rm -rf playwright/.auth/

# Re-run setup only
npx playwright test --project=setup-company --project=setup-student
```

## Configuration

Test configuration is in `playwright.config.ts`. Key settings:

- **Base URL**: `http://localhost:3000`
- **Headless**: `false` (can see browser during tests)
- **Screenshots**: Captured on failure
- **Video**: Recorded on failure
- **Retries**: Configurable for CI/CD

## CI/CD Integration

For CI/CD pipelines:

1. Store test credentials as secrets
2. Set `CI=true` environment variable for headless mode
3. Consider using Playwright's service for cloud testing
4. Save test artifacts (videos, traces) on failure

## Best Practices

1. **Isolation**: Each test should be independent
2. **Cleanup**: Tests should clean up any created data
3. **Waits**: Use `waitForLoadState()` and `waitForTimeout()` appropriately
4. **Selectors**: Use semantic selectors (roles, labels) when possible
5. **Assertions**: Include meaningful error messages

## Notes

- Setup tests run once before their dependent test projects
- Authentication state is shared across all tests in a project
- Tests run in parallel by default (can be configured)
- Browser contexts are isolated per test for security
