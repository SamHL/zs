package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SamHL/zs/internal/api"
	"github.com/SamHL/zs/internal/output"
)

var epicsCmd = &cobra.Command{
	Use:   "epics",
	Short: "Manage epics",
	Long:  `Create, update, and manage epics.`,
}

var epicsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all epics",
	Long:  `List all epics in the current project.`,
	Example: `  # List all epics
  zs epics list

  # List epics in JSON format
  zs epics list -o json`,
	RunE: runEpicsList,
}

var epicsGetCmd = &cobra.Command{
	Use:   "get <epic-id>",
	Short: "Get epic details",
	Long:  `Get detailed information about a specific epic.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEpicsGet,
}

var epicsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new epic",
	Long:  `Create a new epic.`,
	Example: `  # Create an epic
  zs epics create --name "User Authentication"

  # Create an epic with description and color
  zs epics create --name "Payment Integration" --description "Add payment processing" --color "#FF5733"`,
	RunE: runEpicsCreate,
}

var epicsUpdateCmd = &cobra.Command{
	Use:   "update <epic-id>",
	Short: "Update an epic",
	Long:  `Update an epic's details.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEpicsUpdate,
}

var epicsDeleteCmd = &cobra.Command{
	Use:   "delete <epic-id>",
	Short: "Delete an epic",
	Long:  `Delete an epic. Linked items will be unlinked but not deleted.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEpicsDelete,
}

var epicsLinkCmd = &cobra.Command{
	Use:   "link <epic-id> <item-id>",
	Short: "Link an item to an epic",
	Long:  `Link a work item to an epic.`,
	Args:  cobra.ExactArgs(2),
	Example: `  # Link item to epic
  zs epics link 12345 67890`,
	RunE: runEpicsLink,
}

var epicsUnlinkCmd = &cobra.Command{
	Use:   "unlink <epic-id> <item-id>",
	Short: "Unlink an item from an epic",
	Long:  `Remove the link between an item and an epic.`,
	Args:  cobra.ExactArgs(2),
	RunE:  runEpicsUnlink,
}

var epicsItemsCmd = &cobra.Command{
	Use:   "items <epic-id>",
	Short: "List items in an epic",
	Long:  `List all items linked to a specific epic.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runEpicsItems,
}

func init() {
	rootCmd.AddCommand(epicsCmd)
	epicsCmd.AddCommand(epicsListCmd)
	epicsCmd.AddCommand(epicsGetCmd)
	epicsCmd.AddCommand(epicsCreateCmd)
	epicsCmd.AddCommand(epicsUpdateCmd)
	epicsCmd.AddCommand(epicsDeleteCmd)
	epicsCmd.AddCommand(epicsLinkCmd)
	epicsCmd.AddCommand(epicsUnlinkCmd)
	epicsCmd.AddCommand(epicsItemsCmd)

	// Create flags
	epicsCreateCmd.Flags().StringP("name", "n", "", "epic name (required)")
	epicsCreateCmd.Flags().StringP("description", "d", "", "epic description")
	epicsCreateCmd.Flags().StringP("color", "c", "", "epic color (hex code)")
	epicsCreateCmd.MarkFlagRequired("name")

	// Update flags
	epicsUpdateCmd.Flags().StringP("name", "n", "", "new epic name")
	epicsUpdateCmd.Flags().StringP("description", "d", "", "new description")
	epicsUpdateCmd.Flags().StringP("color", "c", "", "new color")
	epicsUpdateCmd.Flags().String("status", "", "new status")

	// Delete flags
	epicsDeleteCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
}

func runEpicsList(cmd *cobra.Command, args []string) error {
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

	epics, err := client.ListEpics(teamID, projectID)
	if err != nil {
		return fmt.Errorf("failed to list epics: %w", err)
	}

	if len(epics) == 0 {
		output.PrintInfo("No epics found")
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(epics)
	default:
		table := output.NewTableData("ID", "NAME", "STATUS", "ITEMS", "POINTS", "PROGRESS")
		for _, e := range epics {
			progress := "0%"
			if e.TotalPoints > 0 {
				progress = fmt.Sprintf("%d%%", (e.CompletedPoints*100)/e.TotalPoints)
			} else if e.ItemCount > 0 {
				progress = fmt.Sprintf("%d%%", (e.CompletedCount*100)/e.ItemCount)
			}
			table.AddRow(
				e.ID,
				e.Name,
				e.Status,
				fmt.Sprintf("%d/%d", e.CompletedCount, e.ItemCount),
				fmt.Sprintf("%d/%d", e.CompletedPoints, e.TotalPoints),
				progress,
			)
		}
		return formatter.Print(table)
	}
}

