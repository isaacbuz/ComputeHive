package io.computehive.sdk.examples;

import io.computehive.sdk.ComputeHiveClient;
import io.computehive.sdk.models.Job;
import io.computehive.sdk.models.Job.ResourceRequirements;
import io.computehive.sdk.models.Job.Status;
import io.computehive.sdk.jobs.JobService;

import java.time.Duration;
import java.util.Arrays;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * Basic example demonstrating job submission and monitoring.
 */
public class BasicJobExample {
    
    public static void main(String[] args) {
        // Create client with API key
        ComputeHiveClient client = ComputeHiveClient.builder()
                .apiKey("your-api-key-here")
                .timeout(Duration.ofMinutes(5))
                .build();
        
        try {
            // Define resource requirements
            ResourceRequirements resources = ResourceRequirements.builder()
                    .cpuCores(2)
                    .memoryMB(4096)
                    .diskGB(50)
                    .build();
            
            // Create a simple job
            Job job = Job.builder()
                    .name("Hello ComputeHive")
                    .description("A simple example job")
                    .image("ubuntu:20.04")
                    .command(Arrays.asList("echo", "Hello from ComputeHive!"))
                    .resources(resources)
                    .environment(Map.of(
                            "GREETING", "Hello World",
                            "PLATFORM", "ComputeHive"
                    ))
                    .timeout(3600L) // 1 hour timeout
                    .maxRetries(3)
                    .build();
            
            System.out.println("Submitting job...");
            
            // Submit the job
            CompletableFuture<Job> submitFuture = client.jobs().submitJob(job);
            Job submittedJob = submitFuture.join();
            
            System.out.println("Job submitted successfully!");
            System.out.println("Job ID: " + submittedJob.getId());
            System.out.println("Status: " + submittedJob.getStatus());
            
            // Monitor job status
            monitorJob(client, submittedJob.getId());
            
            // Get job logs
            System.out.println("\n=== Job Logs ===");
            String logs = client.jobs().getJobLogs(submittedJob.getId(), 100).join();
            System.out.println(logs);
            
            // Get job cost
            System.out.println("\n=== Job Cost ===");
            JobService.JobCost cost = client.jobs().getJobCost(submittedJob.getId()).join();
            System.out.println("Total Cost: $" + cost.getTotalCost());
            System.out.println("Currency: " + cost.getCurrency());
            System.out.println("Duration: " + cost.getDurationSeconds() + " seconds");
            
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
            e.printStackTrace();
        } finally {
            // Clean up
            client.close();
        }
    }
    
    /**
     * Monitor job status until completion.
     */
    private static void monitorJob(ComputeHiveClient client, String jobId) {
        System.out.println("\n=== Monitoring Job ===");
        
        while (true) {
            try {
                Job job = client.jobs().getJob(jobId).join();
                Status status = job.getStatus();
                
                System.out.println("Status: " + status + " | " + 
                        (job.getStartedAt() != null ? "Started: " + job.getStartedAt() : ""));
                
                switch (status) {
                    case COMPLETED:
                        System.out.println("✅ Job completed successfully!");
                        System.out.println("Exit Code: " + job.getExitCode());
                        return;
                        
                    case FAILED:
                        System.err.println("❌ Job failed!");
                        System.err.println("Error: " + job.getErrorMessage());
                        return;
                        
                    case CANCELLED:
                        System.out.println("⚠️ Job was cancelled");
                        return;
                        
                    case TIMEOUT:
                        System.err.println("⏰ Job timed out");
                        return;
                        
                    case PENDING:
                    case QUEUED:
                    case RUNNING:
                    case RETRYING:
                        // Continue monitoring
                        break;
                }
                
                // Wait before next check
                Thread.sleep(5000);
                
            } catch (InterruptedException e) {
                Thread.currentThread().interrupt();
                System.err.println("Monitoring interrupted");
                break;
            } catch (Exception e) {
                System.err.println("Error monitoring job: " + e.getMessage());
                break;
            }
        }
    }
} 