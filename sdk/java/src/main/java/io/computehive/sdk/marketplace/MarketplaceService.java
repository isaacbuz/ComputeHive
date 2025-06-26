package io.computehive.sdk.marketplace;

import com.google.gson.Gson;
import io.computehive.sdk.ComputeHiveClient;
import io.computehive.sdk.exceptions.ComputeHiveException;
import okhttp3.*;
import okhttp3.MediaType;

import java.io.IOException;
import java.util.List;
import java.util.concurrent.CompletableFuture;

/**
 * Service for marketplace operations.
 */
public class MarketplaceService {
    
    private static final MediaType JSON = MediaType.get("application/json; charset=utf-8");
    
    private final ComputeHiveClient client;
    private final Gson gson;
    
    public MarketplaceService(ComputeHiveClient client) {
        this.client = client;
        this.gson = client.getGson();
    }
    
    /**
     * List available compute resources.
     * 
     * @param region Filter by region
     * @param instanceType Filter by instance type
     * @param gpuType Filter by GPU type
     * @param available Filter by availability
     * @return CompletableFuture containing the list of resources
     */
    public CompletableFuture<List<ComputeResource>> listResources(String region, String instanceType, String gpuType, Boolean available) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                HttpUrl.Builder urlBuilder = HttpUrl.parse(client.getApiUrl() + "/marketplace/resources").newBuilder();
                
                if (region != null) {
                    urlBuilder.addQueryParameter("region", region);
                }
                if (instanceType != null) {
                    urlBuilder.addQueryParameter("instanceType", instanceType);
                }
                if (gpuType != null) {
                    urlBuilder.addQueryParameter("gpuType", gpuType);
                }
                if (available != null) {
                    urlBuilder.addQueryParameter("available", available.toString());
                }
                
