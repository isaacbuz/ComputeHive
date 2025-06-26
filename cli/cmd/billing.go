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

// NewBillingCmd creates the billing command
func NewBillingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing",
		Short: "Manage billing and payments",
		Long:  "View usage, invoices, payment methods, and manage billing settings",
		Aliases: []string{"bill"},
	}

	cmd.AddCommand(
		newBillingUsageCmd(),
		newBillingInvoicesCmd(),
		newBillingPaymentMethodsCmd(),
		newBillingHistoryCmd(),
		newBillingAddFundsCmd(),
		newBillingAlertsCmd(),
	)

	return cmd
}

// newBillingUsageCmd creates the usage subcommand
func newBillingUsageCmd() *cobra.Command {
	var (
		period  string
		details bool
		format  string
	)

	cmd := &cobra.Command{
		Use:   "usage",
		Short: "Show resource usage and costs",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			usage, err := apiClient.GetUsage(client.UsageOptions{
				Period:  period,
				Details: details,
			})
			if err != nil {
				return fmt.Errorf("failed to get usage: %w", err)
			}

			// Format output
			switch format {
			case "json":
				return utils.PrintJSON(usage)
			case "csv":
				return printUsageCSV(usage)
			default:
				fmt.Printf("Usage Report (%s)\n", period)
				fmt.Printf("================\n\n")

				fmt.Printf("Summary:\n")
				fmt.Printf("  Total Cost:       $%.2f\n", usage.TotalCost)
				fmt.Printf("  Compute:          $%.2f (%.0f%%)\n", usage.ComputeCost, usage.ComputePercent)
				fmt.Printf("  Storage:          $%.2f (%.0f%%)\n", usage.StorageCost, usage.StoragePercent)
				fmt.Printf("  Network:          $%.2f (%.0f%%)\n", usage.NetworkCost, usage.NetworkPercent)
				fmt.Printf("  Other:            $%.2f\n", usage.OtherCost)

				fmt.Printf("\nResource Usage:\n")
				fmt.Printf("  CPU Hours:        %.1f\n", usage.CPUHours)
				fmt.Printf("  GPU Hours:        %.1f\n", usage.GPUHours)
				fmt.Printf("  Storage GB-Hours: %.1f\n", usage.StorageGBHours)
				fmt.Printf("  Network GB:       %.1f\n", usage.NetworkGB)

				if details && len(usage.DailyBreakdown) > 0 {
					fmt.Printf("\nDaily Breakdown:\n")
					w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
					fmt.Fprintln(w, "DATE\tCOMPUTE\tSTORAGE\tNETWORK\tTOTAL")
					fmt.Fprintln(w, "----\t-------\t-------\t-------\t-----")
					
					for _, day := range usage.DailyBreakdown {
						fmt.Fprintf(w, "%s\t$%.2f\t$%.2f\t$%.2f\t$%.2f\n",
							day.Date.Format("Jan 02"),
							day.Compute,
							day.Storage,
							day.Network,
							day.Total,
						)
					}
					w.Flush()
				}

				// Cost by job type
				if len(usage.JobTypeCosts) > 0 {
					fmt.Printf("\nCost by Job Type:\n")
					for jobType, cost := range usage.JobTypeCosts {
						fmt.Printf("  %-20s $%.2f\n", jobType+":", cost)
					}
				}

				// Projected costs
				fmt.Printf("\nProjected Costs:\n")
				fmt.Printf("  This Month:       $%.2f\n", usage.ProjectedMonthly)
				fmt.Printf("  Annual (at current rate): $%.2f\n", usage.ProjectedAnnual)

				// Cost trends
				if usage.TrendPercent != 0 {
					trend := "↑"
					if usage.TrendPercent < 0 {
						trend = "↓"
					}
					fmt.Printf("\nTrend: %s %.1f%% vs last period\n", trend, usage.TrendPercent)
				}
			}

			return nil
		},
	}

	cmd.Flags().StringVar(&period, "period", "current", "usage period (current, last-month, custom)")
	cmd.Flags().BoolVar(&details, "details", false, "show detailed breakdown")
	cmd.Flags().StringVar(&format, "format", "table", "output format (table, json, csv)")

	return cmd
}

