package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SamHL/zs/internal/api"
	"github.com/SamHL/zs/internal/output"
)

var sprintsCmd = &cobra.Command{
	Use:   "sprints",
	Short: "Manage sprints",
	Long:  `Create, start, complete, and manage sprints.`,
}

var sprintsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List sprints",
	Long:  `List all sprints in the current project.`,
	Example: `  # List all sprints
  zs sprints list

  # List only active sprints
  zs sprints list --status active

  # List sprints in JSON format
  zs sprints list -o json`,
	RunE: runSprintsList,
}

var sprintsGetCmd = &cobra.Command{
	Use:   "get <sprint-id>",
	Short: "Get sprint details",
	Long:  `Get detailed information about a specific sprint.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSprintsGet,
}

var sprintsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new sprint",
	Long:  `Create a new sprint with a name, start date, and end date.`,
	Example: `  # Create a sprint
  zs sprints create --name "Sprint 1" --start 2024-01-01 --end 2024-01-14

  # Create a sprint with a goal
  zs sprints create --name "Sprint 1" --start 2024-01-01 --end 2024-01-14 --goal "Complete auth feature"`,
	RunE: runSprintsCreate,
}

var sprintsStartCmd = &cobra.Command{
	Use:   "start <sprint-id>",
	Short: "Start a sprint",
	Long:  `Start a sprint, changing its status to active.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSprintsStart,
}

var sprintsCompleteCmd = &cobra.Command{
	Use:   "complete <sprint-id>",
	Short: "Complete a sprint",
	Long:  `Complete a sprint, marking it as finished.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSprintsComplete,
}

var sprintsDeleteCmd = &cobra.Command{
	Use:   "delete <sprint-id>",
	Short: "Delete a sprint",
	Long:  `Delete a sprint. Items in the sprint will be moved to backlog.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runSprintsDelete,
}

var sprintsActiveCmd = &cobra.Command{
	Use:   "active",
	Short: "Get the current active sprint",
	Long:  `Get information about the currently active sprint.`,
	RunE:  runSprintsActive,
}

func init() {
	rootCmd.AddCommand(sprintsCmd)
	sprintsCmd.AddCommand(sprintsListCmd)
	sprintsCmd.AddCommand(sprintsGetCmd)
	sprintsCmd.AddCommand(sprintsCreateCmd)
	sprintsCmd.AddCommand(sprintsStartCmd)
	sprintsCmd.AddCommand(sprintsCompleteCmd)
	sprintsCmd.AddCommand(sprintsDeleteCmd)
	sprintsCmd.AddCommand(sprintsActiveCmd)

	// List flags (status: 1=upcoming, 2=active, 3=completed, 4=canceled)
	sprintsListCmd.Flags().StringP("status", "s", "", "filter by status type: 1=upcoming, 2=active, 3=completed, 4=canceled")

	// Create flags
	sprintsCreateCmd.Flags().StringP("name", "n", "", "sprint name (required)")
	sprintsCreateCmd.Flags().String("start", "", "start date ISO format (YYYY-MM-DDTHH:MM:SS+TZ) (required)")
	sprintsCreateCmd.Flags().String("end", "", "end date ISO format (YYYY-MM-DDTHH:MM:SS+TZ) (required)")
	sprintsCreateCmd.Flags().String("description", "", "sprint description")
	sprintsCreateCmd.MarkFlagRequired("name")
	sprintsCreateCmd.MarkFlagRequired("start")
	sprintsCreateCmd.MarkFlagRequired("end")

	// Delete flags
	sprintsDeleteCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
}

func runSprintsList(cmd *cobra.Command, args []string) error {
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	status, _ := cmd.Flags().GetString("status")

	var sprints []api.Sprint
	if status != "" {
		sprints, err = client.ListSprintsByStatus(teamID, projectID, status)
	} else {
		sprints, err = client.ListSprints(teamID, projectID)
	}

	if err != nil {
		return fmt.Errorf("failed to list sprints: %w", err)
	}

	if len(sprints) == 0 {
		output.PrintInfo("No sprints found")
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(sprints)
	default:
		table := output.NewTableData("ID", "NO", "NAME", "STATUS", "DURATION", "START", "END")
		for _, s := range sprints {
			table.AddRow(
				s.ID,
				s.SprintNo,
				s.Name,
				s.Status,
				s.Duration,
				s.StartDate,
				s.EndDate,
			)
		}
		return formatter.Print(table)
	}
}

