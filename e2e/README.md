# End-to-End Tests

This directory contains comprehensive end-to-end tests for the Evidence Wall application using Playwright.

## Test Structure

### Test Files

- **`auth.spec.ts`** - Authentication flow tests (login, register, logout, session management)
- **`boards.spec.ts`** - Board management tests (creation, editing, items, connections)
- **`collaboration.spec.ts`** - Real-time collaboration tests (multi-user, permissions, conflicts)
- **`public-boards.spec.ts`** - Public board access and sharing tests

### Utilities

- **`utils/auth-helpers.ts`** - Authentication helper functions
- **`utils/board-helpers.ts`** - Board interaction helper functions
- **`fixtures/test-data.ts`** - Test data and constants

## Running Tests

### Prerequisites

1. Ensure the development environment is set up:
   ```bash
   make setup
   ```

2. Install Playwright browsers:
   ```bash
   npm run test:e2e:install
   ```

### Running All Tests

```bash
# Run all e2e tests
make test-e2e

# Or using npm
npm run test:e2e
```

### Running Specific Test Suites

```bash
# Run only authentication tests
npx playwright test auth.spec.ts

# Run only board tests
npx playwright test boards.spec.ts

# Run only collaboration tests
npx playwright test collaboration.spec.ts

# Run only public board tests
npx playwright test public-boards.spec.ts
```

### Interactive Testing

```bash
# Run tests with UI mode (interactive)
make test-e2e-ui
# or
npm run test:e2e:ui

# Run tests in debug mode
make test-e2e-debug
# or
npm run test:e2e:debug

# Run tests in headed mode (see browser)
npm run test:e2e:headed
```

### Test Reports

```bash
# View test report
npm run test:e2e:report
```

## Test Configuration

The tests are configured in `playwright.config.ts` with the following features:

- **Multi-browser testing**: Chrome, Firefox, Safari, Mobile Chrome, Mobile Safari
- **Automatic server startup**: Starts backend services and frontend automatically
- **Parallel execution**: Tests run in parallel for faster execution
- **Screenshots and videos**: Captured on test failures
- **Trace collection**: Detailed traces for debugging failures

## Test Data

Test data is defined in `fixtures/test-data.ts` and includes:

- **Test users**: Pre-defined user accounts for testing
- **Test boards**: Sample board configurations
- **Test board items**: Sample post-it notes, suspect cards, and evidence items
- **API endpoints**: Centralized endpoint definitions

## Helper Functions

### AuthHelper

Provides authentication-related functionality:

- `register(user)` - Register a new user
- `login(user)` - Log in an existing user
- `logout()` - Log out current user
- `ensureLoggedIn(user)` - Ensure user is logged in
- `ensureLoggedOut()` - Ensure user is logged out

### BoardHelper

Provides board interaction functionality:

- `createBoard(board)` - Create a new board
- `openBoard(boardId)` - Navigate to a board
- `addPostItNote(item)` - Add a post-it note
- `addSuspectCard(item)` - Add a suspect card
- `connectItems(from, to)` - Create connections between items
- `dragItem(item, position)` - Drag and drop items
- `deleteItem(item)` - Delete board items
- `shareBoard(email, permission)` - Share board with users

## Test Scenarios Covered

### Authentication
- User registration with validation
- User login with various scenarios
- Session persistence and management
- Protected route access control
- User profile management

### Board Management
- Board creation (private/public)
- Board navigation and listing
- Board settings and permissions
- Board search and filtering
- Board deletion

### Board Items
- Adding different types of items (post-its, suspects, evidence)
- Item editing and content management
- Item positioning and drag-and-drop
- Item deletion and management
- Color and styling options

### Connections
- Creating connections between items
- Connection visualization
- Connection deletion
- Multiple connection management

### Real-time Collaboration
- Multi-user simultaneous editing
- Live cursor tracking
- Real-time item updates
- Permission-based collaboration
- Conflict resolution
- User connection/disconnection handling

### Public Boards
- Anonymous access to public boards
- Public board restrictions (read-only)
- SEO and sharing features
- Real-time updates for anonymous users
- Performance with multiple viewers

## Best Practices

1. **Test Isolation**: Each test is independent and doesn't rely on other tests
2. **Data Management**: Tests create their own data and clean up after themselves
3. **Wait Strategies**: Proper waiting for elements and network requests
4. **Error Handling**: Graceful handling of expected failures and edge cases
5. **Parallel Execution**: Tests are designed to run in parallel safely

## Debugging Tests

### Common Issues

1. **Timing Issues**: Use proper wait strategies instead of fixed timeouts
2. **Element Selection**: Use data-testid attributes for reliable element selection
3. **Network Requests**: Ensure proper waiting for API responses
4. **Browser State**: Tests should not depend on browser state from other tests

### Debug Commands

```bash
# Run specific test in debug mode
npx playwright test auth.spec.ts --debug

# Run with headed browser
npx playwright test --headed

# Generate trace for failed tests
npx playwright test --trace on
```

### Test Development

When adding new tests:

1. Add appropriate data-testid attributes to components
2. Use helper functions for common operations
3. Follow the existing test structure and patterns
4. Add test data to fixtures when needed
5. Update this README with new test scenarios

## CI/CD Integration

The tests are configured to run in CI environments with:

- Retry logic for flaky tests
- Proper browser installation
- Artifact collection (screenshots, videos, traces)
- Parallel execution optimization

## Performance Considerations

- Tests run in parallel by default
- Browser contexts are reused when possible
- Test data is optimized for speed
- Network requests are minimized through proper mocking when appropriate
