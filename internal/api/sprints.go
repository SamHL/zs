package api

import (
	"net/url"
)

// Sprint represents a Zoho Sprints sprint
type Sprint struct {
	ID             string `json:"sprint_id"`
	Name           string `json:"sprint_name"`
	Description    string `json:"sprint_desc"`
	Status         string `json:"sprint_status"` // not_started, active, completed
	StartDate      string `json:"start_date"`
	EndDate        string `json:"end_date"`
	Goal           string `json:"sprint_goal"`
	ItemCount      int    `json:"item_count"`
	CompletedCount int    `json:"completed_count"`
	TotalPoints    int    `json:"total_points"`
	CompletedPoints int   `json:"completed_points"`
	CreatedTime    string `json:"created_time"`
	CreatedBy      string `json:"created_by"`
	Velocity       int    `json:"velocity"`
}

// SprintsResponse represents the API response for listing sprints
type SprintsResponse struct {
	Status  string   `json:"status"`
	Sprints []Sprint `json:"sprintJObj"`
}

// SprintResponse represents the API response for a single sprint
type SprintResponse struct {
	Status string `json:"status"`
	Sprint Sprint `json:"sprintJObj"`
}

// ListSprints retrieves all sprints in a project
func (c *Client) ListSprints(teamID, projectID string) ([]Sprint, error) {
	var resp SprintsResponse
	path := c.GetProjectPath(teamID, projectID) + "/sprints/"
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Sprints, nil
}

// ListSprintsByStatus retrieves sprints filtered by status
func (c *Client) ListSprintsByStatus(teamID, projectID, status string) ([]Sprint, error) {
	var resp SprintsResponse
	path := c.GetProjectPath(teamID, projectID) + "/sprints/"
	params := url.Values{}
	params.Set("status", status)
	err := c.GetJSON(path, params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Sprints, nil
}

// GetSprint retrieves a specific sprint by ID
func (c *Client) GetSprint(teamID, projectID, sprintID string) (*Sprint, error) {
	var resp SprintResponse
	path := c.GetProjectPath(teamID, projectID) + "/sprints/" + sprintID + "/"
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Sprint, nil
}

// CreateSprint creates a new sprint
func (c *Client) CreateSprint(teamID, projectID, name, startDate, endDate, goal string) (*Sprint, error) {
	var resp SprintResponse
	path := c.GetProjectPath(teamID, projectID) + "/sprints/"

	data := url.Values{}
	data.Set("name", name)
	data.Set("start_date", startDate)
	data.Set("end_date", endDate)
	if goal != "" {
		data.Set("goal", goal)
	}

	err := c.PostJSON(path, data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Sprint, nil
}

// UpdateSprint updates a sprint's details
func (c *Client) UpdateSprint(teamID, projectID, sprintID string, updates map[string]string) (*Sprint, error) {
	var resp SprintResponse
	path := c.GetProjectPath(teamID, projectID) + "/sprints/" + sprintID + "/"

	data := url.Values{}
	for k, v := range updates {
		data.Set(k, v)
	}

	err := c.PutJSON(path, data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Sprint, nil
}

// StartSprint starts a sprint (changes status to active)
func (c *Client) StartSprint(teamID, projectID, sprintID string) (*Sprint, error) {
	return c.UpdateSprint(teamID, projectID, sprintID, map[string]string{
		"status": "active",
	})
}

// CompleteSprint completes a sprint
func (c *Client) CompleteSprint(teamID, projectID, sprintID string) (*Sprint, error) {
	return c.UpdateSprint(teamID, projectID, sprintID, map[string]string{
		"status": "completed",
	})
}

// DeleteSprint deletes a sprint
func (c *Client) DeleteSprint(teamID, projectID, sprintID string) error {
	path := c.GetProjectPath(teamID, projectID) + "/sprints/" + sprintID + "/"
	_, err := c.Delete(path)
	return err
}

// GetActiveSprint retrieves the currently active sprint
func (c *Client) GetActiveSprint(teamID, projectID string) (*Sprint, error) {
	sprints, err := c.ListSprintsByStatus(teamID, projectID, "active")
	if err != nil {
		return nil, err
	}
	if len(sprints) == 0 {
		return nil, nil
	}
	return &sprints[0], nil
}
