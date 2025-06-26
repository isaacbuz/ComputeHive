package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"
	"github.com/computehive/cli/pkg/client"
	"github.com/computehive/cli/pkg/config"
	"github.com/computehive/cli/pkg/utils"
)

// NewJobCmd creates the job command
func NewJobCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "job",
		Short: "Manage compute jobs",
		Long:  "Submit, monitor, and manage compute jobs on the ComputeHive platform",
	}

	cmd.AddCommand(
		newJobSubmitCmd(),
		newJobListCmd(),
		newJobGetCmd(),
		newJobLogsCmd(),
		newJobCancelCmd(),
		newJobStatusCmd(),
		newJobResultsCmd(),
	)

	return cmd
}

// newJobSubmitCmd creates the job submit command
func newJobSubmitCmd() *cobra.Command {
	var (
		name         string
		dockerImage  string
		scriptFile   string
		command      []string
		cpuCores     int
		memoryGB     int
		gpuCount     int
		gpuModel     string
		storageGB    int
		maxRuntime   string
		priority     string
		environment  []string
		volumes      []string
		wait         bool
		output       string
	)

	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit a new job",
		Long:  "Submit a new compute job to the ComputeHive platform",
		Example: `  # Submit a Docker job
  computehive job submit --docker ubuntu:latest --command "echo hello"

  # Submit a script job
  computehive job submit --script ./train.py --cpu 8 --memory 32 --gpu 2

  # Submit with environment variables
  computehive job submit --docker pytorch/pytorch:latest \
    --env MODEL=resnet50 --env EPOCHS=100 \
    --command "python train.py"

  # Submit and wait for completion
  computehive job submit --docker myapp:latest --wait --output ./results`,
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			// Validate input
			if dockerImage == "" && scriptFile == "" {
				return fmt.Errorf("either --docker or --script must be specified")
			}

			// Parse environment variables
			envMap := make(map[string]string)
			for _, env := range environment {
				parts := strings.SplitN(env, "=", 2)
				if len(parts) != 2 {
					return fmt.Errorf("invalid environment variable format: %s", env)
				}
				envMap[parts[0]] = parts[1]
			}

			// Parse volumes
			volumeMounts := make([]client.VolumeMount, 0)
			for _, vol := range volumes {
				parts := strings.Split(vol, ":")
				if len(parts) != 2 {
					return fmt.Errorf("invalid volume format: %s", vol)
				}
				volumeMounts = append(volumeMounts, client.VolumeMount{
					HostPath:      parts[0],
					ContainerPath: parts[1],
				})
			}

			// Create job specification
			jobSpec := client.JobSpec{
				Name:        name,
				Type:        "docker",
				DockerImage: dockerImage,
				Command:     command,
				Environment: envMap,
				Volumes:     volumeMounts,
				Resources: client.ResourceRequirements{
					CPUCores:  cpuCores,
					MemoryGB:  memoryGB,
					GPUCount:  gpuCount,
					GPUModel:  gpuModel,
					StorageGB: storageGB,
				},
				MaxRuntime: maxRuntime,
				Priority:   priority,
			}

			// Handle script submission
			if scriptFile != "" {
				jobSpec.Type = "script"
				scriptData, err := os.ReadFile(scriptFile)
				if err != nil {
					return fmt.Errorf("failed to read script file: %w", err)
				}
				jobSpec.Script = string(scriptData)
				jobSpec.ScriptName = filepath.Base(scriptFile)
			}

			fmt.Println("Submitting job...")
			job, err := apiClient.SubmitJob(jobSpec)
			if err != nil {
				return fmt.Errorf("failed to submit job: %w", err)
			}

			fmt.Printf("✅ Job submitted successfully!\n")
			fmt.Printf("Job ID: %s\n", job.ID)
			fmt.Printf("Status: %s\n", job.Status)

			if wait {
				fmt.Println("\nWaiting for job completion...")
				finalJob, err := apiClient.WaitForJob(job.ID, 0)
				if err != nil {
					return fmt.Errorf("failed to wait for job: %w", err)
				}

				fmt.Printf("\nJob completed with status: %s\n", finalJob.Status)
				if finalJob.Status == "completed" && output != "" {
					fmt.Printf("Downloading results to %s...\n", output)
					if err := apiClient.DownloadJobResults(job.ID, output); err != nil {
						return fmt.Errorf("failed to download results: %w", err)
					}
					fmt.Println("✅ Results downloaded successfully")
				}
			} else {
				fmt.Printf("\nTo check job status, run:\n")
				fmt.Printf("  computehive job status %s\n", job.ID)
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&name, "name", "", "job name")
	cmd.Flags().StringVar(&dockerImage, "docker", "", "Docker image to run")
	cmd.Flags().StringVar(&scriptFile, "script", "", "script file to run")
	cmd.Flags().StringArrayVar(&command, "command", []string{}, "command to execute")
	cmd.Flags().IntVar(&cpuCores, "cpu", 1, "number of CPU cores")
	cmd.Flags().IntVar(&memoryGB, "memory", 4, "memory in GB")
	cmd.Flags().IntVar(&gpuCount, "gpu", 0, "number of GPUs")
	cmd.Flags().StringVar(&gpuModel, "gpu-model", "", "specific GPU model (e.g., nvidia-a100)")
	cmd.Flags().IntVar(&storageGB, "storage", 10, "storage in GB")
	cmd.Flags().StringVar(&maxRuntime, "max-runtime", "1h", "maximum runtime (e.g., 30m, 2h, 1d)")
	cmd.Flags().StringVar(&priority, "priority", "normal", "job priority (low, normal, high)")
	cmd.Flags().StringArrayVar(&environment, "env", []string{}, "environment variables (KEY=VALUE)")
	cmd.Flags().StringArrayVar(&volumes, "volume", []string{}, "volume mounts (host:container)")
	cmd.Flags().BoolVar(&wait, "wait", false, "wait for job completion")
	cmd.Flags().StringVar(&output, "output", "", "download results to this directory (requires --wait)")

	return cmd
}