                Request request = new Request.Builder()
                        .url(urlBuilder.build())
                        .get()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to list resources: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    ResourceListResponse resourceListResponse = gson.fromJson(responseBody, ResourceListResponse.class);
                    return resourceListResponse.getResources();
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to list resources", e);
            }
        });
    }
    
    /**
     * Get resource details by ID.
     * 
     * @param resourceId The resource ID
     * @return CompletableFuture containing the resource details
     */
    public CompletableFuture<ComputeResource> getResource(String resourceId) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/marketplace/resources/" + resourceId)
                        .get()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get resource: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, ComputeResource.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to get resource", e);
            }
        });
    }
    
    /**
     * Reserve a compute resource.
     * 
     * @param resourceId The resource ID to reserve
     * @param duration Duration in hours
     * @return CompletableFuture containing the reservation
     */
    public CompletableFuture<ResourceReservation> reserveResource(String resourceId, int duration) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                ReservationRequest requestBody = ReservationRequest.builder()
                        .resourceId(resourceId)
                        .duration(duration)
                        .build();
                
                String json = gson.toJson(requestBody);
                RequestBody body = RequestBody.create(json, JSON);
                
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/marketplace/reservations")
                        .post(body)
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to reserve resource: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    return gson.fromJson(responseBody, ResourceReservation.class);
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to reserve resource", e);
            }
        });
    }
    
    /**
     * Cancel a resource reservation.
     * 
     * @param reservationId The reservation ID to cancel
     * @return CompletableFuture that completes when reservation is cancelled
     */
    public CompletableFuture<Void> cancelReservation(String reservationId) {
        return CompletableFuture.runAsync(() -> {
            try {
                Request request = new Request.Builder()
                        .url(client.getApiUrl() + "/marketplace/reservations/" + reservationId)
                        .delete()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to cancel reservation: " + response.code());
                    }
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to cancel reservation", e);
            }
        });
    }
    
    /**
     * List user's reservations.
     * 
     * @param status Filter by reservation status
     * @return CompletableFuture containing the list of reservations
     */
    public CompletableFuture<List<ResourceReservation>> listReservations(ReservationStatus status) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                HttpUrl.Builder urlBuilder = HttpUrl.parse(client.getApiUrl() + "/marketplace/reservations").newBuilder();
                
                if (status != null) {
                    urlBuilder.addQueryParameter("status", status.name());
                }
                
                Request request = new Request.Builder()
                        .url(urlBuilder.build())
                        .get()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to list reservations: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    ReservationListResponse reservationListResponse = gson.fromJson(responseBody, ReservationListResponse.class);
                    return reservationListResponse.getReservations();
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to list reservations", e);
            }
        });
    }
    
    /**
     * Get marketplace pricing.
     * 
     * @param region Filter by region
     * @param instanceType Filter by instance type
     * @return CompletableFuture containing the pricing information
     */
    public CompletableFuture<List<PricingInfo>> getPricing(String region, String instanceType) {
        return CompletableFuture.supplyAsync(() -> {
            try {
                HttpUrl.Builder urlBuilder = HttpUrl.parse(client.getApiUrl() + "/marketplace/pricing").newBuilder();
                
                if (region != null) {
                    urlBuilder.addQueryParameter("region", region);
                }
                if (instanceType != null) {
                    urlBuilder.addQueryParameter("instanceType", instanceType);
                }
                
                Request request = new Request.Builder()
                        .url(urlBuilder.build())
                        .get()
                        .build();
                
                try (Response response = client.getHttpClient().newCall(request).execute()) {
                    if (!response.isSuccessful()) {
                        throw new ComputeHiveException("Failed to get pricing: " + response.code());
                    }
                    
                    String responseBody = response.body().string();
                    PricingListResponse pricingListResponse = gson.fromJson(responseBody, PricingListResponse.class);
                    return pricingListResponse.getPricing();
                }
            } catch (IOException e) {
                throw new ComputeHiveException("Failed to get pricing", e);
            }
        });
    }
    
    // Helper classes
    private static class ResourceListResponse {
        private List<ComputeResource> resources;
        private int total;
        
        public List<ComputeResource> getResources() { return resources; }
        public void setResources(List<ComputeResource> resources) { this.resources = resources; }
        public int getTotal() { return total; }
        public void setTotal(int total) { this.total = total; }
    }
    
    private static class ReservationListResponse {
        private List<ResourceReservation> reservations;
        private int total;
        
        public List<ResourceReservation> getReservations() { return reservations; }
        public void setReservations(List<ResourceReservation> reservations) { this.reservations = reservations; }
        public int getTotal() { return total; }
        public void setTotal(int total) { this.total = total; }
    }
    
    private static class PricingListResponse {
        private List<PricingInfo> pricing;
        
        public List<PricingInfo> getPricing() { return pricing; }
        public void setPricing(List<PricingInfo> pricing) { this.pricing = pricing; }
    }
    
    /**
     * Compute resource information.
     */
    public static class ComputeResource {
        private String id;
        private String name;
        private String region;
        private String zone;
        private String instanceType;
        private int cpuCores;
        private long memoryGB;
        private int gpuCount;
        private String gpuType;
        private double hourlyRate;
        private String currency;
        private boolean available;
        private String provider;
        private String status;
        private long lastUpdated;
        
        // Getters and setters
        public String getId() { return id; }
        public void setId(String id) { this.id = id; }
        public String getName() { return name; }
        public void setName(String name) { this.name = name; }
        public String getRegion() { return region; }
        public void setRegion(String region) { this.region = region; }
        public String getZone() { return zone; }
        public void setZone(String zone) { this.zone = zone; }
        public String getInstanceType() { return instanceType; }
        public void setInstanceType(String instanceType) { this.instanceType = instanceType; }
        public int getCpuCores() { return cpuCores; }
        public void setCpuCores(int cpuCores) { this.cpuCores = cpuCores; }
        public long getMemoryGB() { return memoryGB; }
        public void setMemoryGB(long memoryGB) { this.memoryGB = memoryGB; }
        public int getGpuCount() { return gpuCount; }
        public void setGpuCount(int gpuCount) { this.gpuCount = gpuCount; }
        public String getGpuType() { return gpuType; }
        public void setGpuType(String gpuType) { this.gpuType = gpuType; }
        public double getHourlyRate() { return hourlyRate; }
        public void setHourlyRate(double hourlyRate) { this.hourlyRate = hourlyRate; }
        public String getCurrency() { return currency; }
        public void setCurrency(String currency) { this.currency = currency; }
        public boolean isAvailable() { return available; }
        public void setAvailable(boolean available) { this.available = available; }
        public String getProvider() { return provider; }
        public void setProvider(String provider) { this.provider = provider; }
        public String getStatus() { return status; }
        public void setStatus(String status) { this.status = status; }
        public long getLastUpdated() { return lastUpdated; }
        public void setLastUpdated(long lastUpdated) { this.lastUpdated = lastUpdated; }
    }
    
    /**
     * Resource reservation information.
     */
    public static class ResourceReservation {
        private String id;
        private String resourceId;
        private String userId;
        private ReservationStatus status;
        private long startTime;
        private long endTime;
        private int duration;
        private double totalCost;
        private String currency;
        private long createdAt;
        
        // Getters and setters
        public String getId() { return id; }
        public void setId(String id) { this.id = id; }
        public String getResourceId() { return resourceId; }
        public void setResourceId(String resourceId) { this.resourceId = resourceId; }
        public String getUserId() { return userId; }
        public void setUserId(String userId) { this.userId = userId; }
        public ReservationStatus getStatus() { return status; }
        public void setStatus(ReservationStatus status) { this.status = status; }
        public long getStartTime() { return startTime; }
        public void setStartTime(long startTime) { this.startTime = startTime; }
        public long getEndTime() { return endTime; }
        public void setEndTime(long endTime) { this.endTime = endTime; }
        public int getDuration() { return duration; }
        public void setDuration(int duration) { this.duration = duration; }
        public double getTotalCost() { return totalCost; }
        public void setTotalCost(double totalCost) { this.totalCost = totalCost; }
        public String getCurrency() { return currency; }
        public void setCurrency(String currency) { this.currency = currency; }
        public long getCreatedAt() { return createdAt; }
        public void setCreatedAt(long createdAt) { this.createdAt = createdAt; }
    }
    
    /**
     * Reservation status enumeration.
     */
    public enum ReservationStatus {
        PENDING,
        ACTIVE,
        COMPLETED,
        CANCELLED,
        EXPIRED
    }
    
    /**
     * Pricing information.
     */
    public static class PricingInfo {
        private String region;
        private String instanceType;
        private double onDemandPrice;
        private double spotPrice;
        private double reservedPrice;
        private String currency;
        private String provider;
        
        // Getters and setters
        public String getRegion() { return region; }
        public void setRegion(String region) { this.region = region; }
        public String getInstanceType() { return instanceType; }
        public void setInstanceType(String instanceType) { this.instanceType = instanceType; }
        public double getOnDemandPrice() { return onDemandPrice; }
        public void setOnDemandPrice(double onDemandPrice) { this.onDemandPrice = onDemandPrice; }
        public double getSpotPrice() { return spotPrice; }
        public void setSpotPrice(double spotPrice) { this.spotPrice = spotPrice; }
        public double getReservedPrice() { return reservedPrice; }
        public void setReservedPrice(double reservedPrice) { this.reservedPrice = reservedPrice; }
        public String getCurrency() { return currency; }
        public void setCurrency(String currency) { this.currency = currency; }
        public String getProvider() { return provider; }
        public void setProvider(String provider) { this.provider = provider; }
    }
    
    /**
     * Reservation request.
     */
    @lombok.Data
    @lombok.Builder
    @lombok.NoArgsConstructor
    @lombok.AllArgsConstructor
    private static class ReservationRequest {
        private String resourceId;
        private int duration;
    }
} 