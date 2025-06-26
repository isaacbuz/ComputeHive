#!/bin/bash

# Cross-Device UI Deployment Script
# This script builds and deploys the ComputeHive cross-device UI

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
BUILD_DIR="$PROJECT_ROOT/build"
DIST_DIR="$PROJECT_ROOT/dist"

# Environment variables
ENVIRONMENT="${ENVIRONMENT:-production}"
BUILD_NUMBER="${BUILD_NUMBER:-$(date +%Y%m%d-%H%M%S)}"
DEPLOY_TARGET="${DEPLOY_TARGET:-all}"

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

# Check prerequisites
check_prerequisites() {
    log_info "Checking prerequisites..."
    
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
    
    # Check Expo CLI
    if ! command -v expo &> /dev/null; then
        log_warning "Expo CLI not found, installing..."
        npm install -g @expo/cli
    fi
    
    # Check Docker (for containerized deployment)
    if ! command -v docker &> /dev/null; then
        log_warning "Docker not found, some deployment options may not be available"
    fi
    
    log_success "Prerequisites check completed"
}

# Clean build directories
clean_build_dirs() {
    log_info "Cleaning build directories..."
    
    rm -rf "$BUILD_DIR"
    rm -rf "$DIST_DIR"
    rm -rf "$WEB_DIR/build"
    rm -rf "$MOBILE_DIR/build"
    
    mkdir -p "$BUILD_DIR"
    mkdir -p "$DIST_DIR"
    
    log_success "Build directories cleaned"
}

# Install dependencies
install_dependencies() {
    log_info "Installing dependencies..."
    
    # Install web dependencies
    cd "$PROJECT_ROOT"
    npm ci --production=false
    
    # Install mobile dependencies
    cd "$MOBILE_DIR"
    npm ci --production=false
    
    log_success "Dependencies installed"
}

# Run tests
run_tests() {
    log_info "Running tests..."
    
    cd "$PROJECT_ROOT"
    
    # Run web tests
    npm test -- --coverage --watchAll=false
    
    # Run mobile tests
    cd "$MOBILE_DIR"
    npm test -- --coverage --watchAll=false
    
    log_success "All tests passed"
}

