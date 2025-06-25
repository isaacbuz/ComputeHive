# ComputeHive

[![CI/CD](https://github.com/isaacbuz/ComputeHive/actions/workflows/main.yml/badge.svg)](https://github.com/isaacbuz/ComputeHive/actions/workflows/main.yml)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Go Version](https://img.shields.io/badge/go-1.21+-blue.svg)](https://golang.org/)
[![Docker](https://img.shields.io/badge/docker-ready-brightgreen.svg)](https://www.docker.com/)

ComputeHive is a next-generation, distributed compute platform that creates an autonomous, agent-driven marketplace for heterogeneous hardware resources. It enables seamless sharing of computational power across CPUs, GPUs, FPGAs, and TPUs in a secure, efficient, and economically incentivized environment.

## 🌟 Key Features

- **Multi-Platform Agent**: Cross-platform support for Windows, macOS, Linux, Android, and iOS
- **Heterogeneous Hardware**: Support for CPUs, GPUs (NVIDIA, AMD, Intel), FPGAs, and TPUs
- **Zero-Trust Security**: Hardware attestation, mTLS communication, and encrypted execution
- **Blockchain Integration**: Smart contract-based payments and dispute resolution
- **AI-Powered Operations**: Intelligent job scheduling and resource optimization
- **Enterprise Ready**: SLA guarantees, compliance tools, and audit trails

## 🏗️ Architecture

```
┌─────────────────────────────────────────────────────────────────┐
│                          ComputeHive Platform                     │
├─────────────────────────────────────────────────────────────────┤
│                                                                   │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐             │
│  │   Agents    │  │  Dashboard  │  │   Mobile    │             │
│  │  (Compute   │  │    (Web)    │  │   Apps      │             │
│  │  Providers) │  │             │  │             │             │
│  └──────┬──────┘  └──────┬──────┘  └──────┬──────┘             │
│         │                 │                 │                     │
│  ┌──────┴─────────────────┴─────────────────┴──────┐            │
│  │              API Gateway & Load Balancer         │            │
│  └──────────────────────┬───────────────────────────┘            │
│                         │                                         │
│  ┌──────────────────────┴───────────────────────────┐            │
│  │                Core Services                      │            │
│  │  ┌────────┐ ┌────────┐ ┌────────┐ ┌────────┐   │            │
│  │  │  Auth  │ │Schedule│ │Market  │ │Payment │   │            │
│  │  │Service │ │Service │ │Service │ │Service │   │            │
│  │  └────────┘ └────────┘ └────────┘ └────────┘   │            │
│  └───────────────────────────────────────────────────┘            │
│                                                                   │
│  ┌─────────────────┐  ┌─────────────────┐  ┌─────────────────┐  │
│  │   Blockchain    │  │    Database     │  │   Message Queue │  │
│  │   (Ethereum)    │  │  (CockroachDB)  │  │     (NATS)      │  │
│  └─────────────────┘  └─────────────────┘  └─────────────────┘  │
│                                                                   │
└─────────────────────────────────────────────────────────────────┘
```

## 🚀 Getting Started

### Prerequisites

- Go 1.21 or higher
- Docker and Docker Compose
- Node.js 20+ and npm
- Git

### Quick Start

1. **Clone the repository**
   ```bash
   git clone https://github.com/isaacbuz/ComputeHive.git
   cd ComputeHive
   ```

2. **Run the setup script**
   ```bash
   ./scripts/setup-dev.sh
   ```

3. **Start the development environment**
   ```bash
   docker-compose up -d
   ```

4. **Access the dashboard**
   ```
   http://localhost:3000
   ```

### Running the Agent

#### From Source
```bash
cd agent
go build -o computehive-agent ./cmd/agent
./computehive-agent --control-plane http://localhost:8000
```

#### Using Docker
```bash
docker run -d \
  --name computehive-agent \
  -e COMPUTEHIVE_CONTROL_PLANE_URL=https://api.computehive.io \
  -e COMPUTEHIVE_TOKEN=your-token \
  -v /var/run/docker.sock:/var/run/docker.sock \
  computehive/agent:latest
```

#### Using Pre-built Binaries
Download from [Releases](https://github.com/isaacbuz/ComputeHive/releases):
```bash
# Linux/macOS
curl -L https://github.com/isaacbuz/ComputeHive/releases/latest/download/computehive-agent-$(uname -s)-$(uname -m) -o computehive-agent
chmod +x computehive-agent
./computehive-agent --help

# Windows
# Download computehive-agent-windows-amd64.exe from releases page
```

## 📖 Documentation

- [Architecture Overview](docs/architecture.md)
- [API Documentation](docs/api.md)
- [Agent Setup Guide](docs/agent-setup.md)
- [Smart Contract Documentation](docs/contracts.md)
- [Security Model](docs/security.md)

## 🛠️ Development

### Project Structure
```
ComputeHive/
├── agent/              # Distributed compute agent (Go)
├── core-services/      # Microservices (Go)
├── web/               # Web applications
│   ├── dashboard/     # React dashboard
│   └── api-gateway/   # API gateway
├── contracts/         # Smart contracts (Solidity)
├── sdk/              # Client SDKs
├── infrastructure/    # Deployment configs
└── tests/            # Test suites
```

### Building from Source

**Agent:**
```bash
cd agent
go mod download
go build -o bin/computehive-agent ./cmd/agent
```

**Core Services:**
```bash
cd core-services
go mod download
go build -o bin/auth-service ./auth-service
```

**Dashboard:**
```bash
cd web/dashboard
npm install
npm run build
```

### Running Tests

```bash
# Run all tests
make test

# Run specific component tests
cd agent && go test ./...
cd web/dashboard && npm test
cd contracts && npm test
```

## 🤝 Contributing

We welcome contributions! Please see our [Contributing Guidelines](CONTRIBUTING.md) for details.

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Setup

```bash
# Install development dependencies
make dev-setup

# Run linters
make lint

# Run tests with coverage
make test-coverage
```

## 📊 Performance

- **Agent Overhead**: < 2% CPU, < 100MB RAM
- **Job Scheduling**: < 100ms latency
- **Network**: Optimized for high-throughput data transfer
- **Blockchain**: Layer 2 scaling for high transaction volume

## 🔒 Security

- **Zero-Trust Architecture**: All communications are authenticated and encrypted
- **Hardware Attestation**: SGX/SEV support for trusted execution
- **Secure Enclaves**: Sensitive computations in isolated environments
- **Regular Audits**: Automated and manual security testing

See our [Security Policy](SECURITY.md) for reporting vulnerabilities.

## 📄 License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- The open-source community for amazing tools and libraries
- Early adopters and beta testers
- Contributors and maintainers

## 📞 Contact

- **Website**: [computehive.io](https://computehive.io)
- **Email**: support@computehive.io
- **Discord**: [Join our community](https://discord.gg/computehive)
- **Twitter**: [@ComputeHive](https://twitter.com/ComputeHive)

## 🗺️ Roadmap

See our [public roadmap](https://github.com/isaacbuz/ComputeHive/projects/1) for upcoming features and milestones.

---

<p align="center">Built with ❤️ by the ComputeHive Team</p> 