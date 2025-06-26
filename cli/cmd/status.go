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

// NewStatusCmd creates the status command
func NewStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "status",
		Short: "Show system status",
		Long:  "Display the status of ComputeHive services, your agents, and overall system health",
		RunE:  runStatus,
	}

	cmd.AddCommand(
		newStatusServicesCmd(),
		newStatusAgentsCmd(),
		newStatusJobsCmd(),
		newStatusAccountCmd(),
	)

	return cmd
}

// runStatus shows overall system status
func runStatus(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	if cfg.Token == "" {
		fmt.Println("Not logged in. Please login first:")
		fmt.Println("  computehive auth login")
		return nil
	}

	apiClient := client.New(cfg.APIURL, cfg.Token)

	// Get system status
	status, err := apiClient.GetSystemStatus()
	if err != nil {
		return fmt.Errorf("failed to get system status: %w", err)
	}

	fmt.Println("ComputeHive System Status")
	fmt.Println("========================")
	fmt.Printf("Status:         %s\n", colorizeStatus(status.Overall))
	fmt.Printf("API Version:    %s\n", status.APIVersion)
	fmt.Printf("Last Updated:   %s\n", status.LastUpdated.Format("2006-01-02 15:04:05"))
	
	// Service summary
	fmt.Printf("\nServices:       ")
	healthyCount := 0
	for _, svc := range status.Services {
		if svc.Status == "healthy" {
			healthyCount++
		}
	}
	fmt.Printf("%d/%d healthy\n", healthyCount, len(status.Services))

	// Quick stats
	fmt.Printf("\nQuick Stats:\n")
	fmt.Printf("  Active Agents:    %d\n", status.Stats.ActiveAgents)
	fmt.Printf("  Running Jobs:     %d\n", status.Stats.RunningJobs)
	fmt.Printf("  Available GPUs:   %d\n", status.Stats.AvailableGPUs)
	fmt.Printf("  Total Capacity:   %.2f TFLOPS\n", status.Stats.TotalCapacity)

	// Recent incidents
	if len(status.RecentIncidents) > 0 {
		fmt.Printf("\nRecent Incidents:\n")
		for _, incident := range status.RecentIncidents {
			fmt.Printf("  - [%s] %s: %s\n", 
				incident.Time.Format("15:04"),
				incident.Service,
				incident.Message,
			)
		}
	}

	// Maintenance windows
	if len(status.ScheduledMaintenance) > 0 {
		fmt.Printf("\nScheduled Maintenance:\n")
		for _, maint := range status.ScheduledMaintenance {
			fmt.Printf("  - %s: %s (%s)\n",
				maint.StartTime.Format("Jan 02 15:04"),
				maint.Description,
				maint.Duration,
			)
		}
	}

	fmt.Println("\nFor detailed status, use:")
	fmt.Println("  computehive status services  # Service health")
	fmt.Println("  computehive status agents    # Your agents")
	fmt.Println("  computehive status jobs      # Job statistics")

	return nil
}

// newStatusServicesCmd creates the services subcommand
func newStatusServicesCmd() *cobra.Command {
	var watch bool

	cmd := &cobra.Command{
		Use:   "services",
		Short: "Show service health status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			if watch {
				return watchServiceStatus(apiClient)
			}

			services, err := apiClient.GetServiceHealth()
			if err != nil {
				return fmt.Errorf("failed to get service health: %w", err)
			}

			// Print table
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "SERVICE\tSTATUS\tRESPONSE TIME\tUPTIME\tVERSION")
			fmt.Fprintln(w, "-------\t------\t-------------\t------\t-------")
			
			for _, svc := range services {
				uptime := calculateUptime(svc.StartTime)
				fmt.Fprintf(w, "%s\t%s\t%dms\t%s\t%s\n",
					svc.Name,
					colorizeStatus(svc.Status),
					svc.ResponseTime,
					uptime,
					svc.Version,
				)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch service status in real-time")

	return cmd
}

