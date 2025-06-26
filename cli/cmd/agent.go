package cmd

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/spf13/cobra"
	"github.com/computehive/cli/pkg/client"
	"github.com/computehive/cli/pkg/config"
)

// NewAgentCmd creates the agent command
func NewAgentCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "agent",
		Short: "Manage ComputeHive agents",
		Long:  "Deploy, configure, and monitor ComputeHive compute agents",
	}

	cmd.AddCommand(
		newAgentDeployCmd(),
		newAgentListCmd(),
		newAgentInfoCmd(),
		newAgentStartCmd(),
		newAgentStopCmd(),
		newAgentRestartCmd(),
		newAgentLogsCmd(),
		newAgentUpdateCmd(),
		newAgentUninstallCmd(),
	)

	return cmd
}

// newAgentDeployCmd creates the agent deploy command
func newAgentDeployCmd() *cobra.Command {
	var (
		name         string
		cpuCores     int
		memoryGB     int
		gpuCount     int
		storageGB    int
		location     string
		tags         []string
		autoStart    bool
	)

	cmd := &cobra.Command{
		Use:   "deploy",
		Short: "Deploy a new agent",
		Long:  "Deploy a new ComputeHive agent on the current machine",
		Example: `  # Deploy agent with default settings
  computehive agent deploy

  # Deploy agent with custom resources
  computehive agent deploy --name gpu-node-1 --cpu 32 --memory 128 --gpu 4

  # Deploy agent with tags
  computehive agent deploy --tags gpu,high-memory,us-east`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			// Detect system resources if not specified
			if cpuCores == 0 || memoryGB == 0 {
				fmt.Println("Detecting system resources...")
				// In real implementation, would detect actual resources
				if cpuCores == 0 {
					cpuCores = 4 // Default
				}
				if memoryGB == 0 {
					memoryGB = 8 // Default
				}
			}

			fmt.Printf("Deploying agent with:\n")
			fmt.Printf("  Name: %s\n", name)
			fmt.Printf("  CPU Cores: %d\n", cpuCores)
			fmt.Printf("  Memory: %d GB\n", memoryGB)
			fmt.Printf("  GPU Count: %d\n", gpuCount)
			fmt.Printf("  Storage: %d GB\n", storageGB)
			fmt.Printf("  Location: %s\n", location)
			fmt.Printf("  Tags: %v\n", tags)

			// Deploy agent
			agent, err := apiClient.DeployAgent(client.AgentConfig{
				Name:      name,
				CPUCores:  cpuCores,
				MemoryGB:  memoryGB,
				GPUCount:  gpuCount,
				StorageGB: storageGB,
				Location:  location,
				Tags:      tags,
			})
			if err != nil {
				return fmt.Errorf("failed to deploy agent: %w", err)
			}

			fmt.Printf("\n✅ Agent deployed successfully!\n")
			fmt.Printf("Agent ID: %s\n", agent.ID)
			fmt.Printf("Status: %s\n", agent.Status)

			if autoStart {
				fmt.Println("\nStarting agent...")
				if err := apiClient.StartAgent(agent.ID); err != nil {
					return fmt.Errorf("failed to start agent: %w", err)
				}
				fmt.Println("✅ Agent started")
			}

			fmt.Printf("\nTo start the agent manually, run:\n")
			fmt.Printf("  computehive agent start %s\n", agent.ID)

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "agent name (auto-generated if not specified)")
	cmd.Flags().IntVar(&cpuCores, "cpu", 0, "number of CPU cores (auto-detected if not specified)")
	cmd.Flags().IntVar(&memoryGB, "memory", 0, "memory in GB (auto-detected if not specified)")
	cmd.Flags().IntVar(&gpuCount, "gpu", 0, "number of GPUs")
	cmd.Flags().IntVar(&storageGB, "storage", 100, "storage in GB")
	cmd.Flags().StringVar(&location, "location", "", "agent location (e.g., us-east-1)")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "agent tags")
	cmd.Flags().BoolVar(&autoStart, "start", true, "automatically start agent after deployment")

	return cmd
}

