package cmd

import (
	"fmt"
	"net/url"
	"strconv"

	"github.com/spf13/cobra"

	"github.com/SamHL/zs/internal/api"
	"github.com/SamHL/zs/internal/output"
)

var itemsCmd = &cobra.Command{
	Use:     "items",
	Aliases: []string{"item"},
	Short:   "Manage work items",
	Long:    `Create, update, and manage work items (stories, bugs, tasks).`,
}

var itemsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List items",
	Long:  `List work items in the current project.`,
	Example: `  # List all items
  zs items list

  # List items in a specific sprint
  zs items list --sprint 12345

  # List backlog items
  zs items list --backlog

  # List only bugs
  zs items list --type Bug

  # List items assigned to a user
  zs items list --assignee 67890`,
	RunE: runItemsList,
}

var itemsGetCmd = &cobra.Command{
	Use:   "get <item-id>",
	Short: "Get item details",
	Long:  `Get detailed information about a specific item.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runItemsGet,
}

var itemsCreateCmd = &cobra.Command{
	Use:   "create",
	Short: "Create a new item",
	Long:  `Create a new work item (story, bug, or task).`,
	Example: `  # Create a story
  zs items create --type Story --name "User login feature"

  # Create a bug with priority
  zs items create --type Bug --name "Login fails on Safari" --priority High

  # Create a story with points
  zs items create --type Story --name "Add search" --points 5 --description "Implement search functionality"`,
	RunE: runItemsCreate,
}

var itemsUpdateCmd = &cobra.Command{
	Use:   "update <item-id>",
	Short: "Update an item",
	Long:  `Update an item's details.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Update item name
  zs items update 12345 --name "New name"

  # Update item status
  zs items update 12345 --status-id 67890`,
	RunE: runItemsUpdate,
}

var itemsMoveCmd = &cobra.Command{
	Use:   "move <item-id>",
	Short: "Move item to sprint or backlog",
	Long:  `Move an item to a sprint or back to the backlog.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Move to a sprint
  zs items move 12345 --sprint 67890

  # Move to backlog
  zs items move 12345 --backlog`,
	RunE: runItemsMove,
}

var itemsAssignCmd = &cobra.Command{
	Use:   "assign <item-id>",
	Short: "Assign item to a user",
	Long:  `Assign an item to a team member.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Assign to a user
  zs items assign 12345 --user 67890`,
	RunE: runItemsAssign,
}

var itemsDeleteCmd = &cobra.Command{
	Use:   "delete <item-id>",
	Short: "Delete an item",
	Long:  `Delete a work item. This action cannot be undone.`,
	Args:  cobra.ExactArgs(1),
	RunE:  runItemsDelete,
}

var itemsSearchCmd = &cobra.Command{
	Use:   "search <query>",
	Short: "Search for items",
	Long:  `Search for items by name or description.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Search for items
  zs items search "login"`,
	RunE: runItemsSearch,
}

func init() {
	rootCmd.AddCommand(itemsCmd)
	itemsCmd.AddCommand(itemsListCmd)
	itemsCmd.AddCommand(itemsGetCmd)
	itemsCmd.AddCommand(itemsCreateCmd)
	itemsCmd.AddCommand(itemsUpdateCmd)
	itemsCmd.AddCommand(itemsMoveCmd)
	itemsCmd.AddCommand(itemsAssignCmd)
	itemsCmd.AddCommand(itemsDeleteCmd)
	itemsCmd.AddCommand(itemsSearchCmd)

	// List flags
	itemsListCmd.Flags().String("sprint", "", "filter by sprint ID")
	itemsListCmd.Flags().Bool("backlog", false, "show only backlog items")
	itemsListCmd.Flags().String("type", "", "filter by type: Story, Bug, Task")
	itemsListCmd.Flags().String("status", "", "filter by status ID")
	itemsListCmd.Flags().String("assignee", "", "filter by assignee ID")

	// Create flags
	itemsCreateCmd.Flags().StringP("name", "n", "", "item name (required)")
	itemsCreateCmd.Flags().StringP("type", "T", "Story", "item type: Story, Bug, Task")
	itemsCreateCmd.Flags().StringP("description", "d", "", "item description")
	itemsCreateCmd.Flags().IntP("points", "P", 0, "story points")
	itemsCreateCmd.Flags().String("priority", "", "priority: Low, Medium, High")
	itemsCreateCmd.Flags().String("sprint", "", "sprint ID to add item to")
	itemsCreateCmd.Flags().String("epic", "", "epic ID to link item to")
	itemsCreateCmd.Flags().String("assignee", "", "assignee user ID")
	itemsCreateCmd.MarkFlagRequired("name")

	// Update flags
	itemsUpdateCmd.Flags().StringP("name", "n", "", "new item name")
	itemsUpdateCmd.Flags().StringP("description", "d", "", "new description")
	itemsUpdateCmd.Flags().IntP("points", "P", 0, "new story points")
	itemsUpdateCmd.Flags().String("priority", "", "new priority")
	itemsUpdateCmd.Flags().String("status-id", "", "new status ID")

	// Move flags
	itemsMoveCmd.Flags().String("sprint", "", "sprint ID to move to")
	itemsMoveCmd.Flags().Bool("backlog", false, "move to backlog")

	// Assign flags
	itemsAssignCmd.Flags().StringP("user", "u", "", "user ID to assign to (required)")
	itemsAssignCmd.MarkFlagRequired("user")

	// Delete flags
	itemsDeleteCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
}

func runItemsList(cmd *cobra.Command, args []string) error {
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

	params := url.Values{}

	if sprintID, _ := cmd.Flags().GetString("sprint"); sprintID != "" {
		params.Set("sprint_id", sprintID)
	}
	if backlog, _ := cmd.Flags().GetBool("backlog"); backlog {
		params.Set("backlog", "true")
	}
	if itemType, _ := cmd.Flags().GetString("type"); itemType != "" {
		params.Set("item_type", itemType)
	}
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		params.Set("status_id", status)
	}
	if assignee, _ := cmd.Flags().GetString("assignee"); assignee != "" {
		params.Set("assignee", assignee)
	}

	items, err := client.ListItems(teamID, projectID, params)
	if err != nil {
		return fmt.Errorf("failed to list items: %w", err)
	}

	if len(items) == 0 {
		output.PrintInfo("No items found")
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(items)
	default:
		table := output.NewTableData("ID", "NO", "TYPE", "NAME", "STATUS", "ASSIGNEE", "POINTS")
		for _, i := range items {
			points := ""
			if i.Points > 0 {
				points = fmt.Sprintf("%d", i.Points)
			}
			table.AddRow(i.ID, i.ItemNo, i.Type, truncate(i.Name, 40), i.Status, i.AssigneeName, points)
		}
		return formatter.Print(table)
	}
}

func runItemsGet(cmd *cobra.Command, args []string) error {
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

	item, err := client.GetItem(teamID, projectID, itemID)
	if err != nil {
		return fmt.Errorf("failed to get item: %w", err)
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(item)
	default:
		fmt.Printf("Item: %s %s\n", item.ItemNo, item.Name)
		fmt.Printf("ID: %s\n", item.ID)
		fmt.Printf("Type: %s\n", item.Type)
		fmt.Printf("Status: %s\n", item.Status)
		if item.Description != "" {
			fmt.Printf("Description: %s\n", item.Description)
		}
		if item.Priority != "" {
			fmt.Printf("Priority: %s\n", item.Priority)
		}
		if item.Points > 0 {
			fmt.Printf("Points: %d\n", item.Points)
		}
		if item.AssigneeName != "" {
			fmt.Printf("Assignee: %s\n", item.AssigneeName)
		}
		if item.SprintName != "" {
			fmt.Printf("Sprint: %s\n", item.SprintName)
		}
		if item.EpicName != "" {
			fmt.Printf("Epic: %s\n", item.EpicName)
		}
		fmt.Printf("Reporter: %s\n", item.ReporterName)
		fmt.Printf("Created: %s\n", item.CreatedTime)
		if item.UpdatedTime != "" {
			fmt.Printf("Updated: %s\n", item.UpdatedTime)
		}
		return nil
	}
}

func runItemsCreate(cmd *cobra.Command, args []string) error {
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	name, _ := cmd.Flags().GetString("name")
	itemType, _ := cmd.Flags().GetString("type")
	description, _ := cmd.Flags().GetString("description")
	points, _ := cmd.Flags().GetInt("points")
	priority, _ := cmd.Flags().GetString("priority")
	sprintID, _ := cmd.Flags().GetString("sprint")
	epicID, _ := cmd.Flags().GetString("epic")
	assignee, _ := cmd.Flags().GetString("assignee")

	data := map[string]string{
		"name":      name,
		"item_type": itemType,
	}
	if description != "" {
		data["description"] = description
	}
	if points > 0 {
		data["story_points"] = strconv.Itoa(points)
	}
	if priority != "" {
		data["priority"] = priority
	}
	if sprintID != "" {
		data["sprint_id"] = sprintID
	}
	if epicID != "" {
		data["epic_id"] = epicID
	}
	if assignee != "" {
		data["assignee"] = assignee
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	item, err := client.CreateItem(teamID, projectID, data)
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	output.PrintSuccess("%s created: %s %s (%s)", item.Type, item.ItemNo, item.Name, item.ID)

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())
	if GetOutputFormat() == "json" || GetOutputFormat() == "yaml" {
		return formatter.Print(item)
	}
	return nil
}

func runItemsUpdate(cmd *cobra.Command, args []string) error {
	itemID := args[0]
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
	if points, _ := cmd.Flags().GetInt("points"); points > 0 {
		updates["story_points"] = strconv.Itoa(points)
	}
	if priority, _ := cmd.Flags().GetString("priority"); priority != "" {
		updates["priority"] = priority
	}
	if statusID, _ := cmd.Flags().GetString("status-id"); statusID != "" {
		updates["status_id"] = statusID
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates specified")
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	item, err := client.UpdateItem(teamID, projectID, itemID, updates)
	if err != nil {
		return fmt.Errorf("failed to update item: %w", err)
	}

	output.PrintSuccess("Item updated: %s %s", item.ItemNo, item.Name)
	return nil
}

func runItemsMove(cmd *cobra.Command, args []string) error {
	itemID := args[0]
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	sprintID, _ := cmd.Flags().GetString("sprint")
	backlog, _ := cmd.Flags().GetBool("backlog")

	if sprintID == "" && !backlog {
		return fmt.Errorf("specify --sprint or --backlog")
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	var item *api.Item
	if backlog {
		item, err = client.MoveItemToBacklog(teamID, projectID, itemID)
	} else {
		item, err = client.MoveItemToSprint(teamID, projectID, itemID, sprintID)
	}

	if err != nil {
		return fmt.Errorf("failed to move item: %w", err)
	}

	if backlog {
		output.PrintSuccess("Item moved to backlog: %s", item.ItemNo)
	} else {
		output.PrintSuccess("Item moved to sprint: %s", item.ItemNo)
	}
	return nil
}

func runItemsAssign(cmd *cobra.Command, args []string) error {
	itemID := args[0]
	teamID, err := RequireTeamID()
	if err != nil {
		return err
	}
	projectID, err := RequireProjectID()
	if err != nil {
		return err
	}

	userID, _ := cmd.Flags().GetString("user")

	client := api.NewClient()
	client.SetDebug(IsDebug())

	item, err := client.AssignItem(teamID, projectID, itemID, userID)
	if err != nil {
		return fmt.Errorf("failed to assign item: %w", err)
	}

	output.PrintSuccess("Item assigned: %s -> %s", item.ItemNo, item.AssigneeName)
	return nil
}

func runItemsDelete(cmd *cobra.Command, args []string) error {
	itemID := args[0]
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
		item, err := client.GetItem(teamID, projectID, itemID)
		if err != nil {
			return fmt.Errorf("failed to get item: %w", err)
		}

		if !output.Confirm(fmt.Sprintf("Delete item '%s %s'? This cannot be undone", item.ItemNo, item.Name)) {
			output.PrintInfo("Cancelled")
			return nil
		}
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	if err := client.DeleteItem(teamID, projectID, itemID); err != nil {
		return fmt.Errorf("failed to delete item: %w", err)
	}

	output.PrintSuccess("Item deleted")
	return nil
}

func runItemsSearch(cmd *cobra.Command, args []string) error {
	query := args[0]
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

	items, err := client.SearchItems(teamID, projectID, query)
	if err != nil {
		return fmt.Errorf("failed to search items: %w", err)
	}

	if len(items) == 0 {
		output.PrintInfo("No items found matching '%s'", query)
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(items)
	default:
		table := output.NewTableData("ID", "NO", "TYPE", "NAME", "STATUS")
		for _, i := range items {
			table.AddRow(i.ID, i.ItemNo, i.Type, truncate(i.Name, 50), i.Status)
		}
		return formatter.Print(table)
	}
}

// truncate shortens a string to max length
func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