// newStatusAgentsCmd creates the agents subcommand
func newStatusAgentsCmd() *cobra.Command {
	var (
		all bool
		id  string
	)

	cmd := &cobra.Command{
		Use:   "agents",
		Short: "Show agent status",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			// Show specific agent
			if id != "" {
				agent, err := apiClient.GetAgent(id)
				if err != nil {
					return fmt.Errorf("failed to get agent: %w", err)
				}

				fmt.Printf("Agent Details\n")
				fmt.Printf("=============\n")
				fmt.Printf("ID:            %s\n", agent.ID)
				fmt.Printf("Name:          %s\n", agent.Name)
				fmt.Printf("Status:        %s\n", colorizeStatus(agent.Status))
				fmt.Printf("Version:       %s\n", agent.Version)
				fmt.Printf("Platform:      %s/%s\n", agent.OS, agent.Arch)
				fmt.Printf("Last Seen:     %s\n", agent.LastSeen.Format("2006-01-02 15:04:05"))
				
				fmt.Printf("\nResources:\n")
				fmt.Printf("  CPU:         %d cores (%.1f%% used)\n", agent.Resources.CPUCores, agent.Resources.CPUUsage)
				fmt.Printf("  Memory:      %d GB (%.1f%% used)\n", agent.Resources.MemoryGB, agent.Resources.MemoryUsage)
				fmt.Printf("  Storage:     %d GB (%.1f%% used)\n", agent.Resources.StorageGB, agent.Resources.StorageUsage)
				
				if agent.Resources.GPUCount > 0 {
					fmt.Printf("  GPUs:        %d x %s\n", agent.Resources.GPUCount, agent.Resources.GPUModel)
					for i, gpu := range agent.Resources.GPUs {
						fmt.Printf("    GPU %d:     %.1f%% used, %.1f°C\n", i, gpu.Usage, gpu.Temperature)
					}
				}
				
				if agent.CurrentJob != "" {
					fmt.Printf("\nCurrent Job:   %s\n", agent.CurrentJob)
				}
				
				fmt.Printf("\nStatistics:\n")
				fmt.Printf("  Jobs Completed:  %d\n", agent.Stats.JobsCompleted)
				fmt.Printf("  Success Rate:    %.1f%%\n", agent.Stats.SuccessRate)
				fmt.Printf("  Total Runtime:   %s\n", agent.Stats.TotalRuntime)
				
				return nil
			}

			// List all agents
			agents, err := apiClient.ListAgents(client.ListAgentsOptions{
				All: all,
			})
			if err != nil {
				return fmt.Errorf("failed to list agents: %w", err)
			}

			if len(agents) == 0 {
				fmt.Println("No agents found")
				fmt.Println("\nTo start an agent, run:")
				fmt.Println("  computehive agent start")
				return nil
			}

			// Print table
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tSTATUS\tCPU\tMEMORY\tGPU\tJOBS\tLAST SEEN")
			fmt.Fprintln(w, "--\t----\t------\t---\t------\t---\t----\t---------")
			
			for _, agent := range agents {
				lastSeen := "never"
				if !agent.LastSeen.IsZero() {
					if time.Since(agent.LastSeen) < time.Hour {
						lastSeen = fmt.Sprintf("%.0fm ago", time.Since(agent.LastSeen).Minutes())
					} else {
						lastSeen = agent.LastSeen.Format("Jan 02 15:04")
					}
				}
				
				gpuInfo := "-"
				if agent.Resources.GPUCount > 0 {
					gpuInfo = fmt.Sprintf("%d", agent.Resources.GPUCount)
				}
				
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%dGB\t%s\t%d\t%s\n",
					agent.ID[:8],
					utils.Truncate(agent.Name, 20),
					colorizeStatus(agent.Status),
					agent.Resources.CPUCores,
					agent.Resources.MemoryGB,
					gpuInfo,
					agent.Stats.JobsCompleted,
					lastSeen,
				)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().BoolVar(&all, "all", false, "show all agents including inactive")
	cmd.Flags().StringVar(&id, "id", "", "show details for specific agent")

	return cmd
}

// newStatusJobsCmd creates the jobs subcommand
func newStatusJobsCmd() *cobra.Command {
	var period string

	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Show job statistics",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			stats, err := apiClient.GetJobStatistics(period)
			if err != nil {
				return fmt.Errorf("failed to get job statistics: %w", err)
			}

			fmt.Printf("Job Statistics (%s)\n", period)
			fmt.Printf("==================\n\n")

			fmt.Printf("Summary:\n")
			fmt.Printf("  Total Jobs:       %d\n", stats.TotalJobs)
			fmt.Printf("  Completed:        %d (%.1f%%)\n", stats.Completed, stats.CompletionRate)
			fmt.Printf("  Failed:           %d (%.1f%%)\n", stats.Failed, stats.FailureRate)
			fmt.Printf("  Running:          %d\n", stats.Running)
			fmt.Printf("  Pending:          %d\n", stats.Pending)
			
			fmt.Printf("\nPerformance:\n")
			fmt.Printf("  Avg Queue Time:   %s\n", stats.AvgQueueTime)
			fmt.Printf("  Avg Runtime:      %s\n", stats.AvgRuntime)
			fmt.Printf("  Success Rate:     %.1f%%\n", stats.SuccessRate)
			
			fmt.Printf("\nResource Usage:\n")
			fmt.Printf("  Total CPU Hours:  %.1f\n", stats.TotalCPUHours)
			fmt.Printf("  Total GPU Hours:  %.1f\n", stats.TotalGPUHours)
			fmt.Printf("  Total Cost:       $%.2f\n", stats.TotalCost)
			fmt.Printf("  Avg Cost/Job:     $%.2f\n", stats.AvgCostPerJob)
			
			if len(stats.TopErrors) > 0 {
				fmt.Printf("\nTop Errors:\n")
				for i, err := range stats.TopErrors {
					fmt.Printf("  %d. %s (%d occurrences)\n", i+1, err.Error, err.Count)
				}
			}
			
			// Job type breakdown
			if len(stats.JobTypes) > 0 {
				fmt.Printf("\nJob Types:\n")
				for jobType, count := range stats.JobTypes {
					percentage := float64(count) / float64(stats.TotalJobs) * 100
					fmt.Printf("  %-15s %d (%.1f%%)\n", jobType+":", count, percentage)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&period, "period", "24h", "time period (1h, 24h, 7d, 30d)")

	return cmd
}

// newStatusAccountCmd creates the account subcommand
func newStatusAccountCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "account",
		Short: "Show account status and usage",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			account, err := apiClient.GetAccountStatus()
			if err != nil {
				return fmt.Errorf("failed to get account status: %w", err)
			}

			fmt.Println("Account Status")
			fmt.Println("==============")
			fmt.Printf("Account ID:      %s\n", account.ID)
			fmt.Printf("Email:           %s\n", account.Email)
			fmt.Printf("Plan:            %s\n", account.Plan)
			fmt.Printf("Status:          %s\n", colorizeStatus(account.Status))
			
			if account.Organization != "" {
				fmt.Printf("Organization:    %s\n", account.Organization)
			}
			
			fmt.Printf("\nBalance:\n")
			fmt.Printf("  Available:     $%.2f\n", account.Balance.Available)
			fmt.Printf("  Pending:       $%.2f\n", account.Balance.Pending)
			fmt.Printf("  Credit Limit:  $%.2f\n", account.Balance.CreditLimit)
			
			fmt.Printf("\nUsage (This Month):\n")
			fmt.Printf("  Compute:       $%.2f\n", account.Usage.Compute)
			fmt.Printf("  Storage:       $%.2f\n", account.Usage.Storage)
			fmt.Printf("  Network:       $%.2f\n", account.Usage.Network)
			fmt.Printf("  Total:         $%.2f\n", account.Usage.Total)
			
			fmt.Printf("\nQuotas:\n")
			fmt.Printf("  Max Agents:    %d/%d\n", account.Quotas.AgentsUsed, account.Quotas.MaxAgents)
			fmt.Printf("  Max Jobs:      %d/%d per day\n", account.Quotas.JobsToday, account.Quotas.MaxJobsPerDay)
			fmt.Printf("  Max GPUs:      %d/%d\n", account.Quotas.GPUsUsed, account.Quotas.MaxGPUs)
			
			if len(account.Warnings) > 0 {
				fmt.Printf("\nWarnings:\n")
				for _, warning := range account.Warnings {
					fmt.Printf("  ⚠️  %s\n", warning)
				}
			}

			return nil
		},
	}

	return cmd
}

