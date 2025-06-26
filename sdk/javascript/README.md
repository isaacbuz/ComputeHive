# ComputeHive JavaScript/TypeScript SDK

Official JavaScript/TypeScript SDK for ComputeHive - the next-generation distributed compute platform.

## Installation

```bash
npm install @computehive/sdk
# or
yarn add @computehive/sdk
```

## Quick Start

```typescript
import { ComputeHiveClient } from '@computehive/sdk';

// Initialize the client
const client = new ComputeHiveClient({
  apiKey: 'your-api-key',
  // or use email/password authentication
});

// Authenticate (if using email/password)
await client.authenticate('user@example.com', 'password');

// Create a job
const job = await client.createJob({
  name: 'My Training Job',
  type: JobType.BATCH,
  docker_image: 'tensorflow/tensorflow:latest-gpu',
  resource_requirements: {
    cpu: { cores: 4 },
    memory: { size: 16 }, // GB
    gpu: { count: 1, model: 'nvidia-t4' }
  }
});

// Wait for completion
const result = await client.waitForJob(job.id);
console.log('Job completed:', result);
```

## Features

- 🚀 **Easy to use** - Simple, intuitive API
- 📦 **TypeScript support** - Full type definitions included
- 🔄 **Real-time updates** - WebSocket support for live job updates
- 🔐 **Secure** - Built-in authentication and secure communication
- 📊 **Comprehensive** - Access to all ComputeHive features

## Documentation

### Authentication

```typescript
// API Key authentication (recommended)
const client = new ComputeHiveClient({
  apiKey: 'your-api-key'
});

// Email/password authentication
const client = new ComputeHiveClient();
const auth = await client.authenticate('user@example.com', 'password');
```

### Job Management

```typescript
// Create a job
const job = await client.createJob({
  name: 'Data Processing Job',
  type: JobType.BATCH,
  docker_image: 'python:3.9',
  command: ['python', 'process.py'],
  environment: {
    INPUT_PATH: '/data/input',
    OUTPUT_PATH: '/data/output'
  },
  resource_requirements: {
    cpu: { cores: 8 },
    memory: { size: 32 }
  }
});

// List jobs
const jobs = await client.listJobs({
  status: JobStatus.RUNNING,
  limit: 10
});

// Get job details
const jobDetails = await client.getJob(job.id);

// Cancel a job
await client.cancelJob(job.id);

// Get job logs
const logs = await client.getJobLogs(job.id);
```

### Real-time Updates

```typescript
// Connect to WebSocket for real-time updates
client.connect();

// Listen for job updates
client.on('job.updated', (job) => {
  console.log('Job updated:', job);
});

client.on('job.completed', (job) => {
  console.log('Job completed:', job);
});

client.on('job.failed', (job, error) => {
  console.error('Job failed:', job.id, error);
});

// Disconnect when done
client.disconnect();
```

### Marketplace

```typescript
import { MarketplaceClient } from '@computehive/sdk';

const marketplace = new MarketplaceClient(client['http']);

// Get available offers
const offers = await marketplace.getOffers();

// Create a bid for a job
const bid = await marketplace.createBid({
  job_id: 'job-123',
  price: 10.50,
  agent_id: 'agent-456'
});
```

### Payments

```typescript
import { PaymentClient } from '@computehive/sdk';

const payments = new PaymentClient(client['http']);

// Check balance
const balance = await payments.getBalance();

// Get payment history
const history = await payments.getPaymentHistory();

// Make a deposit
const payment = await payments.deposit(100, 'USD');
```

### Telemetry

```typescript
import { TelemetryClient } from '@computehive/sdk';

const telemetry = new TelemetryClient(client['http']);

// Send custom metrics
await telemetry.sendMetrics([
  {
    name: 'model.accuracy',
    value: 0.95,
    timestamp: new Date().toISOString()
  }
]);

// Query metrics
const metrics = await telemetry.queryMetrics({
  metric: 'gpu.utilization',
  start: '2024-01-01T00:00:00Z',
  agent_id: 'agent-123'
});

// Create alerts
const alert = await telemetry.createAlert({
  name: 'High GPU Temperature',
  metric_name: 'gpu.temperature',
  threshold: 85
});
```

## Error Handling

```typescript
import { ComputeHiveError, AuthenticationError, APIError } from '@computehive/sdk';

try {
  const job = await client.createJob({...});
} catch (error) {
  if (error instanceof AuthenticationError) {
    // Handle authentication errors
    console.error('Authentication failed:', error.message);
  } else if (error instanceof APIError) {
    // Handle API errors
    console.error('API error:', error.statusCode, error.message);
  } else if (error instanceof ComputeHiveError) {
    // Handle other ComputeHive errors
    console.error('ComputeHive error:', error.code, error.message);
  } else {
    // Handle unexpected errors
    console.error('Unexpected error:', error);
  }
}
```

## Advanced Usage

### Custom Configuration

```typescript
const client = new ComputeHiveClient({
  apiUrl: 'https://custom-api.computehive.io',
  wsUrl: 'wss://custom-api.computehive.io/ws',
  timeout: 60000, // 60 seconds
  maxRetries: 5,
  debug: true // Enable debug logging
});
```

### Batch Operations

```typescript
// Create multiple jobs
const jobs = await Promise.all([
  client.createJob({ name: 'Job 1', ... }),
  client.createJob({ name: 'Job 2', ... }),
  client.createJob({ name: 'Job 3', ... })
]);

// Wait for all jobs to complete
const results = await Promise.all(
  jobs.map(job => client.waitForJob(job.id))
);
```

### Streaming Logs

```typescript
// Stream logs in real-time (requires WebSocket connection)
client.on('job.logs', (jobId, logs) => {
  console.log(`[${jobId}] ${logs}`);
});
```

## API Reference

### ComputeHiveClient

- `constructor(config?: ComputeHiveConfig)`
- `authenticate(email: string, password: string): Promise<AuthResponse>`
- `createJob(request: CreateJobRequest): Promise<Job>`
- `getJob(jobId: string): Promise<Job>`
- `listJobs(params?: ListJobsParams): Promise<Job[]>`
- `cancelJob(jobId: string): Promise<void>`
- `getJobLogs(jobId: string): Promise<string>`
- `waitForJob(jobId: string, timeout?: number): Promise<Job>`
- `connect(): void`
- `disconnect(): void`
- `on(event: string, handler: Function): void`
- `off(event: string, handler: Function): void`

### Types

See [src/types.ts](src/types.ts) for complete type definitions.

## Examples

Check out the [examples](examples/) directory for more detailed examples:

- [Basic job submission](examples/basic-job.ts)
- [GPU training job](examples/gpu-training.ts)
- [Real-time monitoring](examples/realtime-monitoring.ts)
- [Marketplace bidding](examples/marketplace.ts)

## Contributing

We welcome contributions! Please see our [Contributing Guide](CONTRIBUTING.md) for details.

## License

MIT License - see [LICENSE](LICENSE) for details.

## Support

- 📧 Email: support@computehive.io
- 💬 Discord: [Join our community](https://discord.gg/computehive)
- 📚 Documentation: [https://docs.computehive.io](https://docs.computehive.io)
- 🐛 Issues: [GitHub Issues](https://github.com/computehive/sdk-js/issues) 