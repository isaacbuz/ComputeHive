package io.computehive.sdk.models;


import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;

/**
 * Represents a compute job in the ComputeHive platform.
 */
public class Job {
    
    /**
     * Unique job identifier.
     */
    private String id;
    
    /**
     * Job name.
     */
    private String name;
    
    /**
     * Job description.
     */
    private String description;
    
    /**
     * Job status.
     */
    private Status status;
    
    /**
     * Job priority.
     */
    private Priority priority;
    
    /**
     * Job type.
     */
    private JobType type;
    
    /**
     * Container image to use.
     */
    private String image;
    
    /**
     * Container command to execute.
     */
    private List<String> command;
    
    /**
     * Environment variables.
     */
    private Map<String, String> environment;
    
    /**
     * Resource requirements.
     */
    private ResourceRequirements resources;
    
    /**
     * Storage configuration.
     */
    private StorageConfig storage;
    
    /**
     * Network configuration.
     */
    private NetworkConfig network;
    
    /**
     * Security configuration.
     */
    private SecurityConfig security;
    
    /**
     * Scheduling configuration.
     */
    private SchedulingConfig scheduling;
    
    /**
     * Job metadata.
     */
    private Map<String, String> metadata;
    
    /**
     * Job tags.
     */
    private List<String> tags;
    
    /**
     * Job owner/user ID.
     */
    private String userId;
    
    /**
     * Project ID this job belongs to.
     */
    private String projectId;
    
    /**
     * Job creation timestamp.
     */
    private LocalDateTime createdAt;
    
    /**
     * Job start timestamp.
     */
    private LocalDateTime startedAt;
    
    /**
     * Job completion timestamp.
     */
    private LocalDateTime completedAt;
    
    /**
     * Job timeout in seconds.
     */
    private Long timeout;
    
    /**
     * Maximum retry attempts.
     */
    private Integer maxRetries;
    
    /**
     * Current retry count.
     */
    private Integer retryCount;
    
    /**
     * Job exit code.
     */
    private Integer exitCode;
    
    /**
     * Job error message.
     */
    private String errorMessage;
    
    /**
     * Job logs URL.
     */
    private String logsUrl;
    
    /**
     * Job results URL.
     */
    private String resultsUrl;
    
    /**
     * Job cost information.
     */
    private JobCost cost;
    
    /**
     * Job metrics.
     */
    private JobMetrics metrics;
    
    /**
     * Job status enumeration.
     */
    public enum Status {
        PENDING,
        QUEUED,
        RUNNING,
        COMPLETED,
        FAILED,
        CANCELLED,
        TIMEOUT,
        RETRYING
    }
    
    /**
     * Job priority enumeration.
     */
    public enum Priority {
        LOW,
        NORMAL,
        HIGH,
        URGENT
    }
    
    /**
     * Job type enumeration.
     */
    public enum JobType {
        BATCH,
        INTERACTIVE,
        STREAMING,
        ML_TRAINING,
        ML_INFERENCE,
        DATA_PROCESSING,
        SIMULATION,
        RENDERING
    }
    
    /**
     * Resource requirements for the job.
     */
    public static class ResourceRequirements {
        private CPU cpu;
        
        private Memory memory;
        
        private GPU gpu;
        
        private Storage storage;
        
        private Network network;
        
        public static class CPU {
            private Integer cores;
            
            private String architecture;
        }
        
        public static class Memory {
            private Integer size; // in GB
            
            private String type;
        }
        
        public static class GPU {
            private Integer count;
            
            private String model;
            
            private Integer memory; // in GB
            
            private String computeCapability;
        }
        
        public static class Storage {
            private Integer size; // in GB
            
            private StorageType type;
            
            private Integer iops;
            
            public enum StorageType {
                SSD, HDD, NVME
            }
        }
        
        public static class Network {
            private Integer bandwidth; // in Mbps
            
            private Integer latency; // in ms
        }
    }
    
    /**
     * Storage configuration for the job.
     */
    public static class StorageConfig {
        private List<VolumeMount> volumes;
        
        private List<String> persistentVolumes;
        
        private long tempStorageGB;
        
        private boolean enableCaching;
        
        private String cacheStrategy;
    }
    
    /**
     * Volume mount configuration.
     */
    public static class VolumeMount {
        private String name;
        
        private String mountPath;
        
        private String sourcePath;
        
        private boolean readOnly;
        
        private String type; // host, empty, configmap, secret
    }
    
    /**
     * Network configuration for the job.
     */
    public static class NetworkConfig {
        private List<Integer> ports;
        
        private boolean enableLoadBalancer;
        
        private String loadBalancerType;
        
        private List<String> allowedIPs;
        
        private boolean enableVPC;
        
        private String vpcId;
        
        private List<String> securityGroups;
    }
    
    /**
     * Security configuration for the job.
     */
    public static class SecurityConfig {
        private boolean runAsRoot;
        
        private String runAsUser;
        
        private List<String> capabilities;
        
        private boolean enableSeccomp;
        
        private String seccompProfile;
        
        private boolean enableAppArmor;
        
        private String appArmorProfile;
        
        private List<String> allowedSyscalls;
        
        private boolean enableNetworkIsolation;
    }
    
    /**
     * Scheduling configuration for the job.
     */
    public static class SchedulingConfig {
        private String region;
        
        private List<String> zones;
        
        private String instanceType;
        
        private String nodeSelector;
        
        private Map<String, String> nodeAffinity;
        
        private Map<String, String> podAffinity;
        
        private Map<String, String> tolerations;
        
        private boolean enableAutoScaling;
        
        private int minReplicas;
        
        private int maxReplicas;
        
        private double targetCPUUtilization;
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
        
        private LocalDateTime calculatedAt;
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
        
        private LocalDateTime lastUpdated;
    }
} 