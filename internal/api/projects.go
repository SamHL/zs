package api

import (
	"fmt"
	"net/url"
)

// Project represents a Zoho Sprints project
type Project struct {
	ID            string `json:"project_id"`
	Name          string `json:"project_name"`
	Description   string `json:"project_desc"`
	Status        string `json:"status"`
	OwnerID       string `json:"owner_zsoid"`
	OwnerName     string `json:"owner_name"`
	Prefix        string `json:"project_prefix"`
	CreatedTime   string `json:"created_time"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	SprintCount   int    `json:"sprint_count"`
	ItemCount     int    `json:"item_count"`
	BacklogCount  int    `json:"backlog_count"`
	CompletedCount int   `json:"completed_count"`
}

// ProjectsResponse represents the API response for listing projects
type ProjectsResponse struct {
	Status   string    `json:"status"`
	Projects []Project `json:"projectJObj"`
}

// ProjectResponse represents the API response for a single project
type ProjectResponse struct {
	Status  string  `json:"status"`
	Project Project `json:"projectJObj"`
}

// ListProjects retrieves all projects in a team
func (c *Client) ListProjects(teamID string) ([]Project, error) {
	var resp ProjectsResponse
	path := c.GetTeamPath(teamID) + "/projects/"
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Projects, nil
}

// GetProject retrieves a specific project by ID
func (c *Client) GetProject(teamID, projectID string) (*Project, error) {
	var resp ProjectResponse
	path := c.GetProjectPath(teamID, projectID) + "/"
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Project, nil
}

// CreateProject creates a new project
func (c *Client) CreateProject(teamID, name, description, prefix string) (*Project, error) {
	var resp ProjectResponse
	path := c.GetTeamPath(teamID) + "/projects/"

	data := url.Values{}
	data.Set("name", name)
	if description != "" {
		data.Set("description", description)
	}
	if prefix != "" {
		data.Set("prefix", prefix)
	}

	err := c.PostJSON(path, data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Project, nil
}

// UpdateProject updates a project's details
func (c *Client) UpdateProject(teamID, projectID string, updates map[string]string) (*Project, error) {
	var resp ProjectResponse
	path := c.GetProjectPath(teamID, projectID) + "/"

	data := url.Values{}
	for k, v := range updates {
		data.Set(k, v)
	}

	err := c.PutJSON(path, data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Project, nil
}

// DeleteProject deletes a project
func (c *Client) DeleteProject(teamID, projectID string) error {
	path := c.GetProjectPath(teamID, projectID) + "/"
	_, err := c.Delete(path)
	return err
}

// ProjectStatus represents available project statuses
type ProjectStatus struct {
	ID    string `json:"status_id"`
	Name  string `json:"status_name"`
	Color string `json:"status_color"`
	Type  string `json:"status_type"` // open, in_progress, completed
}

// GetProjectStatuses retrieves available statuses for a project
func (c *Client) GetProjectStatuses(teamID, projectID string) ([]ProjectStatus, error) {
	var resp struct {
		Status   string          `json:"status"`
		Statuses []ProjectStatus `json:"statusJObj"`
	}
	path := fmt.Sprintf("%s/itemstatus/", c.GetProjectPath(teamID, projectID))
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Statuses, nil
}
