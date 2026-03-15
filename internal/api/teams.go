package api

import (
	"fmt"
	"net/url"
	"strings"
)

// Team represents a Zoho Sprints team
type Team struct {
	ID          string `json:"zsoid"`
	Name        string `json:"teamName"`
	OrgName     string `json:"orgName"`
	OwnerID     string `json:"portalOwner"`
	OwnerName   string `json:"owner_name"`
	Type        int    `json:"type"`
	IsShowTeam  bool   `json:"isShowTeam"`
	MemberCount int    `json:"member_count"`
}

// TeamsResponse represents the API response for listing teams
type TeamsResponse struct {
	Status    string `json:"status"`
	PortalID  string `json:"portal_zsoid"`
	TeamsList []Team `json:"portals"`
}

// TeamResponse represents the API response for a single team
type TeamResponse struct {
	Status string `json:"status"`
	Team   Team   `json:"portal"`
}

// ListTeams retrieves all teams the user has access to
func (c *Client) ListTeams() ([]Team, error) {
	var resp TeamsResponse
	err := c.GetJSON("/teams/", nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.TeamsList, nil
}

// GetTeam retrieves a specific team by ID
func (c *Client) GetTeam(teamID string) (*Team, error) {
	// List all teams and find the one with matching ID
	teams, err := c.ListTeams()
	if err != nil {
		return nil, err
	}
	for _, t := range teams {
		if t.ID == teamID {
			return &t, nil
		}
	}
	return nil, fmt.Errorf("team not found: %s", teamID)
}

// TeamMember represents a team member
type TeamMember struct {
	ID       string `json:"zsoid"`
	Name     string `json:"name"`
	Email    string `json:"email"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	ImageURL string `json:"image_url"`
}

// TeamMembersResponse represents the API response for team members
type TeamMembersResponse struct {
	Status  string       `json:"status"`
	Members []TeamMember `json:"users"`
}

// ListTeamMembers retrieves all members of a team
// Note: This endpoint requires additional OAuth scopes that may not be available
// in the standard Zoho Sprints API. Consider using the userDisplayName map from
// other API responses to get user information.
func (c *Client) ListTeamMembers(teamID string) ([]TeamMember, error) {
	var resp TeamMembersResponse
	path := c.GetTeamPath(teamID) + "/users/"
	params := url.Values{}
	params.Set("action", "data")
	params.Set("index", "1")
	params.Set("range", "100")
	err := c.GetJSON(path, params, &resp)
	if err != nil {
		// Check for OAuth scope error and provide helpful message
		if strings.Contains(err.Error(), "oauthscope") || strings.Contains(err.Error(), "7601") {
			return nil, fmt.Errorf("users API requires additional OAuth scopes not available in standard access. User names are available in item/sprint listings via userDisplayName")
		}
		return nil, err
	}
	return resp.Members, nil
}

// SearchTeamMembers searches for team members by name or email
func (c *Client) SearchTeamMembers(teamID, query string) ([]TeamMember, error) {
	var resp TeamMembersResponse
	path := c.GetTeamPath(teamID) + "/users/"
	params := url.Values{}
	params.Set("search", query)
	err := c.GetJSON(path, params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Members, nil
}
