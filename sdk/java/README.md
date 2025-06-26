# ComputeHive Java SDK

Official Java SDK for the ComputeHive distributed compute platform. This SDK provides a comprehensive interface for interacting with ComputeHive's cloud computing services, including job management, marketplace operations, payments, and real-time monitoring.

## Features

- **Job Management**: Submit, monitor, and manage compute jobs
- **Marketplace Integration**: Browse and reserve compute resources
- **Payment Processing**: Handle billing and payment operations
- **Real-time Monitoring**: WebSocket-based event streaming
- **Authentication**: Secure API authentication with JWT tokens
- **Telemetry**: Send metrics and logs for monitoring
- **Async Operations**: Non-blocking API calls with CompletableFuture
- **Type Safety**: Full type safety with comprehensive model classes

## Requirements

- Java 11 or higher
- Maven 3.6+ or Gradle 6.0+

## Installation

### Maven

Add the following dependency to your `pom.xml`:

```xml
<dependency>
    <groupId>io.computehive</groupId>
    <artifactId>computehive-sdk</artifactId>
    <version>1.0.0</version>
</dependency>
```

### Gradle

Add the following dependency to your `build.gradle`:

```gradle
implementation 'io.computehive:computehive-sdk:1.0.0'
```

## Quick Start

### Basic Usage

```java
import io.computehive.sdk.ComputeHiveClient;
import io.computehive.sdk.models.Job;
import io.computehive.sdk.models.auth.AuthResponse;

// Create a client with API key
ComputeHiveClient client = ComputeHiveClient.builder()
    .apiKey("your-api-key")
    .build();

// Or authenticate with email/password
ComputeHiveClient client = ComputeHiveClient.builder()
    .apiUrl("https://api.computehive.io")
    .build();

// Authenticate
AuthResponse authResponse = client.authenticate("user@example.com", "password").join();

// Submit a job
Job job = Job.builder()
    .name("My Compute Job")
    .description("A sample compute job")
    .image("ubuntu:20.04")
    .command(Arrays.asList("echo", "Hello, ComputeHive!"))
    .build();

Job submittedJob = client.jobs().submitJob(job).join();
System.out.println("Job submitted with ID: " + submittedJob.getId());

// Monitor job status
Job.Status status = submittedJob.getStatus();
System.out.println("Job status: " + status);
```

### Real-time Event Handling

```java
// Connect to WebSocket for real-time updates
client.connect();

// Subscribe to job events
client.events().on("job.started", (eventType, data) -> {
    System.out.println("Job started: " + data);
});

client.events().on("job.completed", (eventType, data) -> {
    System.out.println("Job completed: " + data);
});

client.events().on("job.failed", (eventType, data) -> {
    System.out.println("Job failed: " + data);
});
```

### Job Management

```java
// List all jobs
List<Job> jobs = client.jobs().listJobs(null, 10, 0).join();

// Get specific job
Job job = client.jobs().getJob("job-id").join();

// Cancel a job
client.jobs().cancelJob("job-id").join();

// Get job logs
String logs = client.jobs().getJobLogs("job-id", 100).join();

// Get job metrics
JobService.JobMetrics metrics = client.jobs().getJobMetrics("job-id").join();
```

### Marketplace Operations

```java
// List available resources
List<MarketplaceService.ComputeResource> resources = 
    client.marketplace().listResources("us-west-1", null, null, true).join();

// Reserve a resource
ResourceReservation reservation = 
    client.marketplace().reserveResource("resource-id", 24).join();

// Get pricing information
List<MarketplaceService.PricingInfo> pricing = 
    client.marketplace().getPricing("us-west-1", "gpu-instance").join();
```

### Payment Operations

```java
// Get account balance
AccountBalance balance = client.payments().getBalance().join();

// Add funds
PaymentTransaction transaction = 
    client.payments().addFunds(100.0, "USD", "credit-card").join();

// Get payment history
List<PaymentTransaction> history = 
    client.payments().getPaymentHistory(10, 0).join();

// Get billing information
BillingInfo billing = client.payments().getBillingInfo().join();
```

### Telemetry and Monitoring

```java
// Send custom metrics
MetricsData metrics = new MetricsData();
metrics.setJobId("job-id");
metrics.setTimestamp(System.currentTimeMillis());
metrics.setMetrics(Map.of("cpu_usage", 75.5, "memory_usage", 60.2));
client.telemetry().sendMetrics(metrics).join();

// Send logs
LogData logData = new LogData();
logData.setJobId("job-id");
logData.setLevel("INFO");
logData.setMessage("Job processing completed");
client.telemetry().sendLogs(logData).join();

// Get system metrics
SystemMetrics systemMetrics = 
    client.telemetry().getSystemMetrics(null, TimeRange.LAST_HOUR).join();
```

## Configuration

### Client Configuration

```java
ComputeHiveClient client = ComputeHiveClient.builder()
    .apiUrl("https://api.computehive.io")
    .wsUrl("wss://api.computehive.io/ws")
    .apiKey("your-api-key")
    .timeout(Duration.ofMinutes(5))
    .debug(true)
    .build();
```

### Custom HTTP Client