// newAgentListCmd creates the agent list command
func newAgentListCmd() *cobra.Command {
	var (
		status   string
		location string
		tags     []string
		limit    int
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List agents",
		Long:  "List all ComputeHive agents associated with your account",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			agents, err := apiClient.ListAgents(client.ListAgentsOptions{
				Status:   status,
				Location: location,
				Tags:     tags,
				Limit:    limit,
			})
			if err != nil {
				return fmt.Errorf("failed to list agents: %w", err)
			}

			if len(agents) == 0 {
				fmt.Println("No agents found")
				return nil
			}

			// Print table
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tSTATUS\tCPU\tMEMORY\tGPU\tLOCATION\tUPTIME")
			fmt.Fprintln(w, "--\t----\t------\t---\t------\t---\t--------\t------")
			
			for _, agent := range agents {
				fmt.Fprintf(w, "%s\t%s\t%s\t%d\t%dGB\t%d\t%s\t%s\n",
					agent.ID[:8],
					agent.Name,
					agent.Status,
					agent.CPUCores,
					agent.MemoryGB,
					agent.GPUCount,
					agent.Location,
					agent.Uptime,
				)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "filter by status (online, offline, busy)")
	cmd.Flags().StringVar(&location, "location", "", "filter by location")
	cmd.Flags().StringSliceVar(&tags, "tags", []string{}, "filter by tags")
	cmd.Flags().IntVar(&limit, "limit", 50, "maximum number of agents to list")

	return cmd
}

// newAgentInfoCmd creates the agent info command
func newAgentInfoCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "info [agent-id]",
		Short: "Show detailed agent information",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			agentID := args[0]

			agent, err := apiClient.GetAgent(agentID)
			if err != nil {
				return fmt.Errorf("failed to get agent: %w", err)
			}

			// Print detailed info
			fmt.Printf("Agent Information\n")
			fmt.Printf("================\n")
			fmt.Printf("ID:           %s\n", agent.ID)
			fmt.Printf("Name:         %s\n", agent.Name)
			fmt.Printf("Status:       %s\n", agent.Status)
			fmt.Printf("Version:      %s\n", agent.Version)
			fmt.Printf("Location:     %s\n", agent.Location)
			fmt.Printf("Tags:         %v\n", agent.Tags)
			fmt.Printf("Created:      %s\n", agent.CreatedAt)
			fmt.Printf("Last Seen:    %s\n", agent.LastSeen)
			fmt.Printf("Uptime:       %s\n", agent.Uptime)
			fmt.Printf("\nResources\n")
			fmt.Printf("---------\n")
			fmt.Printf("CPU Cores:    %d (%.1f%% used)\n", agent.CPUCores, agent.CPUUsage)
			fmt.Printf("Memory:       %d GB (%.1f%% used)\n", agent.MemoryGB, agent.MemoryUsage)
			fmt.Printf("GPU Count:    %d\n", agent.GPUCount)
			fmt.Printf("Storage:      %d GB (%.1f%% used)\n", agent.StorageGB, agent.StorageUsage)
			fmt.Printf("Network:      %.1f Mbps\n", agent.NetworkBandwidth)
			fmt.Printf("\nJobs\n")
			fmt.Printf("----\n")
			fmt.Printf("Running:      %d\n", agent.RunningJobs)
			fmt.Printf("Completed:    %d\n", agent.CompletedJobs)
			fmt.Printf("Failed:       %d\n", agent.FailedJobs)
			fmt.Printf("\nEarnings\n")
			fmt.Printf("--------\n")
			fmt.Printf("Today:        $%.2f\n", agent.EarningsToday)
			fmt.Printf("This Month:   $%.2f\n", agent.EarningsMonth)
			fmt.Printf("Total:        $%.2f\n", agent.EarningsTotal)

			return nil
		},
	}

	return cmd
}

// newAgentStartCmd creates the agent start command
func newAgentStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start [agent-id]",
		Short: "Start an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			agentID := args[0]

			fmt.Printf("Starting agent %s...\n", agentID)
			if err := apiClient.StartAgent(agentID); err != nil {
				return fmt.Errorf("failed to start agent: %w", err)
			}

			fmt.Println("✅ Agent started successfully")
			return nil
		},
	}

	return cmd
}

