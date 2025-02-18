package cmd

import (
	"fmt"
	"os"

	"plannet/internal/auth"
	"plannet/internal/config"
	"plannet/pkg/cli"

	"github.com/spf13/cobra"
)

var (
	// Flags
	systemType string
	verbose    bool
)
var (
	cfg *config.Config
	cm  *auth.CredentialManager
)

func initConfig() error {
	var err error
	cfg, err = config.LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	cm = auth.NewCredentialManager(cfg.SystemType)
	return nil
}

func Execute() error {
	if err := initConfig(); err != nil {
		fmt.Println("Error initializing config:", err)
		os.Exit(1)
	}

	rootCmd := &cobra.Command{
		Use:   "plannet",
		Short: "Plannet - Project planning made simple",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			// Skip auto-login for login and init commands
			if cmd.Name() != "login" && cmd.Name() != "init" {
				if !cm.IsAuthenticated(cfg.Username) {
					fmt.Println("Not authenticated. Please run 'plannet login' first.")
					os.Exit(1)
				}
			}
		},
	}

	rootCmd.AddCommand(initCmd())
	// rootCmd.AddCommand(loginCmd())
	// rootCmd.AddCommand(listCmd())
	// rootCmd.AddCommand(logoutCmd())

	return rootCmd.Execute()
}

func initCmd() *cobra.Command {
	var systemType, baseURL string

	cmd := &cobra.Command{
		Use:   "init",
		Short: "Initialize Plannet configuration",
		Run: func(cmd *cobra.Command, args []string) {
			username := cli.PromptUser("Username: ")

			err := config.InitializeConfig(systemType, username, baseURL)
			if err != nil {
				fmt.Printf("Failed to initialize config: %v\n", err)
				os.Exit(1)
			}
			fmt.Println("Configuration initialized successfully!")
		},
	}

	cmd.Flags().StringVar(&systemType, "system", "jira", "Ticket system type (jira, etc)")
	cmd.Flags().StringVar(&baseURL, "url", "", "Base URL for the ticket system")
	cmd.MarkFlagRequired("url")

	return cmd
}

// func loginCmd() *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "login",
// 		Short: "Log in to ticket system",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			if cm.IsAuthenticated(cfg.Username) {
// 				fmt.Println("Already logged in!")
// 				return
// 			}

// 			password := cli.PromptPassword("Password: ")
// 			system, err := createTicketSystem()
// 			if err != nil {
// 				fmt.Println(err)
// 				os.Exit(1)
// 			}

// 			if err := system.Authenticate(cfg.Username, password); err != nil {
// 				fmt.Printf("Login failed: %v\n", err)
// 				os.Exit(1)
// 			}
// 			fmt.Println("Successfully logged in!")
// 		},
// 	}
// }

// func createTicketSystem() (systems.TicketSystem, error) {
// 	// Factory method to create ticket system based on configuration/flags
// 	switch systemType {
// 	case "jira":
// 		return jira.NewTicketSystem(), nil
// 	default:
// 		return nil, fmt.Errorf("unsupported ticket system: %s", systemType)
// 	}
// }

// func listCmd() *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "list",
// 		Short: "List tickets assigned to you",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			system, err := createTicketSystem()
// 			if err != nil {
// 				fmt.Println(err)
// 				os.Exit(1)
// 			}

// 			tickets, err := system.GetAssignedTickets()
// 			if err != nil {
// 				fmt.Println("Failed to retrieve tickets:", err)
// 				os.Exit(1)
// 			}

// 			for _, ticket := range tickets {
// 				fmt.Printf("ID: %s\nTitle: %s\nStatus: %s\nURL: %s\n\n",
// 					ticket.ID, ticket.Title, ticket.Status, ticket.URL)
// 			}
// 		},
// 	}
// }

// func logoutCmd() *cobra.Command {
// 	return &cobra.Command{
// 		Use:   "logout",
// 		Short: "Log out of ticket system",
// 		Run: func(cmd *cobra.Command, args []string) {
// 			system, err := createTicketSystem()
// 			if err != nil {
// 				fmt.Println(err)
// 				os.Exit(1)
// 			}

// 			if err := system.Logout(); err != nil {
// 				fmt.Println("Logout failed:", err)
// 				os.Exit(1)
// 			}
// 			fmt.Println("Successfully logged out!")
// 		},
// 	}
// }
