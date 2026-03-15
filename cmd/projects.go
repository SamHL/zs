package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SamHL/zs/internal/api"
	"github.com/SamHL/zs/internal/config"
	"github.com/SamHL/zs/internal/output"
)

var projectsCmd = &cobra.Command{
	Use:   "projects",
	Short: "Manage projects",
	Long:  `Create, update, and manage Zoho Sprints projects.`,
}

var projectsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all projects",
	Long:  `List all projects in the current team.`,
	Example: `  # List projects in default team
  zs projects list

  # List projects in specific team
  zs projects list --team 12345`,
	RunE: runProjectsList,
}

var projectsGetCmd = &cobra.Command{
	Use:   "get <project-id>",
	Short: "Get project details",
	Long:  `Get detailed information about a specific project.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Get project details
  zs projects get 12345`,
	RunE: runProjectsGet,
}

var projectsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new project",
	Long:  `Create a new project in the current team.`,
	Example: `  # Create a project
  zs projects create --name "My Project" --description "Project description"

  # Create a project with prefix
  zs projects create --name "API Project" --prefix "API"`,
	RunE: runProjectsCreate,
}

var projectsUpdateCmd = &cobra.Command{
	Use:   "update <project-id>",
	Short: "Update a project",
	Long:  `Update a project's details.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Update project name
  zs projects update 12345 --name "New Name"`,
	RunE: runProjectsUpdate,
}

var projectsDeleteCmd = &cobra.Command{
	Use:   "delete <project-id>",
	Short: "Delete a project",
	Long:  `Delete a project. This action cannot be undone.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Delete a project
  zs projects delete 12345`,
	RunE: runProjectsDelete,
}

var projectsSetDefaultCmd = &cobra.Command{
	Use:   "set-default <project-id>",
	Short: "Set the default project",
	Long:  `Set a project as the default for all commands.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Set default project
  zs projects set-default 12345`,
	RunE: runProjectsSetDefault,
}

func init() {
	rootCmd.AddCommand(projectsCmd)
	projectsCmd.AddCommand(projectsListCmd)
	projectsCmd.AddCommand(projectsGetCmd)
	projectsCmd.AddCommand(projectsCreateCmd)
	projectsCmd.AddCommand(projectsUpdateCmd)
	projectsCmd.AddCommand(projectsDeleteCmd)
	projectsCmd.AddCommand(projectsSetDefaultCmd)

	// Create flags
	projectsCreateCmd.Flags().StringP("name", "n", "", "project name (required)")
	projectsCreateCmd.Flags().StringP("description", "d", "", "project description")
	projectsCreateCmd.Flags().String("prefix", "", "project prefix for item IDs")
	projectsCreateCmd.MarkFlagRequired("name")

	// Update flags
	projectsUpdateCmd.Flags().StringP("name", "n", "", "new project name")
	projectsUpdateCmd.Flags().StringP("description", "d", "", "new project description")

	// Delete flags
	projectsDeleteCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
}

func runProjectsList(cmd *cobra.Command, args []string) error {
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	projects, err := client.ListProjects(teamID)
	if err != nil {
		return fmt.Errorf("failed to list projects: %w", err)
	}

	if len(projects) == 0 {
		output.PrintInfo("No projects found")
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(projects)
	default:
		table := output.NewTableData("ID", "NAME", "OWNER", "GROUP")
		for _, p := range projects {
			table.AddRow(p.ID, p.Name, p.OwnerName, p.GroupName)
		}
		return formatter.Print(table)
	}
}

func runProjectsGet(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	project, err := client.GetProject(teamID, projectID)
	if err != nil {
		return fmt.Errorf("failed to get project: %w", err)
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(project)
	default:
		fmt.Printf("Project: %s\n", project.Name)
		fmt.Printf("ID: %s\n", project.ID)
		if project.OwnerName != "" {
			fmt.Printf("Owner: %s\n", project.OwnerName)
		}
		if project.GroupName != "" {
			fmt.Printf("Group: %s\n", project.GroupName)
		}
		return nil
	}
}

func runProjectsCreate(cmd *cobra.Command, args []string) error {
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	prefix, _ := cmd.Flags().GetString("prefix")

	client := api.NewClient()
	client.SetDebug(IsDebug())

	project, err := client.CreateProject(teamID, name, description, prefix)
	if err != nil {
		return fmt.Errorf("failed to create project: %w", err)
	}

	output.PrintSuccess("Project created: %s (%s)", project.Name, project.ID)

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())
	if GetOutputFormat() == "json" || GetOutputFormat() == "yaml" {
		return formatter.Print(project)
	}
	return nil
}

func runProjectsUpdate(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}

	updates := make(map[string]string)
	if name, _ := cmd.Flags().GetString("name"); name != "" {
		updates["name"] = name
	}
	if description, _ := cmd.Flags().GetString("description"); description != "" {
		updates["description"] = description
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates specified")
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	project, err := client.UpdateProject(teamID, projectID, updates)
	if err != nil {
		return fmt.Errorf("failed to update project: %w", err)
	}

	output.PrintSuccess("Project updated: %s", project.Name)
	return nil
}

func runProjectsDelete(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")

	if !force {
		// Get project info first
		client := api.NewClient()
		project, err := client.GetProject(teamID, projectID)
		if err != nil {
			return fmt.Errorf("failed to get project: %w", err)
		}

		if !output.Confirm(fmt.Sprintf("Delete project '%s'? This cannot be undone", project.Name)) {
			output.PrintInfo("Cancelled")
			return nil
		}
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	if err := client.DeleteProject(teamID, projectID); err != nil {
		return fmt.Errorf("failed to delete project: %w", err)
	}

	output.PrintSuccess("Project deleted")
	return nil
}

func runProjectsSetDefault(cmd *cobra.Command, args []string) error {
	projectID := args[0]
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}

	// Verify project exists
	client := api.NewClient()
	client.SetDebug(IsDebug())

	project, err := client.GetProject(teamID, projectID)
	if err != nil {
		return fmt.Errorf("failed to verify project: %w", err)
	}

	if err := config.SetDefault("project_id", projectID); err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	output.PrintSuccess("Default project set to: %s (%s)", project.Name, projectID)
	return nil
}