// Helper functions

func colorizeStatus(status string) string {
	// In a real implementation, this would use color codes
	switch status {
	case "healthy", "active", "running", "completed", "online":
		return "✅ " + status
	case "degraded", "warning", "pending":
		return "⚠️  " + status
	case "unhealthy", "error", "failed", "offline":
		return "❌ " + status
	default:
		return status
	}
}

func calculateUptime(startTime time.Time) string {
	if startTime.IsZero() {
		return "unknown"
	}
	
	duration := time.Since(startTime)
	days := int(duration.Hours() / 24)
	hours := int(duration.Hours()) % 24
	
	if days > 0 {
		return fmt.Sprintf("%dd %dh", days, hours)
	}
	return fmt.Sprintf("%dh %dm", hours, int(duration.Minutes())%60)
}

func watchServiceStatus(client *client.Client) error {
	for {
		// Clear screen (in production, use a proper terminal library)
		fmt.Print("\033[H\033[2J")
		
		services, err := client.GetServiceHealth()
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			fmt.Printf("Service Health Monitor - %s\n", time.Now().Format("15:04:05"))
			fmt.Println("=====================================")
			
			for _, svc := range services {
				fmt.Printf("%-20s %s (%.0fms)\n", 
					svc.Name,
					colorizeStatus(svc.Status),
					svc.ResponseTime,
				)
			}
		}
		
		fmt.Println("\nPress Ctrl+C to exit")
		time.Sleep(5 * time.Second)
	}
} 