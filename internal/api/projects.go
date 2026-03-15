package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Project represents a Zoho Sprints project
type Project struct {
	ID        string `json:"project_id"`
	Name      string `json:"project_name"`
	Sequence  string `json:"sequence"`
	StartDate string `json:"start_date"`
	EndDate   string `json:"end_date"`
	OwnerID   string `json:"owner_zsoid"`
	OwnerName string `json:"owner_name"`
	GroupID   string `json:"group_id"`
	GroupName string `json:"group_name"`
}

// ProjectsRawResponse represents the raw API response for listing projects
type ProjectsRawResponse struct {
	Status          string                   `json:"status"`
	ProjectJObj     map[string][]interface{} `json:"projectJObj"`
	ProjectIDs      []string                 `json:"projectIds"`
	UserDisplayName map[string]string        `json:"userDisplayName"`
}

// ProjectResponse represents the API response for a single project
type ProjectResponse struct {
	Status  string  `json:"status"`
	Project Project `json:"projectJObj"`
}

// ListProjects retrieves all projects in a team
func (c *Client) ListProjects(teamID string) ([]Project, error) {
	path := c.GetTeamPath(teamID) + "/projects/"
	params := url.Values{}
	params.Set("action", "data")
	params.Set("index", "1")
	params.Set("range", "100")

	data, err := c.Get(path, params)
	if err != nil {
		return nil, err
	}

	var rawResp ProjectsRawResponse
	if err := json.Unmarshal(data, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to parse projects response: %w", err)
	}

	// Parse the custom response format
	// Array indices: 0=name, 1=projNo, 2=startDate, 3=endDate, 4=estimationType,
	// 5=owner, 6=createdBy, 7=sequence, 8=status, 9=favouriteType, 10=groupId, 11=groupName, 12=workflowId
	var projects []Project
	for _, projectID := range rawResp.ProjectIDs {
		if arr, ok := rawResp.ProjectJObj[projectID]; ok && len(arr) >= 12 {
			ownerID := toString(arr[5])
			ownerName := rawResp.UserDisplayName[ownerID]

			projects = append(projects, Project{
				ID:        projectID,
				Name:      toString(arr[0]),
				Sequence:  toString(arr[7]),
				StartDate: toString(arr[2]),
				EndDate:   toString(arr[3]),
				OwnerID:   ownerID,
				OwnerName: ownerName,
				GroupID:   toString(arr[10]),
				GroupName: toString(arr[11]),
			})
		}
	}

	return projects, nil
}

// toString converts an interface{} to string
func toString(v interface{}) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return fmt.Sprintf("%.0f", val)
	default:
		return fmt.Sprintf("%v", val)
	}
}

// GetProject retrieves a specific project by ID
func (c *Client) GetProject(teamID, projectID string) (*Project, error) {
	// List all projects and find the one with matching ID
	projects, err := c.ListProjects(teamID)
	if err != nil {
		return nil, err
	}
	for _, p := range projects {
		if p.ID == projectID {
			return &p, nil
		}
	}
	return nil, fmt.Errorf("project not found: %s", projectID)
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
