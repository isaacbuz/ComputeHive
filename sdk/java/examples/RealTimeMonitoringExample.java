package io.computehive.sdk.examples;

import io.computehive.sdk.ComputeHiveClient;
import io.computehive.sdk.events.EventListener;
import io.computehive.sdk.telemetry.TelemetryService;
import io.computehive.sdk.telemetry.TelemetryService.MetricsData;
import io.computehive.sdk.telemetry.TelemetryService.LogData;
import io.computehive.sdk.telemetry.TelemetryService.EventData;
import io.computehive.sdk.telemetry.TelemetryService.SystemMetrics;
import io.computehive.sdk.telemetry.TelemetryService.Alert;
import io.computehive.sdk.telemetry.TelemetryService.TimeRange;
import io.computehive.sdk.telemetry.TelemetryService.AlertSeverity;

import java.time.Duration;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.CompletableFuture;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.TimeUnit;

/**
 * Example demonstrating real-time monitoring and telemetry.
 */
public class RealTimeMonitoringExample {
    
    private static final ScheduledExecutorService scheduler = Executors.newScheduledThreadPool(2);
    
    public static void main(String[] args) {
        // Create client
        ComputeHiveClient client = ComputeHiveClient.builder()
                .apiKey("your-api-key-here")
                .timeout(Duration.ofMinutes(5))
                .build();
        
        try {
            // Connect to WebSocket for real-time events
            connectToWebSocket(client);
            
            // Start telemetry monitoring
            startTelemetryMonitoring(client);
            
            // Monitor system metrics
            monitorSystemMetrics(client);
            
            // Check for alerts
            checkAlerts(client);
            
            // Keep the application running
            System.out.println("Real-time monitoring active. Press Ctrl+C to exit.");
            Thread.sleep(Long.MAX_VALUE);
            
        } catch (InterruptedException e) {
            System.out.println("Monitoring interrupted");
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
            e.printStackTrace();
        } finally {
            // Cleanup
            scheduler.shutdown();
            client.close();
        }
    }
    
    /**
     * Connect to WebSocket and subscribe to real-time events.
     */
    private static void connectToWebSocket(ComputeHiveClient client) {
        System.out.println("=== Connecting to WebSocket ===");
        
        // Connect to WebSocket
        client.connect();
        
        // Wait for connection
        try {
            Thread.sleep(2000);
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            return;
        }
        
        if (!client.isConnected()) {
            System.err.println("Failed to connect to WebSocket");
            return;
        }
        
        System.out.println("✅ Connected to WebSocket");
        
        // Subscribe to various events
        subscribeToEvents(client);
    }
    
