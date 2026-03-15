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

var itemsTypesCmd = &cobra.Command{
	Use:   "types",
	Short: "List available item types",
	Long:  `List all available item types (Story, Bug, Task, etc.) for the current project.`,
	RunE:  runItemsTypes,
}

var itemsPrioritiesCmd = &cobra.Command{
	Use:   "priorities",
	Short: "List available priorities",
	Long:  `List all available priority levels for the current project.`,
	RunE:  runItemsPriorities,
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
	itemsCmd.AddCommand(itemsTypesCmd)
	itemsCmd.AddCommand(itemsPrioritiesCmd)

	// List flags
	itemsListCmd.Flags().StringP("sprint", "s", "", "sprint ID (required)")
	itemsListCmd.Flags().String("type", "", "filter by type: Story, Bug, Task")
	itemsListCmd.Flags().String("status", "", "filter by status ID")
	itemsListCmd.Flags().String("assignee", "", "filter by assignee ID")
	itemsListCmd.MarkFlagRequired("sprint")

	// Get flags
	itemsGetCmd.Flags().StringP("sprint", "s", "", "sprint ID (required)")
	itemsGetCmd.MarkFlagRequired("sprint")

	// Create flags - support both friendly names and raw IDs
	itemsCreateCmd.Flags().StringP("name", "n", "", "item name (required)")
	itemsCreateCmd.Flags().StringP("sprint", "s", "", "sprint ID (required)")
	itemsCreateCmd.Flags().String("description", "", "item description")
	itemsCreateCmd.Flags().IntP("points", "P", 0, "story points")
	itemsCreateCmd.Flags().String("type", "", "item type name (e.g., Story, Bug, Task)")
	itemsCreateCmd.Flags().String("priority", "", "priority name (e.g., High, Medium, Low)")
	itemsCreateCmd.Flags().String("item-type-id", "", "item type ID (alternative to --type)")
	itemsCreateCmd.Flags().String("priority-id", "", "priority ID (alternative to --priority)")
	itemsCreateCmd.Flags().String("epic", "", "epic ID to link item to")
	itemsCreateCmd.Flags().String("assignee", "", "assignee user ID")
	itemsCreateCmd.MarkFlagRequired("name")
	itemsCreateCmd.MarkFlagRequired("sprint")

	// Update flags
	itemsUpdateCmd.Flags().StringP("sprint", "s", "", "sprint ID (required)")
	itemsUpdateCmd.Flags().StringP("name", "n", "", "new item name")
	itemsUpdateCmd.Flags().String("description", "", "new description")
	itemsUpdateCmd.Flags().IntP("points", "P", 0, "new story points")
	itemsUpdateCmd.Flags().String("status-id", "", "new status ID")
	itemsUpdateCmd.MarkFlagRequired("sprint")

	// Move flags
	itemsMoveCmd.Flags().String("from-sprint", "", "current sprint ID (required)")
	itemsMoveCmd.Flags().String("to-sprint", "", "target sprint ID (required)")
	itemsMoveCmd.MarkFlagRequired("from-sprint")
	itemsMoveCmd.MarkFlagRequired("to-sprint")

	// Assign flags
	itemsAssignCmd.Flags().StringP("sprint", "s", "", "sprint ID (required)")
	itemsAssignCmd.Flags().StringP("user", "u", "", "user ID to assign to (required)")
	itemsAssignCmd.MarkFlagRequired("sprint")
	itemsAssignCmd.MarkFlagRequired("user")

	// Delete flags
	itemsDeleteCmd.Flags().StringP("sprint", "s", "", "sprint ID (required)")
	itemsDeleteCmd.Flags().BoolP("force", "f", false, "skip confirmation prompt")
	itemsDeleteCmd.MarkFlagRequired("sprint")

	// Search flags
	itemsSearchCmd.Flags().StringP("sprint", "s", "", "sprint ID (required)")
	itemsSearchCmd.MarkFlagRequired("sprint")
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
	sprintID, _ := cmd.Flags().GetString("sprint")

	client := api.NewClient()
	client.SetDebug(IsDebug())

	params := url.Values{}

	if itemType, _ := cmd.Flags().GetString("type"); itemType != "" {
		params.Set("item_type", itemType)
	}
	if status, _ := cmd.Flags().GetString("status"); status != "" {
		params.Set("status_id", status)
	}
	if assignee, _ := cmd.Flags().GetString("assignee"); assignee != "" {
		params.Set("assignee", assignee)
	}

	items, err := client.ListItems(teamID, projectID, sprintID, params)
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
		table := output.NewTableData("ID", "NO", "NAME", "DURATION", "POINTS", "CREATED BY")
		for _, i := range items {
			points := ""
			if i.Points > 0 {
				points = fmt.Sprintf("%d", i.Points)
			}
			table.AddRow(i.ID, i.ItemNo, truncate(i.Name, 50), i.Duration, points, i.CreatedBy)
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
	sprintID, _ := cmd.Flags().GetString("sprint")

	client := api.NewClient()
	client.SetDebug(IsDebug())

	item, err := client.GetItem(teamID, projectID, sprintID, itemID)
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
		if item.Duration != "" {
			fmt.Printf("Duration: %s\n", item.Duration)
		}
		if item.Points > 0 {
			fmt.Printf("Points: %d\n", item.Points)
		}
		fmt.Printf("Created by: %s\n", item.CreatedBy)
		fmt.Printf("Created: %s\n", item.CreatedTime)
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
	sprintID, _ := cmd.Flags().GetString("sprint")
	description, _ := cmd.Flags().GetString("description")
	points, _ := cmd.Flags().GetInt("points")
	epicID, _ := cmd.Flags().GetString("epic")
	assignee, _ := cmd.Flags().GetString("assignee")

	// Get type - either by name or raw ID
	typeName, _ := cmd.Flags().GetString("type")
	itemTypeID, _ := cmd.Flags().GetString("item-type-id")

	// Get priority - either by name or raw ID
	priorityName, _ := cmd.Flags().GetString("priority")
	priorityID, _ := cmd.Flags().GetString("priority-id")

	client := api.NewClient()
	client.SetDebug(IsDebug())

	// Resolve type name to ID if provided
	if typeName != "" {
		itemTypeID, err = client.ResolveItemType(teamID, projectID, typeName)
		if err != nil {
			return fmt.Errorf("failed to resolve item type: %w", err)
		}
	}
	if itemTypeID == "" {
		return fmt.Errorf("item type required: use --type (e.g., Story, Bug) or --item-type-id")
	}

	// Resolve priority name to ID if provided
	if priorityName != "" {
		priorityID, err = client.ResolvePriority(teamID, projectID, priorityName)
		if err != nil {
			return fmt.Errorf("failed to resolve priority: %w", err)
		}
	}
	if priorityID == "" {
		return fmt.Errorf("priority required: use --priority (e.g., High, Medium, Low) or --priority-id")
	}

	data := map[string]string{
		"name":           name,
		"projitemtypeid": itemTypeID,
		"projpriorityid": priorityID,
	}
	if description != "" {
		data["description"] = description
	}
	if points > 0 {
		data["point"] = strconv.Itoa(points)
	}
	if epicID != "" {
		data["epicid"] = epicID
	}
	if assignee != "" {
		data["users"] = "[\"" + assignee + "\"]"
	}

	item, err := client.CreateItem(teamID, projectID, sprintID, data)
	if err != nil {
		return fmt.Errorf("failed to create item: %w", err)
	}

	output.PrintSuccess("Item created: %s %s (%s)", item.ItemNo, item.Name, item.ID)

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
	sprintID, _ := cmd.Flags().GetString("sprint")

	updates := make(map[string]string)

	if name, _ := cmd.Flags().GetString("name"); name != "" {
		updates["name"] = name
	}
	if desc, _ := cmd.Flags().GetString("description"); desc != "" {
		updates["description"] = desc
	}
	if points, _ := cmd.Flags().GetInt("points"); points > 0 {
		updates["point"] = strconv.Itoa(points)
	}
	if statusID, _ := cmd.Flags().GetString("status-id"); statusID != "" {
		updates["statusid"] = statusID
	}

	if len(updates) == 0 {
		return fmt.Errorf("no updates specified")
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	item, err := client.UpdateItem(teamID, projectID, sprintID, itemID, updates)
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

	fromSprintID, _ := cmd.Flags().GetString("from-sprint")
	toSprintID, _ := cmd.Flags().GetString("to-sprint")

	client := api.NewClient()
	client.SetDebug(IsDebug())

	item, err := client.MoveItemToSprint(teamID, projectID, fromSprintID, itemID, toSprintID)
	if err != nil {
		return fmt.Errorf("failed to move item: %w", err)
	}

	output.PrintSuccess("Item moved to sprint: %s", item.ItemNo)
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
	sprintID, _ := cmd.Flags().GetString("sprint")
	userID, _ := cmd.Flags().GetString("user")

	client := api.NewClient()
	client.SetDebug(IsDebug())

	item, err := client.AssignItem(teamID, projectID, sprintID, itemID, userID)
	if err != nil {
		return fmt.Errorf("failed to assign item: %w", err)
	}

	output.PrintSuccess("Item assigned: %s", item.ItemNo)
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
	sprintID, _ := cmd.Flags().GetString("sprint")
	force, _ := cmd.Flags().GetBool("force")

	client := api.NewClient()
	client.SetDebug(IsDebug())

	if !force {
		item, err := client.GetItem(teamID, projectID, sprintID, itemID)
		if err != nil {
			return fmt.Errorf("failed to get item: %w", err)
		}

		if !output.Confirm(fmt.Sprintf("Delete item '%s %s'? This cannot be undone", item.ItemNo, item.Name)) {
			output.PrintInfo("Cancelled")
			return nil
		}
	}

	if err := client.DeleteItem(teamID, projectID, sprintID, itemID); err != nil {
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
	sprintID, _ := cmd.Flags().GetString("sprint")

	client := api.NewClient()
	client.SetDebug(IsDebug())

	items, err := client.SearchItems(teamID, projectID, sprintID, query)
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
		table := output.NewTableData("ID", "NO", "NAME")
		for _, i := range items {
			table.AddRow(i.ID, i.ItemNo, truncate(i.Name, 60))
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

func runItemsTypes(cmd *cobra.Command, args []string) error {
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

	itemTypes, err := client.ListItemTypes(teamID, projectID)
	if err != nil {
		return fmt.Errorf("failed to list item types: %w", err)
	}

	if len(itemTypes) == 0 {
		output.PrintInfo("No item types found")
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(itemTypes)
	default:
		table := output.NewTableData("ID", "NAME")
		for _, t := range itemTypes {
			table.AddRow(t.ID, t.Name)
		}
		return formatter.Print(table)
	}
}

func runItemsPriorities(cmd *cobra.Command, args []string) error {
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

	priorities, err := client.ListPriorities(teamID, projectID)
	if err != nil {
		return fmt.Errorf("failed to list priorities: %w", err)
	}

	if len(priorities) == 0 {
		output.PrintInfo("No priorities found")
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(priorities)
	default:
		table := output.NewTableData("ID", "NAME")
		for _, p := range priorities {
			table.AddRow(p.ID, p.Name)
		}
		return formatter.Print(table)
	}
}
