package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/SamHL/zs/internal/config"
)

var (
	// Version info set by main
	version = "dev"
	commit  = "none"
	date    = "unknown"

	// Global flags
	cfgFile    string
	teamID     string
	projectID  string
	outputFmt  string
	debugMode  bool
)

// rootCmd represents the base command
var rootCmd = &cobra.Command{
	Use:   "zs",
	Short: "CLI tool for managing Zoho Sprints",
	Long: `zs is a command-line tool for managing your Zoho Sprints
projects, sprints, and work items.

It provides full CRUD operations for:
  - Teams: List and view team details
  - Projects: Create, update, and manage projects
  - Sprints: Create, start, complete, and manage sprints
  - Items: Create and manage stories, bugs, and tasks
  - Backlogs: Manage and prioritize backlog items
  - Epics: Create and link items to epics

Use 'zs docs' to view LLM-friendly documentation.`,
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		// Skip config init for completion and help commands
		if cmd.Name() == "completion" || cmd.Name() == "help" {
			return nil
		}

		// Initialize config
		if err := config.Init(); err != nil {
			return fmt.Errorf("failed to initialize config: %w", err)
		}

		// Override defaults with flags
		cfg := config.Get()
		if teamID != "" {
			cfg.Defaults.TeamID = teamID
		}
		if projectID != "" {
			cfg.Defaults.ProjectID = projectID
		}
		if outputFmt != "" {
			cfg.Output.Format = outputFmt
		}

		return nil
	},
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute runs the root command
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
		os.Exit(1)
	}
}

// SetVersionInfo sets version information
func SetVersionInfo(v, c, d string) {
	version = v
	commit = c
	date = d
}

// GetVersion returns the version string
func GetVersion() string {
	return version
}

// GetRootCmd returns the root command (for docs generation)
func GetRootCmd() *cobra.Command {
	return rootCmd
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is ~/.zs/config.yaml)")
	rootCmd.PersistentFlags().StringVarP(&teamID, "team", "t", "", "team ID to use (overrides default)")
	rootCmd.PersistentFlags().StringVarP(&projectID, "project", "p", "", "project ID to use (overrides default)")
	rootCmd.PersistentFlags().StringVarP(&outputFmt, "output", "o", "", "output format: json, yaml, table, plain")
	rootCmd.PersistentFlags().BoolVarP(&debugMode, "debug", "d", false, "enable debug mode")

	// Version command
	rootCmd.AddCommand(&cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("zs %s\n", version)
			fmt.Printf("  commit: %s\n", commit)
			fmt.Printf("  built:  %s\n", date)
		},
	})
}

// RequireTeamID returns an error if no team ID is available
func RequireTeamID() (string, error) {
	if teamID != "" {
		return teamID, nil
	}
	cfg := config.Get()
	if cfg.Defaults.TeamID != "" {
		return cfg.Defaults.TeamID, nil
	}
	return "", fmt.Errorf("no team ID specified. Use --team flag or set a default with 'zs teams set-default <team-id>'")
}

// RequireProjectID returns an error if no project ID is available
func RequireProjectID() (string, error) {
	if projectID != "" {
		return projectID, nil
	}
	cfg := config.Get()
	if cfg.Defaults.ProjectID != "" {
		return cfg.Defaults.ProjectID, nil
	}
	return "", fmt.Errorf("no project ID specified. Use --project flag or set a default with 'zs projects set-default <project-id>'")
}

// IsDebug returns whether debug mode is enabled
func IsDebug() bool {
	return debugMode
}

// GetOutputFormat returns the output format to use
func GetOutputFormat() string {
	if outputFmt != "" {
		return outputFmt
	}
	return config.Get().Output.Format
}