    /**
     * Subscribe to different types of events.
     */
    private static void subscribeToEvents(ComputeHiveClient client) {
        System.out.println("=== Subscribing to Events ===");
        
        // Job events
        client.events().on("job.started", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.out.println("🚀 Job Started: " + data);
            }
        });
        
        client.events().on("job.completed", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.out.println("✅ Job Completed: " + data);
            }
        });
        
        client.events().on("job.failed", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.err.println("❌ Job Failed: " + data);
            }
        });
        
        client.events().on("job.progress", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.out.println("📊 Job Progress: " + data);
            }
        });
        
        // Resource events
        client.events().on("resource.available", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.out.println("🖥️ Resource Available: " + data);
            }
        });
        
        client.events().on("resource.allocated", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.out.println("🔒 Resource Allocated: " + data);
            }
        });
        
        // System events
        client.events().on("system.alert", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.err.println("🚨 System Alert: " + data);
            }
        });
        
        client.events().on("system.maintenance", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.out.println("🔧 System Maintenance: " + data);
            }
        });
        
        // Payment events
        client.events().on("payment.processed", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.out.println("💳 Payment Processed: " + data);
            }
        });
        
        client.events().on("payment.failed", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.err.println("💸 Payment Failed: " + data);
            }
        });
        
        // Connection events
        client.events().on("connected", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.out.println("🔗 WebSocket Connected");
            }
        });
        
        client.events().on("disconnected", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.err.println("🔌 WebSocket Disconnected");
            }
        });
        
        client.events().on("error", new EventListener() {
            @Override
            public void onEvent(String eventType, Object data) {
                System.err.println("⚠️ WebSocket Error: " + data);
            }
        });
        
        System.out.println("✅ Subscribed to all events");
    }
    
    /**
     * Start periodic telemetry monitoring.
     */
    private static void startTelemetryMonitoring(ComputeHiveClient client) {
        System.out.println("=== Starting Telemetry Monitoring ===");
        
        // Schedule periodic metrics collection
        scheduler.scheduleAtFixedRate(() -> {
            try {
                sendMetrics(client);
            } catch (Exception e) {
                System.err.println("Error sending metrics: " + e.getMessage());
            }
        }, 0, 30, TimeUnit.SECONDS);
        
        // Schedule periodic log collection
        scheduler.scheduleAtFixedRate(() -> {
            try {
                sendLogs(client);
            } catch (Exception e) {
                System.err.println("Error sending logs: " + e.getMessage());
            }
        }, 10, 60, TimeUnit.SECONDS);
        
        // Schedule periodic event sending
        scheduler.scheduleAtFixedRate(() -> {
            try {
                sendEvents(client);
            } catch (Exception e) {
                System.err.println("Error sending events: " + e.getMessage());
            }
        }, 20, 120, TimeUnit.SECONDS);
        
        System.out.println("✅ Telemetry monitoring started");
    }
    
    /**
     * Send custom metrics.
     */
    private static void sendMetrics(ComputeHiveClient client) {
        MetricsData metrics = new MetricsData();
        metrics.setJobId("example-job-id");
        metrics.setNodeId("example-node-id");
        metrics.setTimestamp(System.currentTimeMillis());
        
        Map<String, Double> metricValues = new HashMap<>();
        metricValues.put("cpu_usage", Math.random() * 100);
        metricValues.put("memory_usage", Math.random() * 100);
        metricValues.put("disk_usage", Math.random() * 100);
        metricValues.put("network_usage", Math.random() * 100);
        metricValues.put("gpu_usage", Math.random() * 100);
        metrics.setMetrics(metricValues);
        
        Map<String, String> tags = new HashMap<>();
        tags.put("environment", "production");
        tags.put("region", "us-west-1");
        tags.put("instance_type", "gpu-instance");
        metrics.setTags(tags);
        
        client.telemetry().sendMetrics(metrics).join();
        System.out.println("📊 Sent metrics: " + metricValues);
    }
    
    /**
     * Send application logs.
     */
    private static void sendLogs(ComputeHiveClient client) {
        LogData logData = new LogData();
        logData.setJobId("example-job-id");
        logData.setNodeId("example-node-id");
        logData.setTimestamp(System.currentTimeMillis());
        logData.setLevel("INFO");
        logData.setMessage("Application heartbeat - all systems operational");
        logData.setSource("RealTimeMonitoringExample");
        
        Map<String, String> context = new HashMap<>();
        context.put("component", "monitoring");
        context.put("version", "1.0.0");
        logData.setContext(context);
        
        client.telemetry().sendLogs(logData).join();
        System.out.println("📝 Sent log: " + logData.getMessage());
    }
    
    /**
     * Send custom events.
     */
    private static void sendEvents(ComputeHiveClient client) {
        EventData event = new EventData();
        event.setJobId("example-job-id");
        event.setNodeId("example-node-id");
        event.setTimestamp(System.currentTimeMillis());
        event.setEventType("custom");
        event.setEventName("monitoring.heartbeat");
        
        Map<String, Object> eventData = new HashMap<>();
        eventData.put("uptime", System.currentTimeMillis());
        eventData.put("active_connections", 5);
        eventData.put("memory_available", 8192);
        event.setData(eventData);
        
        Map<String, String> tags = new HashMap<>();
        tags.put("service", "monitoring");
        tags.put("environment", "production");
        event.setTags(tags);
        
        client.telemetry().sendEvent(event).join();
        System.out.println("🎯 Sent event: " + event.getEventName());
    }
    
    /**
     * Monitor system metrics.
     */
    private static void monitorSystemMetrics(ComputeHiveClient client) {
        System.out.println("=== Monitoring System Metrics ===");
        
        // Schedule periodic system metrics monitoring
        scheduler.scheduleAtFixedRate(() -> {
            try {
                CompletableFuture<SystemMetrics> future = client.telemetry()
                        .getSystemMetrics(null, TimeRange.LAST_HOUR);
                
                SystemMetrics metrics = future.join();
                
                System.out.println("📈 System Metrics:");
                System.out.println("  CPU Usage: " + String.format("%.1f%%", metrics.getCpuUsage()));
                System.out.println("  Memory Usage: " + String.format("%.1f%%", metrics.getMemoryUsage()));
                System.out.println("  Disk Usage: " + String.format("%.1f%%", metrics.getDiskUsage()));
                System.out.println("  Network Usage: " + String.format("%.1f%%", metrics.getNetworkUsage()));
                System.out.println("  Active Jobs: " + metrics.getActiveJobs());
                System.out.println("  Total Nodes: " + metrics.getTotalNodes());
                System.out.println("  Available Nodes: " + metrics.getAvailableNodes());
                
            } catch (Exception e) {
                System.err.println("Error getting system metrics: " + e.getMessage());
            }
        }, 0, 60, TimeUnit.SECONDS);
        
        System.out.println("✅ System metrics monitoring started");
    }
    
    /**
     * Check for system alerts.
     */
    private static void checkAlerts(ComputeHiveClient client) {
        System.out.println("=== Checking Alerts ===");
        
        // Schedule periodic alert checking
        scheduler.scheduleAtFixedRate(() -> {
            try {
                List<Alert> alerts = client.telemetry()
                        .getAlerts(null, AlertSeverity.ERROR)
                        .join();
                
                if (!alerts.isEmpty()) {
                    System.err.println("🚨 Found " + alerts.size() + " active alerts:");
                    
                    for (Alert alert : alerts) {
                        System.err.println("  Alert: " + alert.getName());
                        System.err.println("  Description: " + alert.getDescription());
                        System.err.println("  Severity: " + alert.getSeverity());
                        System.err.println("  Status: " + alert.getStatus());
                        System.err.println("  Created: " + alert.getCreatedAt());
                        System.err.println("  ---");
                    }
                } else {
                    System.out.println("✅ No active alerts");
                }
                
            } catch (Exception e) {
                System.err.println("Error checking alerts: " + e.getMessage());
            }
        }, 0, 300, TimeUnit.SECONDS); // Check every 5 minutes
        
        System.out.println("✅ Alert monitoring started");
    }
    
    /**
     * Example of performance monitoring.
     */
    private static void performanceMonitoringExample(ComputeHiveClient client) {
        System.out.println("=== Performance Monitoring Example ===");
        
        // Monitor CPU performance
        scheduler.scheduleAtFixedRate(() -> {
            try {
                TelemetryService.PerformanceMetrics cpuMetrics = client.telemetry()
                        .getPerformanceMetrics(null, TelemetryService.MetricType.CPU, TimeRange.LAST_HOUR)
                        .join();
                
                System.out.println("CPU Performance Metrics:");
                System.out.println("  Metric Type: " + cpuMetrics.getMetricType());
                System.out.println("  Unit: " + cpuMetrics.getUnit());
                System.out.println("  Data Points: " + cpuMetrics.getDataPoints().size());
                
                if (!cpuMetrics.getDataPoints().isEmpty()) {
                    TelemetryService.DataPoint latest = cpuMetrics.getDataPoints().get(cpuMetrics.getDataPoints().size() - 1);
                    System.out.println("  Latest Value: " + latest.getValue());
                    System.out.println("  Timestamp: " + latest.getTimestamp());
                }
                
            } catch (Exception e) {
                System.err.println("Error getting CPU metrics: " + e.getMessage());
            }
        }, 0, 120, TimeUnit.SECONDS);
        
        System.out.println("✅ Performance monitoring started");
    }
} 