// newJobListCmd creates the job list command
func newJobListCmd() *cobra.Command {
	var (
		status string
		limit  int
		since  string
		user   string
		all    bool
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List jobs",
		Long:  "List compute jobs with optional filters",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			// Parse since time
			var sinceTime *time.Time
			if since != "" {
				duration, err := time.ParseDuration(since)
				if err != nil {
					return fmt.Errorf("invalid duration format: %w", err)
				}
				t := time.Now().Add(-duration)
				sinceTime = &t
			}

			jobs, err := apiClient.ListJobs(client.ListJobsOptions{
				Status: status,
				Limit:  limit,
				Since:  sinceTime,
				UserID: user,
				All:    all,
			})
			if err != nil {
				return fmt.Errorf("failed to list jobs: %w", err)
			}

			if len(jobs) == 0 {
				fmt.Println("No jobs found")
				return nil
			}

			// Print table
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "ID\tNAME\tSTATUS\tTYPE\tCPU\tMEMORY\tGPU\tCREATED\tRUNTIME")
			fmt.Fprintln(w, "--\t----\t------\t----\t---\t------\t---\t-------\t-------")
			
			for _, job := range jobs {
				runtime := "-"
				if job.StartedAt != nil && job.CompletedAt != nil {
					runtime = job.CompletedAt.Sub(*job.StartedAt).String()
				} else if job.StartedAt != nil {
					runtime = time.Since(*job.StartedAt).String()
				}

				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%d\t%dGB\t%d\t%s\t%s\n",
					job.ID[:8],
					utils.Truncate(job.Name, 20),
					job.Status,
					job.Type,
					job.Resources.CPUCores,
					job.Resources.MemoryGB,
					job.Resources.GPUCount,
					job.CreatedAt.Format("2006-01-02 15:04"),
					runtime,
				)
			}
			w.Flush()

			return nil
		},
	}

	cmd.Flags().StringVar(&status, "status", "", "filter by status (pending, running, completed, failed)")
	cmd.Flags().IntVar(&limit, "limit", 50, "maximum number of jobs to list")
	cmd.Flags().StringVar(&since, "since", "", "show jobs created since (e.g., 2h, 7d)")
	cmd.Flags().StringVar(&user, "user", "", "filter by user ID")
	cmd.Flags().BoolVar(&all, "all", false, "show all jobs (including archived)")

	return cmd
}

