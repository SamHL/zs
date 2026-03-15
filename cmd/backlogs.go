package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SamHL/zs/internal/api"
	"github.com/SamHL/zs/internal/output"
)

var backlogCmd = &cobra.Command{
	Use:   "backlog",
	Short: "Manage the backlog",
	Long:  `View and manage backlog items.`,
}

var backlogListCmd = &cobra.Command{
	Use:   "list",
	Short: "List backlog items",
	Long:  `List all items in the backlog, ordered by priority.`,
	Example: `  # List backlog items
  zs backlog list

  # List in JSON format
  zs backlog list -o json`,
	RunE: runBacklogList,
}

var backlogAddCmd = &cobra.Command{
	Use:   "add <item-id>",
	Short: "Add item to backlog",
	Long:  `Move an item to the backlog, removing it from any sprint.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Add item to backlog
  zs backlog add 12345`,
	RunE: runBacklogAdd,
}

var backlogPrioritizeCmd = &cobra.Command{
	Use:   "prioritize <item-id>",
	Short: "Set backlog item priority",
	Long:  `Set the position of an item in the backlog.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Move item to top of backlog
  zs backlog prioritize 12345 --position 1

  # Move item to position 5
  zs backlog prioritize 12345 --position 5`,
	RunE: runBacklogPrioritize,
}

var backlogMoveToSprintCmd = &cobra.Command{
	Use:   "move-to-sprint",
	Short: "Move backlog items to a sprint",
	Long:  `Move one or more backlog items to a sprint.`,
	Example: `  # Move items to sprint
  zs backlog move-to-sprint --sprint 12345 --items 111,222,333`,
	RunE: runBacklogMoveToSprint,
}

var backlogStatsCmd = &cobra.Command{
	Use:   "stats",
	Short: "Show backlog statistics",
	Long:  `Display statistics about the backlog.`,
	RunE:  runBacklogStats,
}

func init() {
	rootCmd.AddCommand(backlogCmd)
	backlogCmd.AddCommand(backlogListCmd)
	backlogCmd.AddCommand(backlogAddCmd)
	backlogCmd.AddCommand(backlogPrioritizeCmd)
	backlogCmd.AddCommand(backlogMoveToSprintCmd)
	backlogCmd.AddCommand(backlogStatsCmd)

	// Prioritize flags
	backlogPrioritizeCmd.Flags().IntP("position", "P", 1, "position in backlog (1 = top)")
	backlogPrioritizeCmd.MarkFlagRequired("position")

	// Move to sprint flags
	backlogMoveToSprintCmd.Flags().StringP("sprint", "s", "", "sprint ID to move items to (required)")
	backlogMoveToSprintCmd.Flags().StringSliceP("items", "i", []string{}, "item IDs to move (comma-separated)")
	backlogMoveToSprintCmd.MarkFlagRequired("sprint")
	backlogMoveToSprintCmd.MarkFlagRequired("items")
}

func runBacklogList(cmd *cobra.Command, args []string) error {
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

	items, err := client.ListBacklog(teamID, projectID)
	if err != nil {
		return fmt.Errorf("failed to list backlog: %w", err)
	}

	if len(items) == 0 {
		output.PrintInfo("Backlog is empty")
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(items)
	default:
		table := output.NewTableData("PRI", "ID", "NO", "TYPE", "NAME", "POINTS")
		for i, item := range items {
			points := ""
			if item.Points > 0 {
				points = fmt.Sprintf("%d", item.Points)
			}
			table.AddRow(
				fmt.Sprintf("%d", i+1),
				item.ID,
				item.ItemNo,
				item.Type,
				truncate(item.Name, 40),
				points,
			)
		}
		return formatter.Print(table)
	}
}

func runBacklogAdd(cmd *cobra.Command, args []string) error {
	itemID := args[0]
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

	item, err := client.AddToBacklog(teamID, projectID, itemID)
	if err != nil {
		return fmt.Errorf("failed to add to backlog: %w", err)
	}

	output.PrintSuccess("Item added to backlog: %s %s", item.ItemNo, item.Name)
	return nil
}

func runBacklogPrioritize(cmd *cobra.Command, args []string) error {
	itemID := args[0]
	position, _ := cmd.Flags().GetInt("position")

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

	item, err := client.PrioritizeBacklogItem(teamID, projectID, itemID, position)
	if err != nil {
		return fmt.Errorf("failed to prioritize item: %w", err)
	}

	output.PrintSuccess("Item prioritized: %s moved to position %d", item.ItemNo, position)
	return nil
}

func runBacklogMoveToSprint(cmd *cobra.Command, args []string) error {
	sprintID, _ := cmd.Flags().GetString("sprint")
	itemIDs, _ := cmd.Flags().GetStringSlice("items")

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

	if err := client.MoveBacklogItemsToSprint(teamID, projectID, sprintID, itemIDs); err != nil {
		return fmt.Errorf("failed to move items: %w", err)
	}

	output.PrintSuccess("Moved %d items to sprint", len(itemIDs))
	return nil
}

func runBacklogStats(cmd *cobra.Command, args []string) error {
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

	stats, err := client.GetBacklogStats(teamID, projectID)
	if err != nil {
		return fmt.Errorf("failed to get backlog stats: %w", err)
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(stats)
	default:
		fmt.Println("Backlog Statistics")
		fmt.Println("------------------")
		fmt.Printf("Total Items:    %d\n", stats.TotalItems)
		fmt.Printf("Total Points:   %d\n", stats.TotalPoints)
		fmt.Printf("Stories:        %d\n", stats.StoriesCount)
		fmt.Printf("Bugs:           %d\n", stats.BugsCount)
		fmt.Printf("Tasks:          %d\n", stats.TasksCount)
		fmt.Printf("Unestimated:    %d\n", stats.UnestimatedCount)
		return nil
	}
}
