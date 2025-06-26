package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/computehive/cli/pkg/client"
	"github.com/computehive/cli/pkg/config"
	"github.com/computehive/cli/pkg/utils"
)

// NewMarketplaceCmd creates the marketplace command
func NewMarketplaceCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "marketplace",
		Short: "Interact with the compute marketplace",
		Long:  "View offers, place bids, and manage marketplace listings",
		Aliases: []string{"market", "mp"},
	}

	cmd.AddCommand(
		newMarketplaceOffersCmd(),
		newMarketplaceBidsCmd(),
		newMarketplaceCreateOfferCmd(),
		newMarketplaceCreateBidCmd(),
		newMarketplacePricesCmd(),
	)

	return cmd
}

// newMarketplaceOffersCmd creates the offers command
func newMarketplaceOffersCmd() *cobra.Command {
	var (
		resourceType string
		location     string
		minCPU       int
		minMemory    int
		minGPU       int
		maxPrice     float64
		limit        int
	)

	cmd := &cobra.Command{
		Use:   "offers",
		Short: "List available resource offers",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			offers, err := apiClient.ListOffers(client.ListOffersOptions{
				ResourceType: resourceType,
				Location:     location,
				MinCPU:       minCPU,
				MinMemory:    minMemory,
				MinGPU:       minGPU,
				MaxPrice:     maxPrice,
				Limit:        limit,
			})
			if err != nil {
				return fmt.Errorf("failed to list offers: %w", err)
			}

			if len(offers) == 0 {
				fmt.Println("No offers found matching criteria")
				return nil
			}

			// Print table
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tPROVIDER\tTYPE\tCPU\tMEMORY\tGPU\tPRICE/HR\tLOCATION\tRATING")
			fmt.Fprintln(w, "--\t--------\t----\t---\t------\t---\t--------\t--------\t------")
			
			for _, offer := range offers {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%dGB\t%d\t$%.2f\t%s\t%.1f★\n",
					offer.ID[:8],
					utils.Truncate(offer.ProviderName, 15),
					offer.ResourceType,
					offer.CPUCores,
					offer.MemoryGB,
					offer.GPUCount,
					offer.PricePerHour,
					offer.Location,
					offer.ReputationScore,
				)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().StringVar(&resourceType, "type", "", "resource type (cpu, gpu)")
	cmd.Flags().StringVar(&location, "location", "", "filter by location")
	cmd.Flags().IntVar(&minCPU, "min-cpu", 0, "minimum CPU cores")
	cmd.Flags().IntVar(&minMemory, "min-memory", 0, "minimum memory in GB")
	cmd.Flags().IntVar(&minGPU, "min-gpu", 0, "minimum GPU count")
	cmd.Flags().Float64Var(&maxPrice, "max-price", 0, "maximum price per hour")
	cmd.Flags().IntVar(&limit, "limit", 50, "maximum number of offers to show")

	return cmd
}

// newMarketplaceBidsCmd creates the bids command
func newMarketplaceBidsCmd() *cobra.Command {
	var (
		status string
		limit  int
		mine   bool
	)

	cmd := &cobra.Command{
		Use:   "bids",
		Short: "List active bids",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			bids, err := apiClient.ListBids(client.ListBidsOptions{
				Status: status,
				Limit:  limit,
				Mine:   mine,
			})
			if err != nil {
				return fmt.Errorf("failed to list bids: %w", err)
			}

			if len(bids) == 0 {
				fmt.Println("No bids found")
				return nil
			}

			// Print table
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tCONSUMER\tCPU\tMEMORY\tGPU\tMAX PRICE\tDURATION\tDEADLINE\tSTATUS")
			fmt.Fprintln(w, "--\t--------\t---\t------\t---\t---------\t--------\t--------\t------")
			
			for _, bid := range bids {
				deadline := bid.Deadline.Format("2006-01-02 15:04")
				if time.Until(bid.Deadline) < 24*time.Hour {
					deadline = fmt.Sprintf("%.1fh", time.Until(bid.Deadline).Hours())
				}
				
				fmt.Fprintf(w, "%s\t%s\t%d\t%dGB\t%d\t$%.2f\t%dh\t%s\t%s\n",
					bid.ID[:8],
					utils.Truncate(bid.ConsumerName, 15),
					bid.Requirements.CPUCores,
					bid.Requirements.MemoryGB,
					bid.Requirements.GPUCount,
					bid.MaxPricePerHour,
					bid.DurationHours,
					deadline,
					bid.Status,
				)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "filter by status (pending, matched, expired)")
	cmd.Flags().IntVar(&limit, "limit", 50, "maximum number of bids to show")
	cmd.Flags().BoolVar(&mine, "mine", false, "show only your bids")

	return cmd
}

// newMarketplaceCreateOfferCmd creates the create-offer command
func newMarketplaceCreateOfferCmd() *cobra.Command {
	var (
		cpuCores         int
		memoryGB         int
		gpuCount         int
		gpuModel         string
		storageGB        int
		networkBandwidth float64
		pricePerHour     float64
		location         string
		duration         string
		autoAccept       bool
	)

	cmd := &cobra.Command{
		Use:   "create-offer",
		Short: "Create a new resource offer",
		Example: `  # Create a CPU offer
  computehive marketplace create-offer --cpu 16 --memory 64 --price 2.50

  # Create a GPU offer
  computehive marketplace create-offer --cpu 32 --memory 128 --gpu 4 --gpu-model "nvidia-a100" --price 15.00`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			// Parse duration
			durationTime, err := time.ParseDuration(duration)
			if err != nil {
				return fmt.Errorf("invalid duration format: %w", err)
			}

			offer := client.CreateOfferRequest{
				CPUCores:         cpuCores,
				MemoryGB:         memoryGB,
				GPUCount:         gpuCount,
				GPUModel:         gpuModel,
				StorageGB:        storageGB,
				NetworkBandwidth: networkBandwidth,
				PricePerHour:     pricePerHour,
				Location:         location,
				Duration:         durationTime,
				AutoAccept:       autoAccept,
			}

			fmt.Println("Creating offer...")
			createdOffer, err := apiClient.CreateOffer(offer)
			if err != nil {
				return fmt.Errorf("failed to create offer: %w", err)
			}

			fmt.Printf("✅ Offer created successfully!\n")
			fmt.Printf("Offer ID: %s\n", createdOffer.ID)
			fmt.Printf("Status: %s\n", createdOffer.Status)
			fmt.Printf("Expires: %s\n", createdOffer.ExpiresAt.Format(time.RFC3339))

			return nil
		},
	}

	cmd.Flags().IntVar(&cpuCores, "cpu", 0, "number of CPU cores (required)")
	cmd.Flags().IntVar(&memoryGB, "memory", 0, "memory in GB (required)")
	cmd.Flags().IntVar(&gpuCount, "gpu", 0, "number of GPUs")
	cmd.Flags().StringVar(&gpuModel, "gpu-model", "", "GPU model (e.g., nvidia-a100)")
	cmd.Flags().IntVar(&storageGB, "storage", 100, "storage in GB")
	cmd.Flags().Float64Var(&networkBandwidth, "network", 1.0, "network bandwidth in Gbps")
	cmd.Flags().Float64Var(&pricePerHour, "price", 0, "price per hour in USD (required)")
	cmd.Flags().StringVar(&location, "location", "", "location/region")
	cmd.Flags().StringVar(&duration, "duration", "24h", "offer duration")
	cmd.Flags().BoolVar(&autoAccept, "auto-accept", false, "automatically accept matching bids")

	cmd.MarkFlagRequired("cpu")
	cmd.MarkFlagRequired("memory")
	cmd.MarkFlagRequired("price")

	return cmd
}

// newMarketplaceCreateBidCmd creates the create-bid command
func newMarketplaceCreateBidCmd() *cobra.Command {
	var (
		cpuCores        int
		memoryGB        int
		gpuCount        int
		gpuModel        string
		storageGB       int
		maxPricePerHour float64
		duration        string
		deadline        string
		location        string
	)

	cmd := &cobra.Command{
		Use:   "create-bid",
		Short: "Create a new resource bid",
		Example: `  # Create a CPU bid
  computehive marketplace create-bid --cpu 8 --memory 32 --max-price 3.00 --duration 4h

  # Create a GPU bid with deadline
  computehive marketplace create-bid --cpu 16 --memory 64 --gpu 2 --max-price 20.00 --duration 8h --deadline 2h`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			// Parse duration
			durationTime, err := time.ParseDuration(duration)
			if err != nil {
				return fmt.Errorf("invalid duration format: %w", err)
			}

			// Parse deadline
			var deadlineTime time.Time
			if deadline != "" {
				deadlineDuration, err := time.ParseDuration(deadline)
				if err != nil {
					return fmt.Errorf("invalid deadline format: %w", err)
				}
				deadlineTime = time.Now().Add(deadlineDuration)
			} else {
				deadlineTime = time.Now().Add(24 * time.Hour) // Default 24h deadline
			}

			bid := client.CreateBidRequest{
				Requirements: client.ResourceRequirements{
					CPUCores:  cpuCores,
					MemoryGB:  memoryGB,
					GPUCount:  gpuCount,
					GPUModel:  gpuModel,
					StorageGB: storageGB,
				},
				MaxPricePerHour: maxPricePerHour,
				DurationHours:   int(durationTime.Hours()),
				Deadline:        deadlineTime,
				Location:        location,
			}

			fmt.Println("Creating bid...")
			createdBid, err := apiClient.CreateBid(bid)
			if err != nil {
				return fmt.Errorf("failed to create bid: %w", err)
			}

			fmt.Printf("✅ Bid created successfully!\n")
			fmt.Printf("Bid ID: %s\n", createdBid.ID)
			fmt.Printf("Status: %s\n", createdBid.Status)
			fmt.Printf("Deadline: %s\n", createdBid.Deadline.Format(time.RFC3339))
			fmt.Printf("\nWe'll notify you when a matching offer is found.\n")

			return nil
		},
	}

	cmd.Flags().IntVar(&cpuCores, "cpu", 0, "number of CPU cores needed (required)")
	cmd.Flags().IntVar(&memoryGB, "memory", 0, "memory in GB needed (required)")
	cmd.Flags().IntVar(&gpuCount, "gpu", 0, "number of GPUs needed")
	cmd.Flags().StringVar(&gpuModel, "gpu-model", "", "preferred GPU model")
	cmd.Flags().IntVar(&storageGB, "storage", 10, "storage in GB needed")
	cmd.Flags().Float64Var(&maxPricePerHour, "max-price", 0, "maximum price per hour in USD (required)")
	cmd.Flags().StringVar(&duration, "duration", "1h", "job duration")
	cmd.Flags().StringVar(&deadline, "deadline", "", "deadline to find match (e.g., 2h, 30m)")
	cmd.Flags().StringVar(&location, "location", "", "preferred location/region")

	cmd.MarkFlagRequired("cpu")
	cmd.MarkFlagRequired("memory")
	cmd.MarkFlagRequired("max-price")

	return cmd
}