// newBillingInvoicesCmd creates the invoices subcommand
func newBillingInvoicesCmd() *cobra.Command {
	var (
		limit    int
		status   string
		download string
	)

	cmd := &cobra.Command{
		Use:   "invoices",
		Short: "List and manage invoices",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			// Download specific invoice
			if download != "" {
				fmt.Printf("Downloading invoice %s...\n", download)
				filename := fmt.Sprintf("invoice-%s.pdf", download)
				if err := apiClient.DownloadInvoice(download, filename); err != nil {
					return fmt.Errorf("failed to download invoice: %w", err)
				}
				fmt.Printf("✅ Invoice saved to %s\n", filename)
				return nil
			}

			// List invoices
			invoices, err := apiClient.ListInvoices(client.ListInvoicesOptions{
				Limit:  limit,
				Status: status,
			})
			if err != nil {
				return fmt.Errorf("failed to list invoices: %w", err)
			}

			if len(invoices) == 0 {
				fmt.Println("No invoices found")
				return nil
			}

			// Print table
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "INVOICE #\tDATE\tAMOUNT\tSTATUS\tDUE DATE")
			fmt.Fprintln(w, "---------\t----\t------\t------\t--------")
			
			for _, invoice := range invoices {
				dueDate := "-"
				if invoice.Status == "pending" && invoice.DueDate != nil {
					dueDate = invoice.DueDate.Format("Jan 02")
				}
				
				fmt.Fprintf(w, "%s\t%s\t$%.2f\t%s\t%s\n",
					invoice.Number,
					invoice.Date.Format("2006-01-02"),
					invoice.Total,
					invoice.Status,
					dueDate,
				)
			}
			w.Flush()

			fmt.Println("\nTo download an invoice:")
			fmt.Println("  computehive billing invoices --download INVOICE_NUMBER")

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 10, "number of invoices to show")
	cmd.Flags().StringVar(&status, "status", "", "filter by status (paid, pending, overdue)")
	cmd.Flags().StringVar(&download, "download", "", "download invoice by number")

	return cmd
}

// newBillingPaymentMethodsCmd creates the payment-methods subcommand
func newBillingPaymentMethodsCmd() *cobra.Command {
	var (
		add      bool
		remove   string
		setDefault string
	)

	cmd := &cobra.Command{
		Use:   "payment-methods",
		Short: "Manage payment methods",
		Aliases: []string{"pm"},
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			// Add payment method
			if add {
				fmt.Println("Adding payment method...")
				fmt.Println("You will be redirected to our secure payment portal.")
				
				url, err := apiClient.GetPaymentMethodAddURL()
				if err != nil {
					return fmt.Errorf("failed to get payment URL: %w", err)
				}
				
				fmt.Printf("\nPlease visit: %s\n", url)
				return nil
			}

			// Remove payment method
			if remove != "" {
				fmt.Printf("Removing payment method %s...\n", remove)
				if err := apiClient.RemovePaymentMethod(remove); err != nil {
					return fmt.Errorf("failed to remove payment method: %w", err)
				}
				fmt.Println("✅ Payment method removed")
				return nil
			}

			// Set default payment method
			if setDefault != "" {
				fmt.Printf("Setting default payment method to %s...\n", setDefault)
				if err := apiClient.SetDefaultPaymentMethod(setDefault); err != nil {
					return fmt.Errorf("failed to set default: %w", err)
				}
				fmt.Println("✅ Default payment method updated")
				return nil
			}

			// List payment methods
			methods, err := apiClient.ListPaymentMethods()
			if err != nil {
				return fmt.Errorf("failed to list payment methods: %w", err)
			}

			if len(methods) == 0 {
				fmt.Println("No payment methods found")
				fmt.Println("\nTo add a payment method:")
				fmt.Println("  computehive billing payment-methods --add")
				return nil
			}

			fmt.Println("Payment Methods")
			fmt.Println("===============")
			for _, method := range methods {
				defaultStr := ""
				if method.IsDefault {
					defaultStr = " (default)"
				}
				
				fmt.Printf("\nID: %s%s\n", method.ID, defaultStr)
				
				switch method.Type {
				case "card":
					fmt.Printf("  Type: Credit Card\n")
					fmt.Printf("  Brand: %s\n", method.Card.Brand)
					fmt.Printf("  Last 4: %s\n", method.Card.Last4)
					fmt.Printf("  Expires: %02d/%04d\n", method.Card.ExpMonth, method.Card.ExpYear)
				case "bank_account":
					fmt.Printf("  Type: Bank Account\n")
					fmt.Printf("  Bank: %s\n", method.BankAccount.BankName)
					fmt.Printf("  Last 4: %s\n", method.BankAccount.Last4)
				case "crypto":
					fmt.Printf("  Type: Cryptocurrency\n")
					fmt.Printf("  Currency: %s\n", method.Crypto.Currency)
					fmt.Printf("  Address: %s...%s\n", 
						method.Crypto.Address[:8], 
						method.Crypto.Address[len(method.Crypto.Address)-8:],
					)
				}
			}

			fmt.Println("\nManage payment methods:")
			fmt.Println("  computehive billing payment-methods --add")
			fmt.Println("  computehive billing payment-methods --remove ID")
			fmt.Println("  computehive billing payment-methods --set-default ID")

			return nil
		},
	}

	cmd.Flags().BoolVar(&add, "add", false, "add a new payment method")
	cmd.Flags().StringVar(&remove, "remove", "", "remove payment method by ID")
	cmd.Flags().StringVar(&setDefault, "set-default", "", "set default payment method by ID")

	return cmd
}