// newAgentStopCmd creates the agent stop command
func newAgentStopCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "stop [agent-id]",
		Short: "Stop an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			agentID := args[0]

			fmt.Printf("Stopping agent %s...\n", agentID)
			if err := apiClient.StopAgent(agentID, force); err != nil {
				return fmt.Errorf("failed to stop agent: %w", err)
			}

			fmt.Println("✅ Agent stopped successfully")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "force stop (terminate running jobs)")

	return cmd
}

// newAgentRestartCmd creates the agent restart command
func newAgentRestartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restart [agent-id]",
		Short: "Restart an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			agentID := args[0]

			fmt.Printf("Restarting agent %s...\n", agentID)
			if err := apiClient.RestartAgent(agentID); err != nil {
				return fmt.Errorf("failed to restart agent: %w", err)
			}

			fmt.Println("✅ Agent restarted successfully")
			return nil
		},
	}

	return cmd
}

// newAgentLogsCmd creates the agent logs command
func newAgentLogsCmd() *cobra.Command {
	var (
		follow bool
		tail   int
		since  string
	)

	cmd := &cobra.Command{
		Use:   "logs [agent-id]",
		Short: "View agent logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			agentID := args[0]

			logs, err := apiClient.GetAgentLogs(agentID, client.LogOptions{
				Follow: follow,
				Tail:   tail,
				Since:  since,
			})
			if err != nil {
				return fmt.Errorf("failed to get logs: %w", err)
			}

			for log := range logs {
				fmt.Printf("[%s] %s: %s\n", log.Timestamp, log.Level, log.Message)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
	cmd.Flags().IntVar(&tail, "tail", 100, "number of lines to show from the end")
	cmd.Flags().StringVar(&since, "since", "", "show logs since timestamp (e.g., 2h, 30m)")

	return cmd
}

// newAgentUpdateCmd creates the agent update command
func newAgentUpdateCmd() *cobra.Command {
	var (
		version   string
		autoApply bool
	)

	cmd := &cobra.Command{
		Use:   "update [agent-id]",
		Short: "Update agent software",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			agentID := args[0]

			// Check for updates
			fmt.Printf("Checking for updates for agent %s...\n", agentID)
			update, err := apiClient.CheckAgentUpdate(agentID, version)
			if err != nil {
				return fmt.Errorf("failed to check updates: %w", err)
			}

			if !update.Available {
				fmt.Println("Agent is already up to date")
				return nil
			}

			fmt.Printf("Update available: %s -> %s\n", update.CurrentVersion, update.NewVersion)
			fmt.Printf("Changes:\n%s\n", update.Changelog)

			if !autoApply {
				fmt.Print("\nApply update? [y/N] ")
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Update cancelled")
					return nil
				}
			}

			fmt.Println("Applying update...")
			if err := apiClient.UpdateAgent(agentID, update.NewVersion); err != nil {
				return fmt.Errorf("failed to update agent: %w", err)
			}

			fmt.Println("✅ Agent updated successfully")
			return nil
		},
	}

	cmd.Flags().StringVar(&version, "version", "", "specific version to update to")
	cmd.Flags().BoolVar(&autoApply, "yes", false, "automatically apply updates without confirmation")

	return cmd
}

// newAgentUninstallCmd creates the agent uninstall command
func newAgentUninstallCmd() *cobra.Command {
	var (
		force       bool
		keepData    bool
	)

	cmd := &cobra.Command{
		Use:   "uninstall [agent-id]",
		Short: "Uninstall an agent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			agentID := args[0]

			if !force {
				fmt.Printf("This will uninstall agent %s. Are you sure? [y/N] ", agentID)
				var response string
				fmt.Scanln(&response)
				if response != "y" && response != "Y" {
					fmt.Println("Uninstall cancelled")
					return nil
				}
			}

			fmt.Printf("Uninstalling agent %s...\n", agentID)
			if err := apiClient.UninstallAgent(agentID, keepData); err != nil {
				return fmt.Errorf("failed to uninstall agent: %w", err)
			}

			fmt.Println("✅ Agent uninstalled successfully")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "skip confirmation prompt")
	cmd.Flags().BoolVar(&keepData, "keep-data", false, "keep agent data after uninstall")

	return cmd
} 