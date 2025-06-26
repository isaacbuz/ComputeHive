#!/bin/bash
set -e

echo "🚀 Setting up ComputeHive project..."

# Create missing directories
mkdir -p bin logs data scripts

# Create package.json for the root project
cat > package.json << 'EOF'
{
  "name": "computehive",
  "version": "1.0.0",
  "description": "Distributed compute platform",
  "private": true,
  "scripts": {
    "setup": "npm install && cd agent && go mod init github.com/computehive/agent && cd ../core-services && go mod init github.com/computehive/core-services",
    "build:agent": "cd agent && go build -o ../bin/computehive-agent ./cmd/agent || echo 'Agent build will be available after implementation'",
    "dev": "docker-compose up -d",
    "test": "echo 'Tests will be added with implementation'"
  },
  "devDependencies": {
    "@types/node": "^20.8.0"
  }
}
EOF

# Create basic docker-compose.yml
cat > docker-compose.yml << 'EOF'
version: '3.8'

services:
  # CockroachDB - Distributed SQL Database
  cockroachdb:
    image: cockroachdb/cockroach:latest-v23.1
    command: start-single-node --insecure --listen-addr=0.0.0.0
    ports:
      - "26257:26257"
      - "8080:8080"
    environment:
      - COCKROACH_DATABASE=computehive
    volumes:
      - cockroach-data:/cockroach/cockroach-data
    networks:
      - computehive-net

  # Redis - Cache and Session Store
  redis:
    image: redis:7-alpine
    command: redis-server --appendonly yes
    ports:
      - "6379:6379"
    volumes:
      - redis-data:/data
    networks:
      - computehive-net

  # MinIO - S3-compatible Object Storage
  minio:
    image: minio/minio:latest
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9000:9000"
      - "9001:9001"
    volumes:
      - minio-data:/data
    networks:
      - computehive-net

networks:
  computehive-net:
    driver: bridge

volumes:
  cockroach-data:
  redis-data:
  minio-data:
EOF

echo "✅ Setup files created"
echo ""
echo "📋 Next steps:"
echo "1. Run: npm install"
echo "2. Start services: docker-compose up -d"
echo "3. Begin implementing components"
echo ""
echo "Ready to begin development!" 