package io.computehive.sdk.examples;

import io.computehive.sdk.ComputeHiveClient;
import io.computehive.sdk.marketplace.MarketplaceService;
import io.computehive.sdk.marketplace.MarketplaceService.ComputeResource;
import io.computehive.sdk.marketplace.MarketplaceService.ResourceReservation;
import io.computehive.sdk.marketplace.MarketplaceService.PricingInfo;

import java.time.Duration;
import java.util.List;
import java.util.concurrent.CompletableFuture;

/**
 * Example demonstrating marketplace operations.
 */
public class MarketplaceExample {
    
    public static void main(String[] args) {
        // Create client
        ComputeHiveClient client = ComputeHiveClient.builder()
                .apiKey("your-api-key-here")
                .timeout(Duration.ofMinutes(5))
                .build();
        
        try {
            // Browse available resources
            browseResources(client);
            
            // Get pricing information
            getPricing(client);
            
            // Reserve a resource (commented out to avoid actual reservation)
            // reserveResource(client);
            
            // List user's reservations
            listReservations(client);
            
        } catch (Exception e) {
            System.err.println("Error: " + e.getMessage());
            e.printStackTrace();
        } finally {
            client.close();
        }
    }
    
    /**
     * Browse available compute resources.
     */
    private static void browseResources(ComputeHiveClient client) {
        System.out.println("=== Available Resources ===");
        
        try {
            // List all available resources
            List<ComputeResource> resources = client.marketplace()
                    .listResources(null, null, null, true)
                    .join();
            
            System.out.println("Found " + resources.size() + " available resources:");
            System.out.println();
            
            for (ComputeResource resource : resources) {
                System.out.println("Resource ID: " + resource.getId());
                System.out.println("Name: " + resource.getName());
                System.out.println("Region: " + resource.getRegion());
                System.out.println("Instance Type: " + resource.getInstanceType());
                System.out.println("CPU Cores: " + resource.getCpuCores());
                System.out.println("Memory: " + resource.getMemoryGB() + " GB");
                
                if (resource.getGpuCount() > 0) {
                    System.out.println("GPU: " + resource.getGpuCount() + "x " + resource.getGpuType());
                }
                
                System.out.println("Hourly Rate: $" + resource.getHourlyRate() + " " + resource.getCurrency());
                System.out.println("Provider: " + resource.getProvider());
                System.out.println("Status: " + resource.getStatus());
                System.out.println("---");
            }
            
        } catch (Exception e) {
            System.err.println("Error browsing resources: " + e.getMessage());
        }
    }
    
    /**
     * Get pricing information for different regions and instance types.
     */
    private static void getPricing(ComputeHiveClient client) {
        System.out.println("\n=== Pricing Information ===");
        
        try {
            // Get pricing for different regions
            String[] regions = {"us-west-1", "us-east-1", "eu-west-1"};
            String[] instanceTypes = {"cpu-instance", "gpu-instance", "high-memory"};
            
            for (String region : regions) {
                System.out.println("\nRegion: " + region);
                System.out.println("------------------------");
                
                for (String instanceType : instanceTypes) {
                    List<PricingInfo> pricing = client.marketplace()
                            .getPricing(region, instanceType)
                            .join();
                    
                    if (!pricing.isEmpty()) {
                        PricingInfo info = pricing.get(0);
                        System.out.println("Instance Type: " + info.getInstanceType());
                        System.out.println("  On-Demand: $" + info.getOnDemandPrice() + "/hour");
                        System.out.println("  Spot: $" + info.getSpotPrice() + "/hour");
                        System.out.println("  Reserved: $" + info.getReservedPrice() + "/hour");
                        System.out.println("  Provider: " + info.getProvider());
                    }
                }
            }
            
        } catch (Exception e) {
            System.err.println("Error getting pricing: " + e.getMessage());
        }
    }
    
