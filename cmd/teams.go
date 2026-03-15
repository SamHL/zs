package cmd

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/SamHL/zs/internal/api"
	"github.com/SamHL/zs/internal/config"
	"github.com/SamHL/zs/internal/output"
)

var teamsCmd = &cobra.Command{
	Use:   "teams",
	Short: "Manage teams",
	Long:  `List and manage Zoho Sprints teams.`,
}

var teamsListCmd = &cobra.Command{
	Use:   "list",
	Short: "List all teams",
	Long:  `List all teams you have access to in Zoho Sprints.`,
	Example: `  # List all teams
  zs teams list

  # List teams in JSON format
  zs teams list -o json`,
	RunE: runTeamsList,
}

var teamsGetCmd = &cobra.Command{
	Use:   "get <team-id>",
	Short: "Get team details",
	Long:  `Get detailed information about a specific team.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Get team details
  zs teams get 12345`,
	RunE: runTeamsGet,
}

var teamsSetDefaultCmd = &cobra.Command{
	Use:   "set-default <team-id>",
	Short: "Set the default team",
	Long:  `Set a team as the default for all commands.`,
	Args:  cobra.ExactArgs(1),
	Example: `  # Set default team
  zs teams set-default 12345`,
	RunE: runTeamsSetDefault,
}

var teamsMembersCmd = &cobra.Command{
	Use:   "members [team-id]",
	Short: "List team members",
	Long:  `List all members of a team.`,
	Example: `  # List members of default team
  zs teams members

  # List members of specific team
  zs teams members 12345`,
	RunE: runTeamsMembers,
}

func init() {
	rootCmd.AddCommand(teamsCmd)
	teamsCmd.AddCommand(teamsListCmd)
	teamsCmd.AddCommand(teamsGetCmd)
	teamsCmd.AddCommand(teamsSetDefaultCmd)
	teamsCmd.AddCommand(teamsMembersCmd)
}

func runTeamsList(cmd *cobra.Command, args []string) error {
	client := api.NewClient()
	client.SetDebug(IsDebug())

	teams, err := client.ListTeams()
	if err != nil {
		return fmt.Errorf("failed to list teams: %w", err)
	}

	if len(teams) == 0 {
		output.PrintInfo("No teams found")
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(teams)
	default:
		table := output.NewTableData("ID", "NAME", "OWNER", "MEMBERS")
		for _, t := range teams {
			table.AddRow(t.ID, t.Name, t.OwnerName, fmt.Sprintf("%d", t.MemberCount))
		}
		return formatter.Print(table)
	}
}

func runTeamsGet(cmd *cobra.Command, args []string) error {
	teamID := args[0]

	client := api.NewClient()
	client.SetDebug(IsDebug())

	team, err := client.GetTeam(teamID)
	if err != nil {
		return fmt.Errorf("failed to get team: %w", err)
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(team)
	default:
		fmt.Printf("Team: %s\n", team.Name)
		fmt.Printf("ID: %s\n", team.ID)
		if team.Description != "" {
			fmt.Printf("Description: %s\n", team.Description)
		}
		fmt.Printf("Owner: %s\n", team.OwnerName)
		fmt.Printf("Members: %d\n", team.MemberCount)
		fmt.Printf("Created: %s\n", team.CreatedTime)
		return nil
	}
}

func runTeamsSetDefault(cmd *cobra.Command, args []string) error {
	teamID := args[0]

	// Verify team exists
	client := api.NewClient()
	client.SetDebug(IsDebug())

	team, err := client.GetTeam(teamID)
	if err != nil {
		return fmt.Errorf("failed to verify team: %w", err)
	}

	if err := config.SetDefault("team_id", teamID); err != nil {
		return fmt.Errorf("failed to set default: %w", err)
	}

	output.PrintSuccess("Default team set to: %s (%s)", team.Name, teamID)
	return nil
}

func runTeamsMembers(cmd *cobra.Command, args []string) error {
	var teamID string
	if len(args) > 0 {
		teamID = args[0]
	} else {
		var err error
		teamID, err = RequireTeamID()
		if err != nil {
			return err
		}
	}

	client := api.NewClient()
	client.SetDebug(IsDebug())

	members, err := client.ListTeamMembers(teamID)
	if err != nil {
		return fmt.Errorf("failed to list members: %w", err)
	}

	if len(members) == 0 {
		output.PrintInfo("No members found")
		return nil
	}

	formatter := output.NewFormatter().WithFormat(GetOutputFormat())

	switch GetOutputFormat() {
	case "json", "yaml":
		return formatter.Print(members)
	default:
		table := output.NewTableData("ID", "NAME", "EMAIL", "ROLE", "STATUS")
		for _, m := range members {
			table.AddRow(m.ID, m.Name, m.Email, m.Role, m.Status)
		}
		return formatter.Print(table)
	}
}
