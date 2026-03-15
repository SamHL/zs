package api

import (
	"net/url"
)

// Team represents a Zoho Sprints team
type Team struct {
	ID          string `json:"zsoid"`
	Name        string `json:"team_name"`
	Description string `json:"team_desc"`
	OwnerID     string `json:"owner_zsoid"`
	OwnerName   string `json:"owner_name"`
	CreatedTime string `json:"created_time"`
	LogoURL     string `json:"logo_url"`
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
	var resp TeamResponse
	path := c.GetTeamPath(teamID) + "/"
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Team, nil
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
func (c *Client) ListTeamMembers(teamID string) ([]TeamMember, error) {
	var resp TeamMembersResponse
	path := c.GetTeamPath(teamID) + "/users/"
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
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