    /**
     * Reserve a compute resource.
     */
    private static void reserveResource(ComputeHiveClient client) {
        System.out.println("\n=== Resource Reservation ===");
        
        try {
            // First, find an available resource
            List<ComputeResource> resources = client.marketplace()
                    .listResources("us-west-1", null, null, true)
                    .join();
            
            if (resources.isEmpty()) {
                System.out.println("No available resources found");
                return;
            }
            
            ComputeResource resource = resources.get(0);
            System.out.println("Reserving resource: " + resource.getName());
            System.out.println("Resource ID: " + resource.getId());
            
            // Reserve for 24 hours
            ResourceReservation reservation = client.marketplace()
                    .reserveResource(resource.getId(), 24)
                    .join();
            
            System.out.println("Reservation successful!");
            System.out.println("Reservation ID: " + reservation.getId());
            System.out.println("Start Time: " + reservation.getStartTime());
            System.out.println("End Time: " + reservation.getEndTime());
            System.out.println("Duration: " + reservation.getDuration() + " hours");
            System.out.println("Total Cost: $" + reservation.getTotalCost() + " " + reservation.getCurrency());
            System.out.println("Status: " + reservation.getStatus());
            
            // Note: In a real application, you might want to cancel the reservation
            // client.marketplace().cancelReservation(reservation.getId()).join();
            
        } catch (Exception e) {
            System.err.println("Error reserving resource: " + e.getMessage());
        }
    }
    
    /**
     * List user's reservations.
     */
    private static void listReservations(ComputeHiveClient client) {
        System.out.println("\n=== User Reservations ===");
        
        try {
            List<ResourceReservation> reservations = client.marketplace()
                    .listReservations(null)
                    .join();
            
            if (reservations.isEmpty()) {
                System.out.println("No reservations found");
                return;
            }
            
            System.out.println("Found " + reservations.size() + " reservations:");
            System.out.println();
            
            for (ResourceReservation reservation : reservations) {
                System.out.println("Reservation ID: " + reservation.getId());
                System.out.println("Resource ID: " + reservation.getResourceId());
                System.out.println("Status: " + reservation.getStatus());
                System.out.println("Start Time: " + reservation.getStartTime());
                System.out.println("End Time: " + reservation.getEndTime());
                System.out.println("Duration: " + reservation.getDuration() + " hours");
                System.out.println("Total Cost: $" + reservation.getTotalCost() + " " + reservation.getCurrency());
                System.out.println("Created At: " + reservation.getCreatedAt());
                System.out.println("---");
            }
            
        } catch (Exception e) {
            System.err.println("Error listing reservations: " + e.getMessage());
        }
    }
    
    /**
     * Example of cost optimization by comparing different instance types.
     */
    private static void costOptimizationExample(ComputeHiveClient client) {
        System.out.println("\n=== Cost Optimization Analysis ===");
        
        try {
            // Compare different instance types for cost optimization
            String[] instanceTypes = {"cpu-instance", "gpu-instance", "high-memory"};
            String region = "us-west-1";
            
            System.out.println("Cost comparison for region: " + region);
            System.out.println("Instance Type | On-Demand | Spot | Reserved | Savings");
            System.out.println("-------------|-----------|------|----------|--------");
            
            for (String instanceType : instanceTypes) {
                List<PricingInfo> pricing = client.marketplace()
                        .getPricing(region, instanceType)
                        .join();
                
                if (!pricing.isEmpty()) {
                    PricingInfo info = pricing.get(0);
                    double onDemand = info.getOnDemandPrice();
                    double spot = info.getSpotPrice();
                    double reserved = info.getReservedPrice();
                    
                    double spotSavings = ((onDemand - spot) / onDemand) * 100;
                    double reservedSavings = ((onDemand - reserved) / onDemand) * 100;
                    
                    System.out.printf("%-13s | $%-8.2f | $%-4.2f | $%-8.2f | Spot: %.1f%%, Reserved: %.1f%%%n",
                            info.getInstanceType(), onDemand, spot, reserved, spotSavings, reservedSavings);
                }
            }
            
        } catch (Exception e) {
            System.err.println("Error in cost optimization analysis: " + e.getMessage());
        }
    }
} 