# Personal Blog Backend Project Instructions

This is the backend API service for a personal blog built with Go. When generating content or code for this project, please consider the following:

## Development Approach
1. **Incremental Development**:
   - Work in small, manageable steps
   - Implement one feature or fix at a time
   - Change only one file per step
   - Always ask for user feedback after each file change
   - Validate each file change works before moving to the next file
   - Never proceed to the next file without explicit user approval

## Project Guidelines
1. Follow idiomatic Go practices and conventions
2. Maintain the existing project structure:
   - `cmd/api` for application entrypoints
   - `internal` for private application code
   - `pkg` for code that could be used by external applications
   - `configs` for configuration files

3. Error handling:
   - Use appropriate error wrapping with context
   - Return descriptive and actionable error messages
   - Avoid panics in production code

4. Database interactions:
   - Use prepared statements for database queries
   - Implement proper connection pooling
   - Consider transaction management for multi-step operations

5. API design:
   - Follow RESTful API conventions
   - Use consistent response formats
   - Implement proper HTTP status codes
   - Document API endpoints with clear request/response examples

6. Authentication & Authorization:
   - Implement secure token-based authentication
   - Use proper password hashing
   - Apply principle of least privilege
   - Validate all user inputs

7. Performance:
   - Optimize database queries
   - Use appropriate caching strategies
   - Consider concurrent processing for independent operations

8. Testing:
   - Write unit tests for business logic
   - Create integration tests for API endpoints
   - Use table-driven tests when appropriate
   - Mock external dependencies

9. Code organization:
   - Keep functions focused and small
   - Use interfaces for flexibility and testability
   - Follow dependency injection patterns
   - Use descriptive naming conventions

10. Logging & Monitoring:
    - Log meaningful events at appropriate levels
    - Include contextual information in logs
    - Consider structured logging for better analysis

11. Security:
    - Sanitize all user inputs
    - Protect against common web vulnerabilities (CSRF, XSS, etc.)
    - Implement rate limiting for public endpoints
    - Use HTTPS for all communications

## Implementation Workflow
1. Start by understanding requirements clearly
2. Break down tasks into small, manageable subtasks
3. Implement changes to one file at a time
4. Test each file change individually
5. Document your changes
6. Request user feedback after each file change before proceeding to the next file
7. Adjust approach based on feedback
8. Only proceed to the next file after receiving explicit approval