// newBillingHistoryCmd creates the history subcommand
func newBillingHistoryCmd() *cobra.Command {
	var (
		limit  int
		filter string
		export string
	)

	cmd := &cobra.Command{
		Use:   "history",
		Short: "Show payment history",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			transactions, err := apiClient.GetPaymentHistory(client.PaymentHistoryOptions{
				Limit:  limit,
				Filter: filter,
			})
			if err != nil {
				return fmt.Errorf("failed to get payment history: %w", err)
			}

			if export != "" {
				return exportPaymentHistory(transactions, export)
			}

			if len(transactions) == 0 {
				fmt.Println("No payment history found")
				return nil
			}

			// Print table
			w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
			fmt.Fprintln(w, "DATE\tTYPE\tAMOUNT\tDESCRIPTION\tSTATUS")
			fmt.Fprintln(w, "----\t----\t------\t-----------\t------")
			
			for _, tx := range transactions {
				amount := fmt.Sprintf("$%.2f", tx.Amount)
				if tx.Type == "credit" {
					amount = "+" + amount
				} else {
					amount = "-" + amount
				}
				
				fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
					tx.Date.Format("2006-01-02"),
					tx.Type,
					amount,
					utils.Truncate(tx.Description, 40),
					tx.Status,
				)
			}
			w.Flush()

			// Show balance
			balance, _ := apiClient.GetBalance()
			fmt.Printf("\nCurrent Balance: $%.2f\n", balance.Available)

			return nil
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "number of transactions to show")
	cmd.Flags().StringVar(&filter, "filter", "", "filter transactions (payments, credits, refunds)")
	cmd.Flags().StringVar(&export, "export", "", "export to file (csv, json)")

	return cmd
}

// newBillingAddFundsCmd creates the add-funds subcommand
func newBillingAddFundsCmd() *cobra.Command {
	var (
		amount        float64
		paymentMethod string
	)

	cmd := &cobra.Command{
		Use:   "add-funds",
		Short: "Add funds to your account",
		Example: `  # Add $100 to account
  computehive billing add-funds --amount 100

  # Add funds with specific payment method
  computehive billing add-funds --amount 50 --payment-method pm_123`,
		RunE: func(cmd *cobra.Command, args []string) error {
			if amount <= 0 {
				return fmt.Errorf("amount must be greater than 0")
			}

			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			fmt.Printf("Adding $%.2f to your account...\n", amount)
			
			transaction, err := apiClient.AddFunds(client.AddFundsRequest{
				Amount:        amount,
				PaymentMethod: paymentMethod,
			})
			if err != nil {
				return fmt.Errorf("failed to add funds: %w", err)
			}

			fmt.Printf("✅ Successfully added $%.2f to your account\n", amount)
			fmt.Printf("Transaction ID: %s\n", transaction.ID)
			fmt.Printf("New Balance: $%.2f\n", transaction.NewBalance)

			return nil
		},
	}

	cmd.Flags().Float64Var(&amount, "amount", 0, "amount to add (in USD)")
	cmd.Flags().StringVar(&paymentMethod, "payment-method", "", "specific payment method ID")
	cmd.MarkFlagRequired("amount")

	return cmd
}

