package io.computehive.sdk.telemetry;

import com.google.gson.Gson;
import com.google.gson.reflect.TypeToken;
import io.computehive.sdk.ComputeHiveClient;
import io.computehive.sdk.exceptions.ComputeHiveException;
import okhttp3.*;

import java.io.IOException;
import java.lang.reflect.Type;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

/**
 * Service for telemetry and monitoring operations.
 */
public class TelemetryService {
    
    private static final MediaType JSON = MediaType.get("application/json; charset=utf-8");
    
    private final ComputeHiveClient client;
    private final Gson gson;
    private final OkHttpClient httpClient;
    
    public TelemetryService(ComputeHiveClient client) {
        this.client = client;
        this.gson = new Gson();
        this.httpClient = client.getHttpClient();
    }
    
    /**
     * Get system metrics.
     */
    public CompletableFuture<SystemMetrics> getSystemMetrics() {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/api/v1/telemetry/system")
                        .get()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get system metrics: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, SystemMetrics.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Error getting system metrics", e);
            }
        });
    }
    
    /**
     * Get job metrics.
     */
    public CompletableFuture<List<JobMetrics>> getJobMetrics(String jobId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/api/v1/telemetry/jobs/" + jobId + "/metrics")
                        .get()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get job metrics: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    Type listType = new TypeToken<List<JobMetrics>>(){}.getType();
                    return gson.fromJson(responseBody, listType);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Error getting job metrics", e);
            }
        });
    }
    
    /**
     * Get resource utilization.
     */
    public CompletableFuture<ResourceUtilization> getResourceUtilization(String resourceId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/api/v1/telemetry/resources/" + resourceId + "/utilization")
                        .get()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get resource utilization: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, ResourceUtilization.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Error getting resource utilization", e);
            }
        });
    }
    
    /**
     * Get performance analytics.
     */
    public CompletableFuture<PerformanceAnalytics> getPerformanceAnalytics(AnalyticsFilter filter) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                HttpUrl.Builder urlBuilder = HttpUrl.parse(client.getApiUrl() + "/api/v1/telemetry/analytics").newBuilder();
                
                if (filter != null) {
                    if (filter.getStartDate() != null) {
                        urlBuilder.addQueryParameter("start_date", filter.getStartDate().toString());
                    }
                    if (filter.getEndDate() != null) {
                        urlBuilder.addQueryParameter("end_date", filter.getEndDate().toString());
                    }
                    if (filter.getMetric() != null) {
                        urlBuilder.addQueryParameter("metric", filter.getMetric());
                    }
                    if (filter.getInterval() != null) {
                        urlBuilder.addQueryParameter("interval", filter.getInterval());
                    }
                }
                
                Request request = new Request.Builder()
                        .url(urlBuilder.build())
                        .get()
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get performance analytics: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, PerformanceAnalytics.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Error getting performance analytics", e);
            }
        });
    }
    
    /**
     * Send custom metrics.
     */
    public CompletableFuture<Void> sendMetrics(MetricsData metrics) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                String json = gson.toJson(metrics);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/api/v1/telemetry/metrics")
                        .post(body)
                        .build();
                
                try (Response response = httpClient.newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to send metrics: " + response.code());
                    }
                    return null;
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Error sending metrics", e);
            }
        });
    }
    
    /**
     * System metrics model.
     */
    public static class SystemMetrics {
        private CpuMetrics cpu;
        private MemoryMetrics memory;
        private NetworkMetrics network;
        private StorageMetrics storage;
        private LocalDateTime timestamp;
        
        // Getters and setters
        public CpuMetrics getCpu() { return cpu; }
        public void setCpu(CpuMetrics cpu) { this.cpu = cpu; }
        
        public MemoryMetrics getMemory() { return memory; }
        public void setMemory(MemoryMetrics memory) { this.memory = memory; }
        
        public NetworkMetrics getNetwork() { return network; }
        public void setNetwork(NetworkMetrics network) { this.network = network; }
        
        public StorageMetrics getStorage() { return storage; }
        public void setStorage(StorageMetrics storage) { this.storage = storage; }
        
        public LocalDateTime getTimestamp() { return timestamp; }
        public void setTimestamp(LocalDateTime timestamp) { this.timestamp = timestamp; }
    }
    
    /**
     * CPU metrics.
     */
    public static class CpuMetrics {
        private Double usagePercent;
        private Integer cores;
        private Double temperature;
        private Double frequency;
        
        // Getters and setters
        public Double getUsagePercent() { return usagePercent; }
        public void setUsagePercent(Double usagePercent) { this.usagePercent = usagePercent; }
        
        public Integer getCores() { return cores; }
        public void setCores(Integer cores) { this.cores = cores; }
        
        public Double getTemperature() { return temperature; }
        public void setTemperature(Double temperature) { this.temperature = temperature; }
        
        public Double getFrequency() { return frequency; }
        public void setFrequency(Double frequency) { this.frequency = frequency; }
    }
    
    /**
     * Memory metrics.
     */
    public static class MemoryMetrics {
        private Long total;
        private Long used;
        private Long available;
        private Double usagePercent;
        
        // Getters and setters
        public Long getTotal() { return total; }
        public void setTotal(Long total) { this.total = total; }
        
        public Long getUsed() { return used; }
        public void setUsed(Long used) { this.used = used; }
        
        public Long getAvailable() { return available; }
        public void setAvailable(Long available) { this.available = available; }
        
        public Double getUsagePercent() { return usagePercent; }
        public void setUsagePercent(Double usagePercent) { this.usagePercent = usagePercent; }
    }
    
    /**
     * Network metrics.
     */
    public static class NetworkMetrics {
        private Long bytesReceived;
        private Long bytesSent;
        private Integer packetsReceived;
        private Integer packetsSent;
        private Double bandwidth;
        
        // Getters and setters
        public Long getBytesReceived() { return bytesReceived; }
        public void setBytesReceived(Long bytesReceived) { this.bytesReceived = bytesReceived; }
        
        public Long getBytesSent() { return bytesSent; }
        public void setBytesSent(Long bytesSent) { this.bytesSent = bytesSent; }
        
        public Integer getPacketsReceived() { return packetsReceived; }
        public void setPacketsReceived(Integer packetsReceived) { this.packetsReceived = packetsReceived; }
        
        public Integer getPacketsSent() { return packetsSent; }
        public void setPacketsSent(Integer packetsSent) { this.packetsSent = packetsSent; }
        
        public Double getBandwidth() { return bandwidth; }
        public void setBandwidth(Double bandwidth) { this.bandwidth = bandwidth; }
    }
    
    /**
     * Storage metrics.
     */
    public static class StorageMetrics {
        private Long total;
        private Long used;
        private Long available;
        private Double usagePercent;
        private Integer iops;
        
        // Getters and setters
        public Long getTotal() { return total; }
        public void setTotal(Long total) { this.total = total; }
        
        public Long getUsed() { return used; }
        public void setUsed(Long used) { this.used = used; }
        
        public Long getAvailable() { return available; }
        public void setAvailable(Long available) { this.available = available; }
        
        public Double getUsagePercent() { return usagePercent; }
        public void setUsagePercent(Double usagePercent) { this.usagePercent = usagePercent; }
        
        public Integer getIops() { return iops; }
        public void setIops(Integer iops) { this.iops = iops; }
    }
    
    /**
     * Job metrics model.
     */
    public static class JobMetrics {
        private String jobId;
        private LocalDateTime timestamp;
        private CpuMetrics cpu;
        private MemoryMetrics memory;
        private NetworkMetrics network;
        private Map<String, Object> customMetrics;
        
        // Getters and setters
        public String getJobId() { return jobId; }
        public void setJobId(String jobId) { this.jobId = jobId; }
        
        public LocalDateTime getTimestamp() { return timestamp; }
        public void setTimestamp(LocalDateTime timestamp) { this.timestamp = timestamp; }
        
        public CpuMetrics getCpu() { return cpu; }
        public void setCpu(CpuMetrics cpu) { this.cpu = cpu; }
        
        public MemoryMetrics getMemory() { return memory; }
        public void setMemory(MemoryMetrics memory) { this.memory = memory; }
        
        public NetworkMetrics getNetwork() { return network; }
        public void setNetwork(NetworkMetrics network) { this.network = network; }
        
        public Map<String, Object> getCustomMetrics() { return customMetrics; }
        public void setCustomMetrics(Map<String, Object> customMetrics) { this.customMetrics = customMetrics; }
    }
    
    /**
     * Resource utilization model.
     */
    public static class ResourceUtilization {
        private String resourceId;
        private LocalDateTime timestamp;
        private Double utilizationPercent;
        private SystemMetrics metrics;
        private Map<String, Object> metadata;
        
        // Getters and setters
        public String getResourceId() { return resourceId; }
        public void setResourceId(String resourceId) { this.resourceId = resourceId; }
        
        public LocalDateTime getTimestamp() { return timestamp; }
        public void setTimestamp(LocalDateTime timestamp) { this.timestamp = timestamp; }
        
        public Double getUtilizationPercent() { return utilizationPercent; }
        public void setUtilizationPercent(Double utilizationPercent) { this.utilizationPercent = utilizationPercent; }
        
        public SystemMetrics getMetrics() { return metrics; }
        public void setMetrics(SystemMetrics metrics) { this.metrics = metrics; }
        
        public Map<String, Object> getMetadata() { return metadata; }
        public void setMetadata(Map<String, Object> metadata) { this.metadata = metadata; }
    }
    
    /**
     * Performance analytics model.
     */
    public static class PerformanceAnalytics {
        private List<DataPoint> dataPoints;
        private String metric;
        private String interval;
        private LocalDateTime startDate;
        private LocalDateTime endDate;
        private Map<String, Object> summary;
        
        // Getters and setters
        public List<DataPoint> getDataPoints() { return dataPoints; }
        public void setDataPoints(List<DataPoint> dataPoints) { this.dataPoints = dataPoints; }
        
        public String getMetric() { return metric; }
        public void setMetric(String metric) { this.metric = metric; }
        
        public String getInterval() { return interval; }
        public void setInterval(String interval) { this.interval = interval; }
        
        public LocalDateTime getStartDate() { return startDate; }
        public void setStartDate(LocalDateTime startDate) { this.startDate = startDate; }
        
        public LocalDateTime getEndDate() { return endDate; }
        public void setEndDate(LocalDateTime endDate) { this.endDate = endDate; }
        
        public Map<String, Object> getSummary() { return summary; }
        public void setSummary(Map<String, Object> summary) { this.summary = summary; }
    }
    
    /**
     * Data point for analytics.
     */
    public static class DataPoint {
        private LocalDateTime timestamp;
        private Double value;
        private Map<String, Object> metadata;
        
        // Getters and setters
        public LocalDateTime getTimestamp() { return timestamp; }
        public void setTimestamp(LocalDateTime timestamp) { this.timestamp = timestamp; }
        
        public Double getValue() { return value; }
        public void setValue(Double value) { this.value = value; }
        
        public Map<String, Object> getMetadata() { return metadata; }
        public void setMetadata(Map<String, Object> metadata) { this.metadata = metadata; }
    }
    
    /**
     * Analytics filter.
     */
    public static class AnalyticsFilter {
        private LocalDateTime startDate;
        private LocalDateTime endDate;
        private String metric;
        private String interval;
        
        // Getters and setters
        public LocalDateTime getStartDate() { return startDate; }
        public void setStartDate(LocalDateTime startDate) { this.startDate = startDate; }
        
        public LocalDateTime getEndDate() { return endDate; }
        public void setEndDate(LocalDateTime endDate) { this.endDate = endDate; }
        
        public String getMetric() { return metric; }
        public void setMetric(String metric) { this.metric = metric; }
        
        public String getInterval() { return interval; }
        public void setInterval(String interval) { this.interval = interval; }
    }
    
    /**
     * Metrics data for sending custom metrics.
     */
    public static class MetricsData {
        private String source;
        private LocalDateTime timestamp;
        private Map<String, Object> metrics;
        private Map<String, Object> tags;
        
        // Getters and setters
        public String getSource() { return source; }
        public void setSource(String source) { this.source = source; }
        
        public LocalDateTime getTimestamp() { return timestamp; }
        public void setTimestamp(LocalDateTime timestamp) { this.timestamp = timestamp; }
        
        public Map<String, Object> getMetrics() { return metrics; }
        public void setMetrics(Map<String, Object> metrics) { this.metrics = metrics; }
        
        public Map<String, Object> getTags() { return tags; }
        public void setTags(Map<String, Object> tags) { this.tags = tags; }
    }
} 