// newMarketplacePricesCmd creates the prices command
func newMarketplacePricesCmd() *cobra.Command {
	var (
		resourceType string
		location     string
		period       string
	)

	cmd := &cobra.Command{
		Use:   "prices",
		Short: "Show current market prices",
		Long:  "Display average prices and trends for different resource types",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			prices, err := apiClient.GetMarketPrices(client.MarketPricesOptions{
				ResourceType: resourceType,
				Location:     location,
				Period:       period,
			})
			if err != nil {
				return fmt.Errorf("failed to get prices: %w", err)
			}

			fmt.Printf("Market Prices (%s)\n", period)
			fmt.Printf("================\n\n")

			// CPU prices
			fmt.Printf("CPU (per core/hour):\n")
			fmt.Printf("  Average:  $%.3f\n", prices.CPU.Average)
			fmt.Printf("  Minimum:  $%.3f\n", prices.CPU.Min)
			fmt.Printf("  Maximum:  $%.3f\n", prices.CPU.Max)
			fmt.Printf("  Trend:    %+.1f%%\n\n", prices.CPU.TrendPercent)

			// GPU prices
			fmt.Printf("GPU (per GPU/hour):\n")
			for model, price := range prices.GPU {
				fmt.Printf("  %s:\n", model)
				fmt.Printf("    Average:  $%.2f\n", price.Average)
				fmt.Printf("    Minimum:  $%.2f\n", price.Min)
				fmt.Printf("    Maximum:  $%.2f\n", price.Max)
				fmt.Printf("    Trend:    %+.1f%%\n", price.TrendPercent)
			}

			// Memory prices
			fmt.Printf("\nMemory (per GB/hour):\n")
			fmt.Printf("  Average:  $%.4f\n", prices.Memory.Average)
			fmt.Printf("  Trend:    %+.1f%%\n", prices.Memory.TrendPercent)

			// Storage prices
			fmt.Printf("\nStorage (per GB/hour):\n")
			fmt.Printf("  Average:  $%.5f\n", prices.Storage.Average)
			fmt.Printf("  Trend:    %+.1f%%\n", prices.Storage.TrendPercent)

			if location != "" {
				fmt.Printf("\nPrices shown for location: %s\n", location)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&resourceType, "type", "all", "resource type (all, cpu, gpu)")
	cmd.Flags().StringVar(&location, "location", "", "filter by location")
	cmd.Flags().StringVar(&period, "period", "24h", "time period (1h, 24h, 7d, 30d)")

	return cmd
} 