// newJobGetCmd creates the job get command
func newJobGetCmd() *cobra.Command {
	var format string

	cmd := &cobra.Command{
		Use:   "get [job-id]",
		Short: "Get job details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			jobID := args[0]

			job, err := apiClient.GetJob(jobID)
			if err != nil {
				return fmt.Errorf("failed to get job: %w", err)
			}

			// Format output
			switch format {
			case "json":
				return utils.PrintJSON(job)
			case "yaml":
				return utils.PrintYAML(job)
			default:
				// Print detailed info
				fmt.Printf("Job Details\n")
				fmt.Printf("===========\n")
				fmt.Printf("ID:           %s\n", job.ID)
				fmt.Printf("Name:         %s\n", job.Name)
				fmt.Printf("Status:       %s\n", job.Status)
				fmt.Printf("Type:         %s\n", job.Type)
				fmt.Printf("Priority:     %s\n", job.Priority)
				fmt.Printf("Created:      %s\n", job.CreatedAt.Format(time.RFC3339))
				
				if job.StartedAt != nil {
					fmt.Printf("Started:      %s\n", job.StartedAt.Format(time.RFC3339))
				}
				if job.CompletedAt != nil {
					fmt.Printf("Completed:    %s\n", job.CompletedAt.Format(time.RFC3339))
				}
				
				fmt.Printf("\nResources\n")
				fmt.Printf("---------\n")
				fmt.Printf("CPU Cores:    %d\n", job.Resources.CPUCores)
				fmt.Printf("Memory:       %d GB\n", job.Resources.MemoryGB)
				fmt.Printf("GPU Count:    %d\n", job.Resources.GPUCount)
				if job.Resources.GPUModel != "" {
					fmt.Printf("GPU Model:    %s\n", job.Resources.GPUModel)
				}
				fmt.Printf("Storage:      %d GB\n", job.Resources.StorageGB)
				
				if job.Type == "docker" && job.DockerImage != "" {
					fmt.Printf("\nDocker\n")
					fmt.Printf("------\n")
					fmt.Printf("Image:        %s\n", job.DockerImage)
					if len(job.Command) > 0 {
						fmt.Printf("Command:      %s\n", strings.Join(job.Command, " "))
					}
				}
				
				if job.AssignedAgentID != "" {
					fmt.Printf("\nExecution\n")
					fmt.Printf("---------\n")
					fmt.Printf("Agent ID:     %s\n", job.AssignedAgentID)
					fmt.Printf("Exit Code:    %d\n", job.ExitCode)
				}
				
				if job.Error != "" {
					fmt.Printf("\nError\n")
					fmt.Printf("-----\n")
					fmt.Printf("%s\n", job.Error)
				}
				
				fmt.Printf("\nCost\n")
				fmt.Printf("----\n")
				fmt.Printf("Estimated:    $%.4f\n", job.EstimatedCost)
				fmt.Printf("Actual:       $%.4f\n", job.ActualCost)
			}

			return nil
		},
	}

	cmd.Flags().StringVarP(&format, "output", "o", "", "output format (json, yaml)")

	return cmd
}