func runEpicsGet(cmd *cobra.Command, args []string) error {
	epicID := args[0]
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

	epic, err := client.GetEpic(teamID, projectID, epicID)
	if err != nil {
		return fmt.Errorf("failed to get epic: %w", err)
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(epic)
	default:
		fmt.Printf("Epic: %s\n", epic.Name)
		fmt.Printf("ID: %s\n", epic.ID)
		fmt.Printf("Status: %s\n", epic.Status)
		if epic.Description != "" {
			fmt.Printf("Description: %s\n", epic.Description)
		}
		if epic.Color != "" {
			fmt.Printf("Color: %s\n", epic.Color)
		}
		fmt.Printf("Owner: %s\n", epic.OwnerName)
		fmt.Printf("Items: %d completed, %d total\n", epic.CompletedCount, epic.ItemCount)
		fmt.Printf("Points: %d completed, %d total\n", epic.CompletedPoints, epic.TotalPoints)
		if epic.StartDate != "" {
			fmt.Printf("Duration: %s to %s\n", epic.StartDate, epic.EndDate)
		}
		fmt.Printf("Created: %s\n", epic.CreatedTime)
		return nil
	}
}

func runEpicsCreate(cmd *cobra.Command, args []string) error {
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	description, _ := cmd.Flags().GetString("description")
	color, _ := cmd.Flags().GetString("color")

	client := api.NewClient()
	client.SetDebug(IsDebug())

	epic, err := client.CreateEpic(teamID, projectID, name, description, color)
	if err != nil {
		return fmt.Errorf("failed to create epic: %w", err)
	}

	output.PrintSuccess("Epic created: %s (%s)", epic.Name, epic.ID)

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())
	if GetOutputFormat() == "json" || GetOutputFormat() == "yaml" {
		return formatter.Print(epic)
	}
	return nil
}

func runEpicsUpdate(cmd *cobra.Command, args []string) error {
	epicID := args[0]
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	updates := make(map[string]string)

	if name, _ := cmd.Flags().GetString("name"); name != "" {
		updates["name"] = name
	}
	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		updates["description"] = desc
	}
	if color, _ := cmd.Flags().GetString("color"); color != "" {
		updates["color"] = color
	}
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		updates["status"] = status
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates specified")
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	epic, err := client.UpdateEpic(teamID, projectID, epicID, updates)
	if err != nil {
		return fmt.Errorf("failed to update epic: %w", err)
	}

	output.PrintSuccess("Epic updated: %s", epic.Name)
	return nil
}

func runEpicsDelete(cmd *cobra.Command, args []string) error {
	epicID := args[0]
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
		epic, err := client.GetEpic(teamID, projectID, epicID)
		if err != nil {
			return fmt.Errorf("failed to get epic: %w", err)
		}

		if !output.Confirm(fmt.Sprintf("Delete epic '%s'? Linked items will be unlinked", epic.Name)) {
			output.PrintInfo("Cancelled")
			return nil
		}
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	if err := client.DeleteEpic(teamID, projectID, epicID); err != nil {
		return fmt.Errorf("failed to delete epic: %w", err)
	}

	output.PrintSuccess("Epic deleted")
	return nil
}

func runEpicsLink(cmd *cobra.Command, args []string) error {
	epicID := args[0]
	itemID := args[1]
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

	item, err := client.LinkItemToEpic(teamID, projectID, itemID, epicID)
	if err != nil {
		return fmt.Errorf("failed to link item: %w", err)
	}

	output.PrintSuccess("Item %s linked to epic", item.ItemNo)
	return nil
}

func runEpicsUnlink(cmd *cobra.Command, args []string) error {
	itemID := args[1]
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

	item, err := client.UnlinkItemFromEpic(teamID, projectID, itemID)
	if err != nil {
		return fmt.Errorf("failed to unlink item: %w", err)
	}

	output.PrintSuccess("Item %s unlinked from epic", item.ItemNo)
	return nil
}

func runEpicsItems(cmd *cobra.Command, args []string) error {
	epicID := args[0]
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

	items, err := client.ListEpicItems(teamID, projectID, epicID)
	if err != nil {
		return fmt.Errorf("failed to list epic items: %w", err)
	}

	if len(items) == 0 {
		output.PrintInfo("No items in this epic")
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(items)
	default:
		table := output.NewTableData("ID", "NO", "TYPE", "NAME", "STATUS", "SPRINT")
		for _, i := range items {
			table.AddRow(i.ID, i.ItemNo, i.Type, truncate(i.Name, 35), i.Status, i.SprintName)
		}
		return formatter.Print(table)
	}
}
