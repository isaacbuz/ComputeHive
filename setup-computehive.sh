#!/bin/bash
set -e

echo "🚀 Beginning ComputeHive Implementation..."
echo ""

cd ComputeHive

# Step 1: Create the basic Go module for the agent
echo "📦 Step 1: Setting up Go modules..."
cd agent
cat > go.mod << 'EOF'
module github.com/computehive/agent

go 1.21

require (
	github.com/google/uuid v1.4.0
	github.com/spf13/cobra v1.7.0
	go.uber.org/zap v1.26.0
)
EOF

# Create a simple agent implementation
mkdir -p cmd/agent
cat > cmd/agent/main.go << 'EOF'
package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	fmt.Println("🔷 ComputeHive Agent v1.0.0")
	fmt.Println("Starting agent...")
	
	// Simulate agent startup
	fmt.Println("✓ Initializing hardware detection...")
	time.Sleep(1 * time.Second)
	
	fmt.Println("✓ Connecting to control plane...")
	time.Sleep(1 * time.Second)
	
	fmt.Println("✓ Agent ready to receive jobs!")
	fmt.Println("")
	fmt.Println("Press Ctrl+C to stop")
	
	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	<-sigChan
	
	fmt.Println("\nShutting down agent...")
}
EOF

cd ../..

# Step 2: Create a simple authentication service
echo "📦 Step 2: Setting up authentication service..."
cd core-services
cat > go.mod << 'EOF'
module github.com/computehive/core-services

go 1.21

require (
	github.com/gorilla/mux v1.8.0
	github.com/golang-jwt/jwt/v5 v5.0.0
)
EOF

mkdir -p auth-service
cat > auth-service/main.go << 'EOF'
package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"
	
	"github.com/gorilla/mux"
)

type HealthResponse struct {
	Status    string    `json:"status"`
	Service   string    `json:"service"`
	Timestamp time.Time `json:"timestamp"`
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	response := HealthResponse{
		Status:    "healthy",
		Service:   "auth-service",
		Timestamp: time.Now(),
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func main() {
	fmt.Println("🔐 ComputeHive Auth Service v1.0.0")
	
	router := mux.NewRouter()
	router.HandleFunc("/health", healthHandler).Methods("GET")
	router.HandleFunc("/api/v1/agents/register", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": true,
			"message": "Agent registered successfully",
			"agentId": "agent-" + fmt.Sprintf("%d", time.Now().Unix()),
		})
	}).Methods("POST")
	
	fmt.Println("Auth service listening on :8081")
	log.Fatal(http.ListenAndServe(":8081", router))
}
EOF

cd ../..

# Step 3: Create a React dashboard
echo "📦 Step 3: Setting up React dashboard..."
cd web/dashboard
cat > package.json << 'EOF'
{
  "name": "computehive-dashboard",
  "version": "1.0.0",
  "private": true,
  "dependencies": {
    "react": "^18.2.0",
    "react-dom": "^18.2.0",
    "react-scripts": "5.0.1"
  },
  "scripts": {
    "start": "react-scripts start",
    "build": "react-scripts build"
  },
  "browserslist": {
    "production": [">0.2%", "not dead", "not op_mini all"],
    "development": ["last 1 chrome version", "last 1 firefox version", "last 1 safari version"]
  }
}
EOF

mkdir -p src
cat > src/index.js << 'EOF'
import React from 'react';
import ReactDOM from 'react-dom/client';
import App from './App';

const root = ReactDOM.createRoot(document.getElementById('root'));
root.render(<App />);
EOF

cat > src/App.js << 'EOF'
import React, { useState, useEffect } from 'react';

function App() {
  const [status, setStatus] = useState('Connecting...');
  const [nodes, setNodes] = useState(0);
  
  useEffect(() => {
    // Simulate connecting to backend
    setTimeout(() => {
      setStatus('Connected');
      setNodes(Math.floor(Math.random() * 100) + 1);
    }, 2000);
  }, []);
  
  return (
    <div style={{ padding: '20px', fontFamily: 'Arial, sans-serif' }}>
      <h1>🔷 ComputeHive Dashboard</h1>
      <div style={{ marginTop: '20px' }}>
        <h2>System Status</h2>
        <p>Status: <strong>{status}</strong></p>
        <p>Active Nodes: <strong>{nodes}</strong></p>
        <p>Jobs Processed: <strong>0</strong></p>
      </div>
      <div style={{ marginTop: '40px', padding: '20px', backgroundColor: '#f0f0f0', borderRadius: '8px' }}>
        <h3>Quick Actions</h3>
        <button style={{ margin: '5px', padding: '10px 20px' }}>Submit Job</button>
        <button style={{ margin: '5px', padding: '10px 20px' }}>View Nodes</button>
        <button style={{ margin: '5px', padding: '10px 20px' }}>Billing</button>
      </div>
    </div>
  );
}

export default App;
EOF

mkdir -p public
cat > public/index.html << 'EOF'
<!DOCTYPE html>
<html lang="en">
<head>
  <meta charset="utf-8" />
  <meta name="viewport" content="width=device-width, initial-scale=1" />
  <title>ComputeHive Dashboard</title>
</head>
<body>
  <div id="root"></div>
</body>
</html>
EOF

cd ../..

# Step 4: Create a simple Makefile for easy building
echo "📦 Step 4: Creating build configuration..."
cat > Makefile << 'EOF'
.PHONY: build run stop clean

build:
	@echo "Building ComputeHive components..."
	cd agent && go build -o ../bin/computehive-agent cmd/agent/main.go
	cd core-services/auth-service && go build -o ../../bin/auth-service main.go

run-agent:
	@echo "Starting ComputeHive Agent..."
	./bin/computehive-agent

run-auth:
	@echo "Starting Auth Service..."
	./bin/auth-service

run-services:
	@echo "Starting Docker services..."
	docker-compose up -d

stop-services:
	@echo "Stopping Docker services..."
	docker-compose down

clean:
	rm -rf bin/*
	docker-compose down -v
EOF

echo ""
echo "✅ ComputeHive basic implementation created!"
echo ""
echo "🎯 Quick Start Commands:"
echo "  1. Install Go dependencies:"
echo "     cd agent && go mod tidy && cd .."
echo "     cd core-services && go mod tidy && cd .."
echo ""
echo "  2. Build the components:"
echo "     make build"
echo ""
echo "  3. Start Docker services:"
echo "     make run-services"
echo ""
echo "  4. Run the agent:"
echo "     make run-agent"
echo ""
echo "  5. Run the auth service:"
echo "     make run-auth"
echo ""
echo "🚀 Ready to start development!" 