```java
OkHttpClient customHttpClient = new OkHttpClient.Builder()
    .connectTimeout(60, TimeUnit.SECONDS)
    .readTimeout(60, TimeUnit.SECONDS)
    .addInterceptor(new CustomInterceptor())
    .build();

ComputeHiveClient client = ComputeHiveClient.builder()
    .customHttpClient(customHttpClient)
    .build();
```

## Error Handling

The SDK uses `ComputeHiveException` for error handling:

```java
try {
    Job job = client.jobs().getJob("invalid-job-id").join();
} catch (CompletionException e) {
    if (e.getCause() instanceof ComputeHiveException) {
        ComputeHiveException che = (ComputeHiveException) e.getCause();
        System.err.println("Error: " + che.getMessage());
        System.err.println("Status Code: " + che.getStatusCode());
        System.err.println("Error Code: " + che.getErrorCode());
    }
}
```

## Authentication

### API Key Authentication

```java
ComputeHiveClient client = ComputeHiveClient.builder()
    .apiKey("your-api-key")
    .build();
```

### Email/Password Authentication

```java
ComputeHiveClient client = ComputeHiveClient.builder()
    .apiUrl("https://api.computehive.io")
    .build();

AuthResponse authResponse = client.authenticate("user@example.com", "password").join();
String accessToken = authResponse.getAccessToken();
```

### Token Refresh

```java
// The SDK automatically handles token refresh
// You can also manually refresh tokens
AuthResponse newAuthResponse = client.auth().refreshToken(refreshToken).join();
```

## Examples

### Complete Job Workflow

```java
import io.computehive.sdk.ComputeHiveClient;
import io.computehive.sdk.models.Job;
import io.computehive.sdk.models.Job.ResourceRequirements;

public class JobWorkflowExample {
    public static void main(String[] args) {
        // Create client
        ComputeHiveClient client = ComputeHiveClient.builder()
            .apiKey("your-api-key")
            .build();

        // Define resource requirements
        ResourceRequirements resources = ResourceRequirements.builder()
            .cpuCores(4)
            .memoryMB(8192)
            .gpuCount(1)
            .gpuType("RTX-4090")
            .diskGB(100)
            .build();

        // Create job
        Job job = Job.builder()
            .name("Machine Learning Training")
            .description("Training a deep learning model")
            .image("pytorch/pytorch:latest")
            .command(Arrays.asList("python", "train.py"))
            .resources(resources)
            .environment(Map.of(
                "DATASET_PATH", "/data/dataset",
                "MODEL_PATH", "/output/model"
            ))
            .build();

        // Submit job
        Job submittedJob = client.jobs().submitJob(job).join();
        System.out.println("Job submitted: " + submittedJob.getId());

        // Monitor job
        while (true) {
            Job currentJob = client.jobs().getJob(submittedJob.getId()).join();
            System.out.println("Job status: " + currentJob.getStatus());

            if (currentJob.getStatus() == Job.Status.COMPLETED) {
                System.out.println("Job completed successfully!");
                break;
            } else if (currentJob.getStatus() == Job.Status.FAILED) {
                System.err.println("Job failed: " + currentJob.getErrorMessage());
                break;
            }

            Thread.sleep(5000); // Wait 5 seconds
        }

        // Get results
        String logs = client.jobs().getJobLogs(submittedJob.getId(), 1000).join();
        System.out.println("Job logs:\n" + logs);
    }
}
```

### Real-time Monitoring

```java
public class RealTimeMonitoringExample {
    public static void main(String[] args) {
        ComputeHiveClient client = ComputeHiveClient.builder()
            .apiKey("your-api-key")
            .build();

        // Connect to WebSocket
        client.connect();

        // Subscribe to events
        client.events().on("job.started", (eventType, data) -> {
            System.out.println("Job started: " + data);
        });

        client.events().on("job.completed", (eventType, data) -> {
            System.out.println("Job completed: " + data);
        });

        client.events().on("job.failed", (eventType, data) -> {
            System.err.println("Job failed: " + data);
        });

        client.events().on("resource.available", (eventType, data) -> {
            System.out.println("Resource available: " + data);
        });

        // Keep the application running
        try {
            Thread.sleep(Long.MAX_VALUE);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
        }
    }
}
```

## API Reference

### Core Classes

- `ComputeHiveClient`: Main client class
- `Job`: Job model and operations
- `AuthResponse`: Authentication response
- `Credentials`: User credentials
- `UserProfile`: User profile information

### Services

- `JobService`: Job management operations
- `MarketplaceService`: Marketplace operations
- `PaymentService`: Payment and billing operations
- `TelemetryService`: Metrics and monitoring
- `AuthService`: Authentication operations
- `WebSocketClient`: Real-time event handling

### Models

- `Job`: Complete job model with all properties
- `ComputeResource`: Marketplace resource information
- `PaymentTransaction`: Payment transaction details
- `SystemMetrics`: System performance metrics
- `Alert`: System alerts and notifications

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Ensure all tests pass
6. Submit a pull request

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

## Support

- Documentation: [https://docs.computehive.io](https://docs.computehive.io)
- Issues: [https://github.com/computehive/computehive-sdk-java/issues](https://github.com/computehive/computehive-sdk-java/issues)
- Email: support@computehive.io 