# OpenPenPal Admin System Comprehensive Testing

This comprehensive testing script validates all aspects of the OpenPenPal admin system, ensuring proper functionality, security, and error handling.

## Features Tested

### 1. Authentication & Authorization Testing
- ✅ Admin token validation
- ✅ Invalid token rejection
- ✅ Missing token rejection
- ✅ Non-admin user access denial
- ✅ Super admin role verification

### 2. User Management Testing
- ✅ Get users list with pagination
- ✅ Handle invalid pagination parameters
- ✅ Create test user for management tests
- ✅ Get specific user details
- ✅ Deactivate/reactivate users
- ✅ Update user roles

### 3. Content Moderation Testing
- ✅ Get moderation queue
- ✅ Get moderation statistics
- ✅ Manage sensitive words (CRUD operations)
- ✅ Manage moderation rules (CRUD operations)
- ✅ Manual content review workflow

### 4. System Configuration Testing
- ✅ Get/update system settings
- ✅ Reset system settings
- ✅ Test email configuration
- ✅ Validate configuration persistence

### 5. Courier Management Testing
- ✅ Get pending courier applications
- ✅ Approve/reject courier applications
- ✅ Manage courier hierarchy
- ✅ Validate courier permissions

### 6. Analytics & Reporting Testing
- ✅ Dashboard statistics
- ✅ Recent activity tracking
- ✅ Analytics dashboard data
- ✅ System analytics
- ✅ Report generation and listing

### 7. Security Testing
- ✅ SQL injection protection
- ✅ XSS prevention
- ✅ Large payload handling
- ✅ Rate limiting checks
- ✅ CSRF protection
- ✅ Authorization bypass attempts

### 8. Error Handling Testing
- ✅ 404 error handling
- ✅ Invalid JSON payload handling
- ✅ Missing required fields
- ✅ Invalid content-type headers
- ✅ Server error recovery

## Usage

### Prerequisites
1. Ensure the OpenPenPal backend is running on `localhost:8080`
2. Have a valid admin token (super_admin role)
3. Node.js installed on your system

### Running the Tests

```bash
# Make the script executable (if not already)
chmod +x test-admin-system-comprehensive.js

# Run all tests
node test-admin-system-comprehensive.js

# Or run directly if executable
./test-admin-system-comprehensive.js
```

### Configuration

The script includes built-in configuration at the top:

```javascript
const config = {
    baseURL: 'http://localhost:8080',
    adminToken: 'your-admin-token-here',
    testUser: {
        username: 'test_admin_user',
        email: 'test@admin.com',
        password: 'password123',
        nickname: 'Test Admin User'
    },
    noProxy: 'localhost,127.0.0.1'
};
```

### Environment Variables

Set `NO_PROXY` to bypass proxy issues:
```bash
export NO_PROXY=localhost,127.0.0.1
node test-admin-system-comprehensive.js
```

## Test Output

The script provides detailed output including:

- ✅ **Passed tests**: Successfully completed operations
- ❌ **Failed tests**: Issues that need attention
- ⏭️ **Skipped tests**: Endpoints not yet implemented
- ⏱️ **Performance metrics**: Response times for each test
- 📊 **Summary statistics**: Overall test results
- 💡 **Recommendations**: Suggestions for improvements

### Sample Output

```
🚀 Starting OpenPenPal Admin System Comprehensive Testing
Base URL: http://localhost:8080
Admin Token: eyJhbGciOiJIUzI1NiIs...

=== 1. AUTHENTICATION & AUTHORIZATION TESTING ===
✅ PASSED: Admin Token Validation (45ms)
✅ PASSED: Invalid Token Rejection (23ms)
✅ PASSED: Missing Token Rejection (18ms)
✅ PASSED: Non-Admin User Access Denial (25ms)
✅ PASSED: Super Admin Role Verification (42ms)

=== 2. USER MANAGEMENT TESTING ===
✅ PASSED: Get Users List with Pagination (67ms)
✅ PASSED: Get Users List with Invalid Pagination (45ms)
...

📊 TEST SUMMARY
============================================================
✅ Passed: 42
❌ Failed: 3
⏭️ Skipped: 5
⏱️ Total Time: 2847ms
📈 Success Rate: 93%
```

## Test Categories in Detail

### Authentication & Authorization
Tests the core security model ensuring only authorized admins can access admin functions.

### User Management
Validates CRUD operations for user accounts, role management, and account status changes.

### Content Moderation  
Tests the content filtering system, sensitive word management, and manual review workflows.

### System Configuration
Verifies system settings management, configuration persistence, and service integrations.

### Courier Management
Tests the four-level courier hierarchy system, application workflows, and permission validation.

### Analytics & Reporting
Validates dashboard data, statistics generation, and reporting functionality.

### Security Testing
Comprehensive security validation including injection attacks, XSS prevention, and authorization bypass attempts.

### Error Handling
Tests graceful error handling, validation, and recovery mechanisms.

## Troubleshooting

### Common Issues

1. **Connection Refused**
   - Ensure backend server is running on localhost:8080
   - Check firewall settings

2. **Authentication Failures**
   - Verify admin token is valid and not expired
   - Ensure token has super_admin role

3. **Test Timeouts**
   - Backend might be slow to respond
   - Check database connectivity
   - Verify all services are running

4. **Permission Errors**
   - Admin token might have insufficient permissions
   - Check user role in JWT token

### Debug Mode

Enable verbose logging by modifying the script:
```javascript
// Add at the top of the script
const DEBUG = true;

// Enable request/response logging
if (DEBUG) {
    console.log('Request:', method, path, data);
    console.log('Response:', response);
}
```

## Integration with CI/CD

This script can be integrated into continuous integration pipelines:

```yaml
# GitHub Actions example
- name: Run Admin System Tests
  run: |
    export NO_PROXY=localhost,127.0.0.1
    node test-admin-system-comprehensive.js
  env:
    ADMIN_TOKEN: ${{ secrets.ADMIN_TOKEN }}
```

## Contributing

When adding new admin features:

1. Add corresponding tests to the appropriate category
2. Update this README with new test descriptions
3. Ensure all tests pass before merging
4. Add security tests for new endpoints

## Security Considerations

This script includes security testing but should only be run in:
- Development environments
- Staging environments  
- Authorized penetration testing scenarios

**Never run security tests against production systems without proper authorization.**

## Performance Benchmarking

The script tracks response times and can be used for performance regression testing:

- Fast responses: < 100ms
- Acceptable responses: 100-500ms  
- Slow responses: > 500ms
- Timeout threshold: 10 seconds

Monitor these metrics over time to catch performance regressions.