# Build web application
build_web() {
    log_info "Building web application..."
    
    cd "$PROJECT_ROOT"
    
    # Set environment variables
    export REACT_APP_ENVIRONMENT="$ENVIRONMENT"
    export REACT_APP_BUILD_NUMBER="$BUILD_NUMBER"
    
    # Build the application
    npm run build
    
    # Optimize build
    npm run build:optimize
    
    # Generate service worker
    npm run build:sw
    
    # Copy build to dist directory
    cp -r build/* "$DIST_DIR/web/"
    
    log_success "Web application built successfully"
}

# Build mobile application
build_mobile() {
    log_info "Building mobile application..."
    
    cd "$MOBILE_DIR"
    
    # Set environment variables
    export EXPO_ENVIRONMENT="$ENVIRONMENT"
    export EXPO_BUILD_NUMBER="$BUILD_NUMBER"
    
    # Build for iOS
    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "ios" ]; then
        log_info "Building for iOS..."
        expo build:ios --non-interactive --no-wait
    fi
    
    # Build for Android
    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "android" ]; then
        log_info "Building for Android..."
        expo build:android --non-interactive --no-wait
    fi
    
    # Build for web (PWA)
    if [ "$DEPLOY_TARGET" = "all" ] || [ "$DEPLOY_TARGET" = "web" ]; then
        log_info "Building mobile web version..."
        expo build:web
        cp -r web-build/* "$DIST_DIR/mobile-web/"
    fi
    
    log_success "Mobile application built successfully"
}

# Build Docker containers
build_docker() {
    log_info "Building Docker containers..."
    
    cd "$PROJECT_ROOT"
    
    # Build web container
    docker build -f Dockerfile.web -t computehive-web:$BUILD_NUMBER .
    
    # Build mobile container (if needed)
    docker build -f Dockerfile.mobile -t computehive-mobile:$BUILD_NUMBER .
    
    log_success "Docker containers built successfully"
}

# Deploy to CDN
deploy_cdn() {
    log_info "Deploying to CDN..."
    
    cd "$PROJECT_ROOT"
    
    # Deploy web assets to CDN
    if [ -n "$CDN_BUCKET" ]; then
        aws s3 sync "$DIST_DIR/web/" s3://$CDN_BUCKET/ --delete
        aws cloudfront create-invalidation --distribution-id $CLOUDFRONT_DISTRIBUTION_ID --paths "/*"
        log_success "Web assets deployed to CDN"
    else
        log_warning "CDN_BUCKET not set, skipping CDN deployment"
    fi
}

# Deploy to app stores
deploy_app_stores() {
    log_info "Deploying to app stores..."
    
    cd "$MOBILE_DIR"
    
    # Deploy to App Store Connect
    if [ -n "$APP_STORE_CONNECT_API_KEY" ]; then
        expo upload:ios --latest
        log_success "iOS app uploaded to App Store Connect"
    else
        log_warning "APP_STORE_CONNECT_API_KEY not set, skipping iOS deployment"
    fi
    
    # Deploy to Google Play Console
    if [ -n "$GOOGLE_PLAY_SERVICE_ACCOUNT" ]; then
        expo upload:android --latest
        log_success "Android app uploaded to Google Play Console"
    else
        log_warning "GOOGLE_PLAY_SERVICE_ACCOUNT not set, skipping Android deployment"
    fi
}

# Deploy to Kubernetes
deploy_kubernetes() {
    log_info "Deploying to Kubernetes..."
    
    cd "$PROJECT_ROOT"
    
    # Apply Kubernetes manifests
    kubectl apply -f k8s/namespace.yaml
    kubectl apply -f k8s/configmap.yaml
    kubectl apply -f k8s/secrets.yaml
    kubectl apply -f k8s/deployment.yaml
    kubectl apply -f k8s/service.yaml
    kubectl apply -f k8s/ingress.yaml
    
    # Update deployment with new image
    kubectl set image deployment/computehive-web computehive-web=computehive-web:$BUILD_NUMBER -n computehive
    
    # Wait for rollout
    kubectl rollout status deployment/computehive-web -n computehive
    
    log_success "Kubernetes deployment completed"
}

# Deploy to Vercel
deploy_vercel() {
    log_info "Deploying to Vercel..."
    
    cd "$PROJECT_ROOT"
    
    # Deploy web application
    vercel --prod
    
    log_success "Vercel deployment completed"
}

# Deploy to Netlify
deploy_netlify() {
    log_info "Deploying to Netlify..."
    
    cd "$PROJECT_ROOT"
    
    # Deploy web application
    netlify deploy --prod --dir=build
    
    log_success "Netlify deployment completed"
}

# Run health checks
run_health_checks() {
    log_info "Running health checks..."
    
    # Check web application
    if [ -n "$WEB_URL" ]; then
        response=$(curl -s -o /dev/null -w "%{http_code}" "$WEB_URL")
        if [ "$response" = "200" ]; then
            log_success "Web application health check passed"
        else
            log_error "Web application health check failed: HTTP $response"
        fi
    fi
    
    # Check API endpoints
    if [ -n "$API_URL" ]; then
        response=$(curl -s -o /dev/null -w "%{http_code}" "$API_URL/health")
        if [ "$response" = "200" ]; then
            log_success "API health check passed"
        else
            log_error "API health check failed: HTTP $response"
        fi
    fi
    
    log_success "Health checks completed"
}

# Generate deployment report
generate_deployment_report() {
    log_info "Generating deployment report..."
    
    local report_file="$PROJECT_ROOT/deployment-report-$(date +%Y%m%d-%H%M%S).md"
    
    cat > "$report_file" << EOF
# ComputeHive Cross-Device UI Deployment Report

Generated on: $(date)

## Deployment Information
- Environment: $ENVIRONMENT
- Build Number: $BUILD_NUMBER
- Deploy Target: $DEPLOY_TARGET

## Build Artifacts

### Web Application
- Build Directory: $DIST_DIR/web/
- Bundle Size: $(du -sh $DIST_DIR/web/ | cut -f1)
- Files: $(find $DIST_DIR/web/ -type f | wc -l)

### Mobile Application
- iOS Build: $(if [ -f "$DIST_DIR/mobile/ios" ]; then echo "✅"; else echo "❌"; fi)
- Android Build: $(if [ -f "$DIST_DIR/mobile/android" ]; then echo "✅"; else echo "❌"; fi)
- Web Build: $(if [ -d "$DIST_DIR/mobile-web" ]; then echo "✅"; else echo "❌"; fi)

## Deployment Status

### Web Deployment
- CDN: $(if [ -n "$CDN_BUCKET" ]; then echo "✅"; else echo "❌"; fi)
- Kubernetes: $(if [ -n "$KUBECONFIG" ]; then echo "✅"; else echo "❌"; fi)
- Vercel: $(if command -v vercel &> /dev/null; then echo "✅"; else echo "❌"; fi)
- Netlify: $(if command -v netlify &> /dev/null; then echo "✅"; else echo "❌"; fi)

### Mobile Deployment
- App Store: $(if [ -n "$APP_STORE_CONNECT_API_KEY" ]; then echo "✅"; else echo "❌"; fi)
- Google Play: $(if [ -n "$GOOGLE_PLAY_SERVICE_ACCOUNT" ]; then echo "✅"; else echo "❌"; fi)

## Performance Metrics
- Web Bundle Size: $(du -sh $DIST_DIR/web/static/js/ | cut -f1)
- Mobile Bundle Size: $(if [ -f "$DIST_DIR/mobile/android/app-release.apk" ]; then du -sh $DIST_DIR/mobile/android/app-release.apk | cut -f1; else echo "N/A"; fi)
- Build Time: $(($(date +%s) - $(date -d "$BUILD_START_TIME" +%s))) seconds

## Health Check Results
- Web Application: $(if [ -n "$WEB_URL" ]; then echo "✅"; else echo "❌"; fi)
- API Endpoints: $(if [ -n "$API_URL" ]; then echo "✅"; else echo "❌"; fi)

## Next Steps
1. Monitor application performance
2. Check error logs
3. Verify user feedback
4. Plan next deployment

## Rollback Information
- Previous Build: $PREVIOUS_BUILD_NUMBER
- Rollback Command: ./scripts/rollback.sh $BUILD_NUMBER
EOF

    log_success "Deployment report generated: $report_file"
}

# Main deployment function
main() {
    log_info "Starting ComputeHive Cross-Device UI Deployment"
    log_info "Environment: $ENVIRONMENT"
    log_info "Build Number: $BUILD_NUMBER"
    log_info "Deploy Target: $DEPLOY_TARGET"
    
    # Record build start time
    BUILD_START_TIME=$(date)
    
    # Run deployment steps
    check_prerequisites
    clean_build_dirs
    install_dependencies
    run_tests
    build_web
    build_mobile
    build_docker
    
    # Deploy based on environment
    case "$ENVIRONMENT" in
        "production")
            deploy_cdn
            deploy_app_stores
            deploy_kubernetes
            ;;
        "staging")
            deploy_vercel
            deploy_netlify
            ;;
        "development")
            log_info "Development deployment - skipping production deployments"
            ;;
        *)
            log_error "Unknown environment: $ENVIRONMENT"
            exit 1
            ;;
    esac
    
    # Run health checks
    run_health_checks
    
    # Generate report
    generate_deployment_report
    
    log_success "Deployment completed successfully! 🚀"
}

# Handle script arguments
case "${1:-}" in
    "web")
        log_info "Deploying web application only..."
        check_prerequisites
        clean_build_dirs
        install_dependencies
        run_tests
        build_web
        deploy_cdn
        run_health_checks
        ;;
    "mobile")
        log_info "Deploying mobile application only..."
        check_prerequisites
        clean_build_dirs
        install_dependencies
        run_tests
        build_mobile
        deploy_app_stores
        ;;
    "docker")
        log_info "Building Docker containers only..."
        check_prerequisites
        build_docker
        ;;
    "kubernetes")
        log_info "Deploying to Kubernetes only..."
        deploy_kubernetes
        ;;
    "help"|"-h"|"--help")
        echo "Usage: $0 [target]"
        echo ""
        echo "Targets:"
        echo "  web         - Deploy web application only"
        echo "  mobile      - Deploy mobile application only"
        echo "  docker      - Build Docker containers only"
        echo "  kubernetes  - Deploy to Kubernetes only"
        echo "  (no args)   - Deploy everything"
        echo ""
        echo "Environment Variables:"
        echo "  ENVIRONMENT - Deployment environment (production|staging|development)"
        echo "  BUILD_NUMBER - Build number (default: timestamp)"
        echo "  DEPLOY_TARGET - Deploy target (all|web|mobile|ios|android)"
        echo ""
        echo "Examples:"
        echo "  $0 web"
        echo "  $0 mobile"
        echo "  ENVIRONMENT=staging $0"
        exit 0
        ;;
    *)
        main
        ;;
esac 