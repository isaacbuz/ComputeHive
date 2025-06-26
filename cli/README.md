# ComputeHive CLI

The ComputeHive Command Line Interface (CLI) is a powerful tool for interacting with the ComputeHive distributed compute platform. It allows you to manage agents, submit jobs, interact with the marketplace, and monitor your usage from the terminal.

## Features

- **Agent Management**: Deploy and manage compute agents on any machine
- **Job Submission**: Submit compute jobs with Docker or scripts
- **Marketplace**: Browse offers, place bids, and monitor prices
- **Billing**: Track usage, manage payments, and set spending alerts
- **Authentication**: Secure login with multiple authentication methods
- **Cross-Platform**: Works on Windows, macOS, and Linux

## Installation

### Pre-built Binaries

Download the latest release for your platform:

```bash
# macOS
curl -L https://github.com/computehive/cli/releases/latest/download/computehive-darwin-amd64 -o computehive
chmod +x computehive
sudo mv computehive /usr/local/bin/

# Linux
curl -L https://github.com/computehive/cli/releases/latest/download/computehive-linux-amd64 -o computehive
chmod +x computehive
sudo mv computehive /usr/local/bin/

# Windows
# Download computehive-windows-amd64.exe from releases and add to PATH
```

### Build from Source

```bash
git clone https://github.com/computehive/cli.git
cd cli
go build -o computehive main.go
```

## Quick Start

### 1. Authentication

```bash
# Login with email/password
computehive auth login

# Login with token
computehive auth login --token YOUR_API_TOKEN

# Login with OAuth
computehive auth login --provider github
```

### 2. Start an Agent

```bash
# Start agent with default settings
computehive agent start

# Start agent with custom name and tags
computehive agent start --name "gpu-server-1" --tag gpu --tag ml

# View agent status
computehive agent list
```

### 3. Submit a Job

```bash
# Submit a Docker job
computehive job submit --docker ubuntu:latest --command "echo 'Hello, World!'"

# Submit a Python script
computehive job submit --script train.py --cpu 8 --memory 32 --gpu 2

# Submit and wait for completion
computehive job submit --docker myapp:latest --wait --output ./results
```

### 4. Monitor Jobs

```bash
# List all jobs
computehive job list

# Get job details
computehive job get JOB_ID

# View job logs
computehive job logs JOB_ID --follow

# Download results
computehive job results JOB_ID --output ./results
```

## Command Reference

### Authentication Commands

```bash
computehive auth login              # Interactive login
computehive auth logout             # Logout
computehive auth status             # Show authentication status
computehive auth token --create     # Create API token
computehive auth token --list       # List API tokens
```

### Agent Commands

```bash
computehive agent start             # Start an agent
computehive agent stop AGENT_ID     # Stop an agent
computehive agent list              # List all agents
computehive agent status AGENT_ID   # Get agent details
computehive agent logs AGENT_ID     # View agent logs
computehive agent config AGENT_ID   # Get agent configuration
```

### Job Commands

```bash
computehive job submit              # Submit a new job
computehive job list                # List jobs
computehive job get JOB_ID          # Get job details
computehive job logs JOB_ID         # View job logs
computehive job cancel JOB_ID       # Cancel a job
computehive job status JOB_ID       # Check job status
computehive job results JOB_ID      # Download job results
```

### Marketplace Commands

```bash
computehive marketplace offers      # List available offers
computehive marketplace bids        # List active bids
computehive marketplace create-offer # Create a resource offer
computehive marketplace create-bid  # Create a resource bid
computehive marketplace prices      # Show current market prices
```

### Billing Commands

```bash
computehive billing usage           # Show resource usage
computehive billing invoices        # List invoices
computehive billing payment-methods # Manage payment methods
computehive billing history         # Show payment history
computehive billing add-funds       # Add funds to account
computehive billing alerts          # Manage billing alerts
```

### Status Commands

```bash
computehive status                  # Show system status
computehive status services         # Show service health
computehive status agents           # Show agent status
computehive status jobs             # Show job statistics
computehive status account          # Show account status
```

### Configuration Commands

```bash
computehive config show             # Show current configuration
computehive config set KEY VALUE    # Set configuration value
computehive config get KEY          # Get configuration value
computehive config list             # List all configuration keys
computehive config reset            # Reset to defaults
```

## Examples

### Submit a Machine Learning Job

```bash
# Submit PyTorch training job with GPU
computehive job submit \
  --docker pytorch/pytorch:latest \
  --script train.py \
  --cpu 16 \
  --memory 64 \
  --gpu 4 \
  --gpu-model "nvidia-a100" \
  --env EPOCHS=100 \
  --env BATCH_SIZE=256 \
  --volume ./data:/data \
  --wait \
  --output ./results
```

### Create a GPU Offer

```bash
# Offer 4 NVIDIA A100 GPUs
computehive marketplace create-offer \
  --cpu 32 \
  --memory 256 \
  --gpu 4 \
  --gpu-model "nvidia-a100" \
  --storage 1000 \
  --price 25.00 \
  --location "us-west-2" \
  --duration 168h \
  --auto-accept
```

### Monitor Spending

```bash
# Set up daily spending alert
computehive billing alerts --add \
  --type daily \
  --threshold 100

# View usage for the last week
computehive billing usage \
  --period 7d \
  --details
```

### Batch Job Submission

```bash
# Submit multiple jobs from a script
for i in {1..10}; do
  computehive job submit \
    --docker myapp:latest \
    --env JOB_ID=$i \
    --name "batch-job-$i"
done

# Monitor all jobs
computehive job list --watch
```

## Configuration

The CLI stores configuration in `~/.computehive/config.json`:

```json
{
  "api_url": "https://api.computehive.io",
  "token": "YOUR_TOKEN",
  "default_region": "us-east-1",
  "output_format": "table",
  "color_output": true
}
```

### Environment Variables

- `COMPUTEHIVE_TOKEN` - Authentication token
- `COMPUTEHIVE_API_URL` - API endpoint URL
- `COMPUTEHIVE_CONFIG` - Path to config file

### Output Formats

```bash
# JSON output
computehive job list -o json

# YAML output
computehive job get JOB_ID -o yaml

# Disable color output
computehive config set color false
```

## Shell Completion

Enable shell completion for better productivity:

```bash
# Bash
computehive completion bash > /etc/bash_completion.d/computehive

# Zsh
computehive completion zsh > "${fpath[1]}/_computehive"

# Fish
computehive completion fish > ~/.config/fish/completions/computehive.fish

# PowerShell
computehive completion powershell | Out-String | Invoke-Expression
```

## Troubleshooting

### Debug Mode

Enable debug output for troubleshooting:

```bash
computehive --debug job submit ...
```

### Common Issues

1. **Authentication Error**
   ```bash
   computehive auth status  # Check auth status
   computehive auth login   # Re-authenticate
   ```

2. **Connection Issues**
   ```bash
   # Use a different API endpoint
   computehive config set api-url https://api.computehive.io
   
   # Set proxy if needed
   computehive config set proxy-url http://proxy.example.com:8080
   ```

3. **Job Submission Failures**
   ```bash
   # Validate job specification
   computehive job submit --dry-run ...
   
   # Check resource availability
   computehive marketplace offers --cpu 8 --memory 32
   ```

## Support

- Documentation: https://docs.computehive.io/cli
- Issues: https://github.com/computehive/cli/issues
- Discord: https://discord.gg/computehive
- Email: support@computehive.io

## License

MIT License - see LICENSE file for details 