// newJobLogsCmd creates the job logs command
func newJobLogsCmd() *cobra.Command {
	var (
		follow bool
		tail   int
		since  string
	)

	cmd := &cobra.Command{
		Use:   "logs [job-id]",
		Short: "View job logs",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			jobID := args[0]

			logs, err := apiClient.GetJobLogs(jobID, client.LogOptions{
				Follow: follow,
				Tail:   tail,
				Since:  since,
			})
			if err != nil {
				return fmt.Errorf("failed to get logs: %w", err)
			}

			for log := range logs {
				fmt.Println(log.Line)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&follow, "follow", "f", false, "follow log output")
	cmd.Flags().IntVar(&tail, "tail", 100, "number of lines to show from the end")
	cmd.Flags().StringVar(&since, "since", "", "show logs since timestamp (e.g., 2h, 30m)")

	return cmd
}

// newJobCancelCmd creates the job cancel command
func newJobCancelCmd() *cobra.Command {
	var force bool

	cmd := &cobra.Command{
		Use:   "cancel [job-id]",
		Short: "Cancel a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			jobID := args[0]

			fmt.Printf("Cancelling job %s...\n", jobID)
			if err := apiClient.CancelJob(jobID, force); err != nil {
				return fmt.Errorf("failed to cancel job: %w", err)
			}

			fmt.Println("✅ Job cancelled successfully")
			return nil
		},
	}

	cmd.Flags().BoolVar(&force, "force", false, "force termination")

	return cmd
}

// newJobStatusCmd creates the job status command
func newJobStatusCmd() *cobra.Command {
	var watch bool

	cmd := &cobra.Command{
		Use:   "status [job-id]",
		Short: "Check job status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			jobID := args[0]

			if watch {
				// Watch for status changes
				return watchJobStatus(apiClient, jobID)
			}

			job, err := apiClient.GetJob(jobID)
			if err != nil {
				return fmt.Errorf("failed to get job: %w", err)
			}

			fmt.Printf("Job %s: %s\n", job.ID[:8], job.Status)
			
			if job.Status == "running" && job.Progress > 0 {
				fmt.Printf("Progress: %.1f%%\n", job.Progress)
			}
			
			if job.Status == "completed" {
				fmt.Printf("Exit Code: %d\n", job.ExitCode)
				fmt.Printf("Runtime: %s\n", job.CompletedAt.Sub(*job.StartedAt))
			}
			
			if job.Status == "failed" && job.Error != "" {
				fmt.Printf("Error: %s\n", job.Error)
			}

			return nil
		},
	}

	cmd.Flags().BoolVarP(&watch, "watch", "w", false, "watch for status changes")

	return cmd
}

// newJobResultsCmd creates the job results command
func newJobResultsCmd() *cobra.Command {
	var (
		output string
		list   bool
	)

	cmd := &cobra.Command{
		Use:   "results [job-id]",
		Short: "Download job results",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)
			jobID := args[0]

			if list {
				// List available results
				files, err := apiClient.ListJobResults(jobID)
				if err != nil {
					return fmt.Errorf("failed to list results: %w", err)
				}

				if len(files) == 0 {
					fmt.Println("No results available")
					return nil
				}

				fmt.Println("Available results:")
				for _, file := range files {
					fmt.Printf("  %s (%.2f MB)\n", file.Name, float64(file.Size)/1024/1024)
				}
				return nil
			}

			// Download results
			if output == "" {
				output = fmt.Sprintf("job-%s-results", jobID[:8])
			}

			fmt.Printf("Downloading results to %s...\n", output)
			if err := apiClient.DownloadJobResults(jobID, output); err != nil {
				return fmt.Errorf("failed to download results: %w", err)
			}

			fmt.Println("✅ Results downloaded successfully")
			return nil
		},
	}

	cmd.Flags().StringVarP(&output, "output", "o", "", "output directory")
	cmd.Flags().BoolVar(&list, "list", false, "list available results without downloading")

	return cmd
}

// Helper functions

func watchJobStatus(client *client.Client, jobID string) error {
	lastStatus := ""
	for {
		job, err := client.GetJob(jobID)
		if err != nil {
			return err
		}

		if job.Status != lastStatus {
			fmt.Printf("[%s] Status: %s", time.Now().Format("15:04:05"), job.Status)
			if job.Progress > 0 {
				fmt.Printf(" (%.1f%%)", job.Progress)
			}
			fmt.Println()
			lastStatus = job.Status
		}

		if job.Status == "completed" || job.Status == "failed" || job.Status == "cancelled" {
			return nil
		}

		time.Sleep(2 * time.Second)
	}
} 