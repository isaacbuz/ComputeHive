package io.computehive.sdk.jobs;

import com.google.gson.Gson;
import io.computehive.sdk.ComputeHiveClient;
import io.computehive.sdk.exceptions.ComputeHiveException;
import io.computehive.sdk.models.Job;
import okhttp3.*;
import okhttp3.MediaType;

import java.io.IOException;
import java.util.List;
import java.util.concurrent.CompletableFuture;

/**
 * Service for managing compute jobs.
 */
public class JobService {
    
    private static final MediaType JSON = MediaType.get("application/json; charset=utf-8");
    
    private final ComputeHiveClient client;
    private final Gson gson;
    
    public JobService(ComputeHiveClient client) {
        this.client = client;
        this.gson = client.getGson();
    }
    
    /**
     * Submit a new job.
     * 
     * @param job The job to submit
     * @return CompletableFuture containing the created job
     */
    public CompletableFuture<Job> submitJob(Job job) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String json = gson.toJson(job);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/jobs")
                        .post(body)
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to submit job: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, Job.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Job submission failed", e);
            }
        });
    }
    
    /**
     * Get a job by ID.
     * 
     * @param jobId The job ID
     * @return CompletableFuture containing the job
     */
    public CompletableFuture<Job> getJob(String jobId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/jobs/" + jobId)
                        .get()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get job: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, Job.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to get job", e);
            }
        });
    }
    
    /**
     * List all jobs with optional filtering.
     * 
     * @param status Filter by job status
     * @param limit Maximum number of jobs to return
     * @param offset Offset for pagination
     * @return CompletableFuture containing the list of jobs
     */
    public CompletableFuture<List<Job>> listJobs(Job.Status status, Integer limit, Integer offset) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                HttpUrl.Builder urlBuilder = HttpUrl.parse(client.getApiUrl() + "/jobs").newBuilder();
                
                if (status != null) {
                    urlBuilder.addQueryParameter("status", status.name());
                }
                if (limit != null) {
                    urlBuilder.addQueryParameter("limit", limit.toString());
                }
                if (offset != null) {
                    urlBuilder.addQueryParameter("offset", offset.toString());
                }
                
                Request request = new Request.Builder()
                        .url(urlBuilder.build())
                        .get()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to list jobs: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    JobListResponse jobListResponse = gson.fromJson(responseBody, JobListResponse.class);
                    return jobListResponse.getJobs();
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to list jobs", e);
            }
        });
    }
    
    /**
     * Cancel a running job.
     * 
     * @param jobId The job ID to cancel
     * @return CompletableFuture containing the updated job
     */
    public CompletableFuture<Job> cancelJob(String jobId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/jobs/" + jobId + "/cancel")
                        .post(RequestBody.create("", null))
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to cancel job: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, Job.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to cancel job", e);
            }
        });
    }
    
    /**
     * Delete a completed job.
     * 
     * @param jobId The job ID to delete
     * @return CompletableFuture that completes when job is deleted
     */
    public CompletableFuture<Void> deleteJob(String jobId) {
        return CompletableFuture.runAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/jobs/" + jobId)
                        .delete()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to delete job: " + response.code());
                    }
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to delete job", e);
            }
        });
    }
    
    /**
     * Get job logs.
     * 
     * @param jobId The job ID
     * @param lines Number of lines to retrieve
     * @return CompletableFuture containing the job logs
     */
    public CompletableFuture<String> getJobLogs(String jobId, Integer lines) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                HttpUrl.Builder urlBuilder = HttpUrl.parse(client.getApiUrl() + "/jobs/" + jobId + "/logs").newBuilder();
                
                if (lines != null) {
                    urlBuilder.addQueryParameter("lines", lines.toString());
                }
                
                Request request = new Request.Builder()
                        .url(urlBuilder.build())
                        .get()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get job logs: " + response.code());
                    }
                    
                    return response.body().string();
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to get job logs", e);
            }
        });
    }
    
    /**
     * Get job metrics.
     * 
     * @param jobId The job ID
     * @return CompletableFuture containing the job metrics
     */
    public CompletableFuture<JobMetrics> getJobMetrics(String jobId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/jobs/" + jobId + "/metrics")
                        .get()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get job metrics: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, JobMetrics.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to get job metrics", e);
            }
        });
    }
    
    /**
     * Retry a failed job.
     * 
     * @param jobId The job ID to retry
     * @return CompletableFuture containing the new job
     */
    public CompletableFuture<Job> retryJob(String jobId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/jobs/" + jobId + "/retry")
                        .post(RequestBody.create("", null))
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to retry job: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, Job.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to retry job", e);
            }
        });
    }
    
    /**
     * Get job cost information.
     * 
     * @param jobId The job ID
     * @return CompletableFuture containing the job cost
     */
    public CompletableFuture<JobCost> getJobCost(String jobId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/jobs/" + jobId + "/cost")
                        .get()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get job cost: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, JobCost.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to get job cost", e);
            }
        });
    }
    
    // Helper classes
    private static class JobListResponse {
        private List<Job> jobs;
        private int total;
        private int limit;
        private int offset;
        
        public List<Job> getJobs() { return jobs; }
        public void setJobs(List<Job> jobs) { this.jobs = jobs; }
        public int getTotal() { return total; }
        public void setTotal(int total) { this.total = total; }
        public int getLimit() { return limit; }
        public void setLimit(int limit) { this.limit = limit; }
        public int getOffset() { return offset; }
        public void setOffset(int offset) { this.offset = offset; }
    }
    
    /**
     * Job metrics information.
     */
    public static class JobMetrics {
        private double cpuUsage;
        private double memoryUsage;
        private double gpuUsage;
        private double networkUsage;
        private double diskUsage;
        private long uptime;
        private int activeConnections;
        
        // Getters and setters
        public double getCpuUsage() { return cpuUsage; }
        public void setCpuUsage(double cpuUsage) { this.cpuUsage = cpuUsage; }
        public double getMemoryUsage() { return memoryUsage; }
        public void setMemoryUsage(double memoryUsage) { this.memoryUsage = memoryUsage; }
        public double getGpuUsage() { return gpuUsage; }
        public void setGpuUsage(double gpuUsage) { this.gpuUsage = gpuUsage; }
        public double getNetworkUsage() { return networkUsage; }
        public void setNetworkUsage(double networkUsage) { this.networkUsage = networkUsage; }
        public double getDiskUsage() { return diskUsage; }
        public void setDiskUsage(double diskUsage) { this.diskUsage = diskUsage; }
        public long getUptime() { return uptime; }
        public void setUptime(long uptime) { this.uptime = uptime; }
        public int getActiveConnections() { return activeConnections; }
        public void setActiveConnections(int activeConnections) { this.activeConnections = activeConnections; }
    }
    
    /**
     * Job cost information.
     */
    public static class JobCost {
        private double computeCost;
        private double storageCost;
        private double networkCost;
        private double totalCost;
        private String currency;
        private double hourlyRate;
        private long durationSeconds;
        
        // Getters and setters
        public double getComputeCost() { return computeCost; }
        public void setComputeCost(double computeCost) { this.computeCost = computeCost; }
        public double getStorageCost() { return storageCost; }
        public void setStorageCost(double storageCost) { this.storageCost = storageCost; }
        public double getNetworkCost() { return networkCost; }
        public void setNetworkCost(double networkCost) { this.networkCost = networkCost; }
        public double getTotalCost() { return totalCost; }
        public void setTotalCost(double totalCost) { this.totalCost = totalCost; }
        public String getCurrency() { return currency; }
        public void setCurrency(String currency) { this.currency = currency; }
        public double getHourlyRate() { return hourlyRate; }
        public void setHourlyRate(double hourlyRate) { this.hourlyRate = hourlyRate; }
        public long getDurationSeconds() { return durationSeconds; }
        public void setDurationSeconds(long durationSeconds) { this.durationSeconds = durationSeconds; }
    }
} 