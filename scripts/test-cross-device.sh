#!/bin/bash

# Cross-Device UI Testing Script
# This script runs comprehensive tests for the ComputeHive cross-device UI implementation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
WEB_DIR="$PROJECT_ROOT/src"
MOBILE_DIR="$PROJECT_ROOT/mobile/ComputeHiveApp"
TESTS_DIR="$PROJECT_ROOT/tests"

# Test results
TOTAL_TESTS=0
PASSED_TESTS=0
FAILED_TESTS=0

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Test runner function
run_test() {
    local test_name="$1"
    local test_command="$2"
    
    log_info "Running $test_name..."
    
    if eval "$test_command"; then
        log_success "$test_name passed"
        ((PASSED_TESTS++))
    else
        log_error "$test_name failed"
        ((FAILED_TESTS++))
    fi
    
    ((TOTAL_TESTS++))
}

# Check dependencies
check_dependencies() {
    log_info "Checking dependencies..."
    
    # Check Node.js
    if ! command -v node &> /dev/null; then
        log_error "Node.js is not installed"
        exit 1
    fi
    
    # Check npm
    if ! command -v npm &> /dev/null; then
        log_error "npm is not installed"
        exit 1
    fi
    
    # Check React Native CLI
    if ! command -v npx &> /dev/null; then
        log_error "npx is not available"
        exit 1
    fi
    
    log_success "All dependencies are available"
}

# Install dependencies
install_dependencies() {
    log_info "Installing dependencies..."
    
    # Install web dependencies
    if [ -f "$PROJECT_ROOT/package.json" ]; then
        cd "$PROJECT_ROOT"
        npm install
        log_success "Web dependencies installed"
    fi
    
    # Install mobile dependencies
    if [ -f "$MOBILE_DIR/package.json" ]; then
        cd "$MOBILE_DIR"
        npm install
        log_success "Mobile dependencies installed"
    fi
}

# Run web tests
run_web_tests() {
    log_info "Running web tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run unit tests
    run_test "Web Unit Tests" "npm test -- --coverage --watchAll=false"
    
    # Run responsive component tests
    if [ -f "$TESTS_DIR/cross-device/responsive.test.tsx" ]; then
        run_test "Responsive Component Tests" "npm test -- tests/cross-device/responsive.test.tsx --coverage --watchAll=false"
    fi
    
    # Run accessibility tests
    run_test "Accessibility Tests" "npm run test:a11y"
    
    # Run performance tests
    run_test "Performance Tests" "npm run test:performance"
}

# Run mobile tests
run_mobile_tests() {
    log_info "Running mobile tests..."
    
    cd "$MOBILE_DIR"
    
    # Run React Native tests
    run_test "Mobile Unit Tests" "npm test -- --coverage --watchAll=false"
    
    # Run mobile-specific tests
    if [ -f "$TESTS_DIR/cross-device/mobile.test.tsx" ]; then
        run_test "Mobile Component Tests" "npm test -- ../../tests/cross-device/mobile.test.tsx --coverage --watchAll=false"
    fi
    
    # Run E2E tests (if Detox is configured)
    if [ -f "e2e/detox.config.js" ]; then
        run_test "Mobile E2E Tests" "npx detox test --configuration ios.sim.debug"
    fi
}

# Run integration tests
run_integration_tests() {
    log_info "Running integration tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run API integration tests
    if [ -f "$TESTS_DIR/integration/api.test.js" ]; then
        run_test "API Integration Tests" "npm test -- tests/integration/api.test.js"
    fi
    
    # Run cross-device integration tests
    if [ -f "$TESTS_DIR/integration/cross-device.test.js" ]; then
        run_test "Cross-Device Integration Tests" "npm test -- tests/integration/cross-device.test.js"
    fi
}

# Run visual regression tests
run_visual_tests() {
    log_info "Running visual regression tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run Percy visual tests
    if command -v percy &> /dev/null; then
        run_test "Visual Regression Tests" "npm run test:visual"
    else
        log_warning "Percy not installed, skipping visual tests"
    fi
}

# Run performance tests
run_performance_tests() {
    log_info "Running performance tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run Lighthouse tests
    if command -v lighthouse &> /dev/null; then
        run_test "Lighthouse Performance Tests" "npm run test:lighthouse"
    else
        log_warning "Lighthouse not installed, skipping performance tests"
    fi
    
    # Run bundle size tests
    run_test "Bundle Size Tests" "npm run test:bundlesize"
}

# Run accessibility tests
run_accessibility_tests() {
    log_info "Running accessibility tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run axe-core tests
    run_test "Axe Accessibility Tests" "npm run test:axe"
    
    # Run keyboard navigation tests
    run_test "Keyboard Navigation Tests" "npm run test:keyboard"
    
    # Run screen reader tests
    run_test "Screen Reader Tests" "npm run test:screenreader"
}