func runSprintsGet(cmd *cobra.Command, args []string) error {
	sprintID := args[0]
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	sprint, err := client.GetSprint(teamID, projectID, sprintID)
	if err != nil {
		return fmt.Errorf("failed to get sprint: %w", err)
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(sprint)
	default:
		fmt.Printf("Sprint: %s (#%s)\n", sprint.Name, sprint.SprintNo)
		fmt.Printf("ID: %s\n", sprint.ID)
		fmt.Printf("Status: %s\n", sprint.Status)
		fmt.Printf("Duration: %s (%s to %s)\n", sprint.Duration, sprint.StartDate, sprint.EndDate)
		fmt.Printf("Created: %s by %s\n", sprint.CreatedTime, sprint.CreatedBy)
		return nil
	}
}

func runSprintsCreate(cmd *cobra.Command, args []string) error {
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	startDate, _ := cmd.Flags().GetString("start")
	endDate, _ := cmd.Flags().GetString("end")
	description, _ := cmd.Flags().GetString("description")

	client := api.NewClient()
	client.SetDebug(IsDebug())

	sprint, err := client.CreateSprint(teamID, projectID, name, description, startDate, endDate)
	if err != nil {
		return fmt.Errorf("failed to create sprint: %w", err)
	}

	output.PrintSuccess("Sprint created: %s (%s)", sprint.Name, sprint.ID)

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())
	if GetOutputFormat() == "json" || GetOutputFormat() == "yaml" {
		return formatter.Print(sprint)
	}
	return nil
}

func runSprintsStart(cmd *cobra.Command, args []string) error {
	sprintID := args[0]
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	if err := client.StartSprint(teamID, projectID, sprintID); err != nil {
		return fmt.Errorf("failed to start sprint: %w", err)
	}

	output.PrintSuccess("Sprint started: %s", sprintID)
	return nil
}

func runSprintsComplete(cmd *cobra.Command, args []string) error {
	sprintID := args[0]
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	if err := client.CompleteSprint(teamID, projectID, sprintID); err != nil {
		return fmt.Errorf("failed to complete sprint: %w", err)
	}

	output.PrintSuccess("Sprint completed: %s", sprintID)
	return nil
}

func runSprintsDelete(cmd *cobra.Command, args []string) error {
	sprintID := args[0]
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	force, _ := cmd.Flags().GetBool("force")

	if !force {
		client := api.NewClient()
		sprint, err := client.GetSprint(teamID, projectID, sprintID)
		if err != nil {
			return fmt.Errorf("failed to get sprint: %w", err)
		}

		if !output.Confirm(fmt.Sprintf("Delete sprint '%s'? Items will be moved to backlog", sprint.Name)) {
			output.PrintInfo("Cancelled")
			return nil
		}
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	if err := client.DeleteSprint(teamID, projectID, sprintID); err != nil {
		return fmt.Errorf("failed to delete sprint: %w", err)
	}

	output.PrintSuccess("Sprint deleted")
	return nil
}

func runSprintsActive(cmd *cobra.Command, args []string) error {
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	sprint, err := client.GetActiveSprint(teamID, projectID)
	if err != nil {
		return fmt.Errorf("failed to get active sprint: %w", err)
	}

	if sprint == nil {
		output.PrintInfo("No active sprint")
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(sprint)
	default:
		fmt.Printf("Active Sprint: %s (#%s)\n", sprint.Name, sprint.SprintNo)
		fmt.Printf("ID: %s\n", sprint.ID)
		fmt.Printf("Duration: %s (%s to %s)\n", sprint.Duration, sprint.StartDate, sprint.EndDate)
		return nil
	}
}
