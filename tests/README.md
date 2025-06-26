# ComputeHive Testing Framework

## Overview

ComputeHive employs a comprehensive testing strategy covering unit tests, integration tests, end-to-end tests, performance tests, security tests, and chaos engineering tests.

## Test Structure

```
tests/
├── unit/               # Unit tests for individual components
├── integration/        # Integration tests with dependencies
├── e2e/               # End-to-end tests with Cypress
├── performance/       # Performance and load tests
├── security/          # Security and vulnerability tests
├── chaos/             # Chaos engineering tests
└── README.md          # This file
```

## Running Tests

### Unit Tests

```bash
# Run all unit tests
go test ./tests/unit/...

# Run with coverage
go test -cover ./tests/unit/...

# Run specific test
go test -run TestAuthService_Login ./tests/unit/
```

### Integration Tests

```bash
# Start test environment
docker-compose -f tests/integration/docker-compose.test.yml up -d

# Run integration tests
go test -tags=integration ./tests/integration/...

# Cleanup
docker-compose -f tests/integration/docker-compose.test.yml down
```

### End-to-End Tests

```bash
# Install Cypress
npm install -g cypress

# Run Cypress tests
cd tests/e2e
cypress run

# Open Cypress UI
cypress open
```

### Performance Tests

```bash
# Install k6
brew install k6

# Run load test
k6 run tests/performance/k6_load_test.js

# Run with custom settings
k6 run --vus 100 --duration 30s tests/performance/k6_load_test.js

# Run with environment variables
k6 run -e BASE_URL=https://staging.computehive.io tests/performance/k6_load_test.js
```

### Security Tests

```bash
# Install dependencies
pip install -r tests/security/requirements.txt

# Run security tests
python tests/security/security_tests.py

# Run specific test class
pytest tests/security/security_tests.py::TestSQLInjection -v
```

### Chaos Engineering Tests

```bash
# Install dependencies
pip install docker psutil pytest

# Run chaos tests (requires elevated privileges)
sudo python tests/chaos/chaos_tests.py

# Run specific chaos scenario
python tests/chaos/chaos_tests.py -k test_network_packet_loss
```

## Test Coverage

### Unit Test Coverage Goals
- Core Services: >90%
- Agent Code: >85%
- SDKs: >80%
- Utilities: >75%

### Integration Test Coverage
- API Endpoints: 100%
- Service Communication: 100%
- Database Operations: >90%
- Message Queue Operations: >90%

## Continuous Integration

### GitHub Actions Workflow

```yaml
name: Tests

on: [push, pull_request]

jobs:
  unit-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.21'
      - run: go test -v ./tests/unit/...

  integration-tests:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: test
      redis:
        image: redis:7
    steps:
      - uses: actions/checkout@v3
      - run: docker-compose -f tests/integration/docker-compose.test.yml up -d
      - run: go test -tags=integration ./tests/integration/...

  e2e-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: cypress-io/github-action@v5
        with:
          working-directory: tests/e2e
          start: npm run start:test
          wait-on: 'http://localhost:3000'

  security-tests:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-python@v4
        with:
          python-version: '3.11'
      - run: |
          pip install -r tests/security/requirements.txt
          python tests/security/security_tests.py
```

## Test Data Management

### Test Fixtures
- Located in `tests/fixtures/`
- JSON files for API responses
- SQL scripts for database setup
- Docker images for service mocking

### Test Database
- Use separate test database
- Reset between test runs
- Seed with known data

## Performance Benchmarks

### API Response Times
- Authentication: < 50ms (p99)
- Job Submission: < 100ms (p99)
- Resource Query: < 30ms (p99)
- Marketplace Search: < 200ms (p99)

### Throughput Targets
- 10,000 concurrent users
- 1,000 jobs/second submission rate
- 50,000 resource queries/second

## Security Testing Checklist

- [x] SQL Injection
- [x] XSS Prevention
- [x] Authentication Bypass
- [x] Authorization Checks
- [x] Rate Limiting
- [x] Input Validation
- [x] File Upload Security
- [x] API Security Headers
- [x] Cryptography Implementation
- [x] Session Management

## Chaos Engineering Scenarios

### Network Chaos
- Packet loss (5%, 10%, 20%)
- Network latency (100ms, 500ms, 1s)
- Bandwidth throttling
- Network partitioning

### Service Chaos
- Random service kills
- Service pausing
- CPU/Memory stress
- Disk space exhaustion

### Application Chaos
- Database connection drops
- Cache failures
- Message queue issues
- API error injection

## Test Reporting

### Coverage Reports
- Generated in `coverage/` directory
- HTML reports for visualization
- Integration with CI/CD

### Performance Reports
- k6 generates HTML reports
- Grafana dashboards for monitoring
- Historical trend analysis

### Security Reports
- OWASP compliance report
- Vulnerability scan results
- Penetration test findings

## Best Practices

1. **Write Tests First**: Follow TDD principles
2. **Keep Tests Fast**: Unit tests should run in milliseconds
3. **Isolate Dependencies**: Use mocks and stubs
4. **Clean Test Data**: Always cleanup after tests
5. **Descriptive Names**: Test names should explain what they test
6. **Avoid Flaky Tests**: Ensure tests are deterministic
7. **Regular Updates**: Keep test dependencies updated

## Troubleshooting

### Common Issues

1. **Port Conflicts**
   ```bash
   # Check ports
   lsof -i :8080
   # Kill process
   kill -9 <PID>
   ```

2. **Docker Issues**
   ```bash
   # Reset Docker
   docker system prune -a
   # Restart Docker daemon
   sudo systemctl restart docker
   ```

3. **Test Database Issues**
   ```bash
   # Reset test database
   docker-compose -f tests/integration/docker-compose.test.yml down -v
   docker-compose -f tests/integration/docker-compose.test.yml up -d
   ```

## Contributing

When adding new features:
1. Write unit tests first
2. Add integration tests for API changes
3. Update e2e tests for UI changes
4. Add performance tests for critical paths
5. Consider security implications
<<<<<<< HEAD
6. Document test scenarios 
=======
6. Document test scenarios
>>>>>>> 4c40309e804c8f522625b7fd70da67d8d7383849