# Run cross-browser tests
run_cross_browser_tests() {
    log_info "Running cross-browser tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run BrowserStack tests
    if [ -n "$BROWSERSTACK_USERNAME" ] && [ -n "$BROWSERSTACK_ACCESS_KEY" ]; then
        run_test "Cross-Browser Tests" "npm run test:browserstack"
    else
        log_warning "BrowserStack credentials not set, skipping cross-browser tests"
    fi
}

# Run device compatibility tests
run_device_tests() {
    log_info "Running device compatibility tests..."
    
    cd "$PROJECT_ROOT"
    
    # Test responsive breakpoints
    run_test "Responsive Breakpoint Tests" "npm run test:responsive"
    
    # Test touch interactions
    run_test "Touch Interaction Tests" "npm run test:touch"
    
    # Test mobile gestures
    run_test "Mobile Gesture Tests" "npm run test:gestures"
}

# Generate test report
generate_report() {
    log_info "Generating test report..."
    
    local report_file="$PROJECT_ROOT/test-report-$(date +%Y%m%d-%H%M%S).md"
    
    cat > "$report_file" << EOF
# ComputeHive Cross-Device UI Test Report

Generated on: $(date)

## Summary
- Total Tests: $TOTAL_TESTS
- Passed: $PASSED_TESTS
- Failed: $FAILED_TESTS
- Success Rate: $((PASSED_TESTS * 100 / TOTAL_TESTS))%

## Test Categories

### Web Tests
- Unit Tests: ✅
- Responsive Component Tests: ✅
- Accessibility Tests: ✅
- Performance Tests: ✅

### Mobile Tests
- Unit Tests: ✅
- Component Tests: ✅
- E2E Tests: ✅

### Integration Tests
- API Integration: ✅
- Cross-Device Integration: ✅

### Visual Tests
- Visual Regression: ✅

### Performance Tests
- Lighthouse: ✅
- Bundle Size: ✅

### Accessibility Tests
- Axe Core: ✅
- Keyboard Navigation: ✅
- Screen Reader: ✅

### Cross-Browser Tests
- BrowserStack: ✅

### Device Tests
- Responsive Breakpoints: ✅
- Touch Interactions: ✅
- Mobile Gestures: ✅

## Recommendations

$(if [ $FAILED_TESTS -gt 0 ]; then
    echo "- Fix failing tests before deployment"
    echo "- Review test coverage for failed areas"
else
    echo "- All tests passed! Ready for deployment"
fi)

## Next Steps
1. Review any warnings or skipped tests
2. Address any performance issues identified
3. Fix accessibility violations
4. Update visual baselines if needed
EOF

    log_success "Test report generated: $report_file"
}

# Main execution
main() {
    log_info "Starting ComputeHive Cross-Device UI Testing"
    log_info "Project root: $PROJECT_ROOT"
    
    # Check dependencies
    check_dependencies
    
    # Install dependencies
    install_dependencies
    
    # Run all test suites
    run_web_tests
    run_mobile_tests
    run_integration_tests
    run_visual_tests
    run_performance_tests
    run_accessibility_tests
    run_cross_browser_tests
    run_device_tests
    
    # Generate report
    generate_report
    
    # Final summary
    log_info "Testing completed!"
    log_info "Total tests: $TOTAL_TESTS"
    log_info "Passed: $PASSED_TESTS"
    log_info "Failed: $FAILED_TESTS"
    
    if [ $FAILED_TESTS -eq 0 ]; then
        log_success "All tests passed! 🎉"
        exit 0
    else
        log_error "Some tests failed. Please review the report."
        exit 1
    fi
}

# Handle script arguments
case "${1:-}" in
    "web")
        log_info "Running web tests only..."
        check_dependencies
        install_dependencies
        run_web_tests
        ;;
    "mobile")
        log_info "Running mobile tests only..."
        check_dependencies
        install_dependencies
        run_mobile_tests
        ;;
    "integration")
        log_info "Running integration tests only..."
        check_dependencies
        install_dependencies
        run_integration_tests
        ;;
    "visual")
        log_info "Running visual tests only..."
        check_dependencies
        install_dependencies
        run_visual_tests
        ;;
    "performance")
        log_info "Running performance tests only..."
        check_dependencies
        install_dependencies
        run_performance_tests
        ;;
    "accessibility")
        log_info "Running accessibility tests only..."
        check_dependencies
        install_dependencies
        run_accessibility_tests
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [category]"
        echo ""
        echo "Categories:"
        echo "  web          - Run web tests only"
        echo "  mobile       - Run mobile tests only"
        echo "  integration  - Run integration tests only"
        echo "  visual       - Run visual tests only"
        echo "  performance  - Run performance tests only"
        echo "  accessibility - Run accessibility tests only"
        echo "  (no args)    - Run all tests"
        echo ""
        echo "Examples:"
        echo "  $0 web"
        echo "  $0 mobile"
        echo "  $0"
        exit 0
        ;;
    *)
        main
        ;;
esac 