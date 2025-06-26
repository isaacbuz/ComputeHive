#!/bin/bash
set -e

echo "🚀 Setting up ComputeHive development environment..."

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to check if command exists
command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check prerequisites
echo "📋 Checking prerequisites..."

if ! command_exists go; then
    echo -e "${RED}❌ Go is not installed. Please install Go 1.21+ from https://golang.org${NC}"
    exit 1
fi

if ! command_exists node; then
    echo -e "${RED}❌ Node.js is not installed. Please install Node.js 18+ from https://nodejs.org${NC}"
    exit 1
fi

if ! command_exists docker; then
    echo -e "${RED}❌ Docker is not installed. Please install Docker from https://docker.com${NC}"
    exit 1
fi

echo -e "${GREEN}✅ All prerequisites are installed${NC}"

# Create necessary directories
echo "📁 Creating directory structure..."
mkdir -p bin
mkdir -p logs
mkdir -p data

# Install Go dependencies
echo "📦 Installing Go dependencies..."
cd agent && go mod download && cd ..
cd core-services && go mod download && cd ..

# Install Node dependencies
echo "📦 Installing Node dependencies..."
npm install

# Install web dashboard dependencies
if [ -d "web/dashboard" ]; then
    cd web/dashboard && npm install && cd ../..
fi

# Install contracts dependencies
if [ -d "contracts" ]; then
    cd contracts && npm install && cd ..
fi

# Install SDK dependencies
if [ -d "sdk/javascript" ]; then
    cd sdk/javascript && npm install && cd ../..
fi

# Start Docker services
echo "🐳 Starting Docker services..."
docker-compose up -d

# Wait for services to be ready
echo "⏳ Waiting for services to be ready..."
sleep 10

# Check if services are running
if docker-compose ps | grep -q "Up"; then
    echo -e "${GREEN}✅ Docker services are running${NC}"
else
    echo -e "${RED}❌ Failed to start Docker services${NC}"
    exit 1
fi

# Create default configuration
echo "⚙️ Creating default configuration..."
mkdir -p ~/.computehive
cat > ~/.computehive/agent.yaml << EOF
control_plane_url: http://localhost:8080
heartbeat_interval: 30s
max_concurrent_jobs: 5
resource_limits:
  max_cpu_percent: 80
  max_memory_percent: 80
  max_disk_percent: 90
security:
  enable_tls: false
  enable_attestation: false
log_level: info
log_format: console
EOF

echo -e "${GREEN}✅ Configuration created at ~/.computehive/agent.yaml${NC}"

# Build the agent
echo "🔨 Building ComputeHive agent..."
cd agent
go build -o ../bin/computehive-agent cmd/agent/main.go
cd ..

if [ -f "bin/computehive-agent" ]; then
    echo -e "${GREEN}✅ Agent built successfully${NC}"
else
    echo -e "${RED}❌ Failed to build agent${NC}"
    exit 1
fi

# Display status
echo ""
echo "🎉 ComputeHive development environment is ready!"
echo ""
echo "📊 Service Status:"
docker-compose ps
echo ""
echo "🚀 Quick Start Commands:"
echo "  - Run agent: ./bin/computehive-agent start"
echo "  - View logs: docker-compose logs -f"
echo "  - Run tests: npm test"
echo "  - Access dashboard: http://localhost:3000"
echo "  - Access Grafana: http://localhost:3000 (admin/admin)"
echo "  - Access MinIO: http://localhost:9001 (computehive/computehive123)"
echo ""
echo "📖 For more information, see README.md" 