// newBillingAlertsCmd creates the alerts subcommand
func newBillingAlertsCmd() *cobra.Command {
	var (
		list      bool
		add       bool
		remove    string
		threshold float64
		alertType string
	)

	cmd := &cobra.Command{
		Use:   "alerts",
		Short: "Manage billing alerts",
		RunE: func(cmd *cobra.Command, args []string) error {
			cfg, err := config.Load()
			if err != nil {
				return fmt.Errorf("failed to load config: %w", err)
			}

			apiClient := client.New(cfg.APIURL, cfg.Token)

			// Add alert
			if add {
				if threshold <= 0 {
					return fmt.Errorf("threshold must be greater than 0")
				}

				alert, err := apiClient.CreateBillingAlert(client.BillingAlert{
					Type:      alertType,
					Threshold: threshold,
				})
				if err != nil {
					return fmt.Errorf("failed to create alert: %w", err)
				}

				fmt.Println("✅ Billing alert created")
				fmt.Printf("Alert ID: %s\n", alert.ID)
				fmt.Printf("You will be notified when %s exceeds $%.2f\n", alertType, threshold)
				return nil
			}

			// Remove alert
			if remove != "" {
				fmt.Printf("Removing alert %s...\n", remove)
				if err := apiClient.RemoveBillingAlert(remove); err != nil {
					return fmt.Errorf("failed to remove alert: %w", err)
				}
				fmt.Println("✅ Alert removed")
				return nil
			}

			// List alerts
			alerts, err := apiClient.ListBillingAlerts()
			if err != nil {
				return fmt.Errorf("failed to list alerts: %w", err)
			}

			if len(alerts) == 0 {
				fmt.Println("No billing alerts configured")
				fmt.Println("\nTo create an alert:")
				fmt.Println("  computehive billing alerts --add --type daily --threshold 100")
				return nil
			}

			fmt.Println("Billing Alerts")
			fmt.Println("==============")
			for _, alert := range alerts {
				status := "Active"
				if !alert.Enabled {
					status = "Disabled"
				}
				
				fmt.Printf("\nID: %s (%s)\n", alert.ID, status)
				fmt.Printf("  Type: %s spending\n", alert.Type)
				fmt.Printf("  Threshold: $%.2f\n", alert.Threshold)
				fmt.Printf("  Current: $%.2f\n", alert.CurrentValue)
				
				if alert.LastTriggered != nil {
					fmt.Printf("  Last Triggered: %s\n", alert.LastTriggered.Format("2006-01-02"))
				}
			}

			return nil
		},
	}

	cmd.Flags().BoolVar(&list, "list", true, "list all alerts")
	cmd.Flags().BoolVar(&add, "add", false, "add a new alert")
	cmd.Flags().StringVar(&remove, "remove", "", "remove alert by ID")
	cmd.Flags().Float64Var(&threshold, "threshold", 0, "spending threshold for alert")
	cmd.Flags().StringVar(&alertType, "type", "daily", "alert type (daily, weekly, monthly)")

	return cmd
}

// Helper functions

func printUsageCSV(usage *client.UsageReport) error {
	fmt.Println("Date,Compute,Storage,Network,Total")
	for _, day := range usage.DailyBreakdown {
		fmt.Printf("%s,%.2f,%.2f,%.2f,%.2f\n",
			day.Date.Format("2006-01-02"),
			day.Compute,
			day.Storage,
			day.Network,
			day.Total,
		)
	}
	return nil
}

func exportPaymentHistory(transactions []client.Transaction, format string) error {
	switch format {
	case "csv":
		fmt.Println("Date,Type,Amount,Description,Status")
		for _, tx := range transactions {
			fmt.Printf("%s,%s,%.2f,%s,%s\n",
				tx.Date.Format("2006-01-02"),
				tx.Type,
				tx.Amount,
				tx.Description,
				tx.Status,
			)
		}
	case "json":
		return utils.PrintJSON(transactions)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
	return nil
} 