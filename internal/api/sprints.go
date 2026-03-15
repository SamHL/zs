package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Sprint represents a Zoho Sprints sprint
type Sprint struct {
	ID          string `json:"sprint_id"`
	Name        string `json:"sprint_name"`
	SprintNo    string `json:"sprint_no"`
	Status      string `json:"sprint_status"` // upcoming, active, completed, canceled
	StartDate   string `json:"start_date"`
	EndDate     string `json:"end_date"`
	Duration    string `json:"duration"`
	CreatedTime string `json:"created_time"`
	CreatedBy   string `json:"created_by"`
}

// SprintsRawResponse represents the raw API response for listing sprints
type SprintsRawResponse struct {
	Status          string                   `json:"status"`
	SprintJObj      map[string][]interface{} `json:"sprintJObj"`
	SprintIDs       []string                 `json:"sprintIds"`
	UserDisplayName map[string]string        `json:"userDisplayName"`
}

// SprintResponse represents the API response for a single sprint
type SprintResponse struct {
	Status string `json:"status"`
	Sprint Sprint `json:"sprintJObj"`
}

// parseSprintsResponse parses the raw sprint response into Sprint objects
func parseSprintsResponse(data []byte) ([]Sprint, error) {
	var rawResp SprintsRawResponse
	if err := json.Unmarshal(data, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to parse sprints response: %w", err)
	}

	// Array indices: 0=sprintName, 1=startDate, 2=endDate, 3=completedOn,
	// 4=duration, 5=sprintType, 6=createdBy, 7=createdTime, 8=canceledOn,
	// 9=canceledBy, 10=sprintNo, 11=scrumMaster, 12=workflowId, 13=startAfter
	var sprints []Sprint
	for _, sprintID := range rawResp.SprintIDs {
		if arr, ok := rawResp.SprintJObj[sprintID]; ok && len(arr) >= 11 {
			// sprintType: 1=upcoming, 2=active, 3=completed, 4=canceled
			sprintType := toString(arr[5])
			status := "upcoming"
			switch sprintType {
			case "2":
				status = "active"
			case "3":
				status = "completed"
			case "4":
				status = "canceled"
			}

			createdByID := toString(arr[6])
			createdByName := rawResp.UserDisplayName[createdByID]

			sprints = append(sprints, Sprint{
				ID:          sprintID,
				Name:        toString(arr[0]),
				SprintNo:    toString(arr[10]),
				Status:      status,
				StartDate:   toString(arr[1]),
				EndDate:     toString(arr[2]),
				Duration:    toString(arr[4]),
				CreatedTime: toString(arr[7]),
				CreatedBy:   createdByName,
			})
		}
	}

	return sprints, nil
}

// ListSprints retrieves all sprints in a project
func (c *Client) ListSprints(teamID, projectID string) ([]Sprint, error) {
	path := c.GetProjectPath(teamID, projectID) + "/sprints/"
	params := url.Values{}
	params.Set("action", "data")
	params.Set("index", "1")
	params.Set("range", "100")
	// Include all sprint types: 1=upcoming, 2=active, 3=completed, 4=canceled
	params.Set("type", "[1,2,3,4]")

	data, err := c.Get(path, params)
	if err != nil {
		return nil, err
	}

	return parseSprintsResponse(data)
}

// ListSprintsByStatus retrieves sprints filtered by status
// status: 1=upcoming, 2=active, 3=completed, 4=canceled
func (c *Client) ListSprintsByStatus(teamID, projectID, status string) ([]Sprint, error) {
	path := c.GetProjectPath(teamID, projectID) + "/sprints/"
	params := url.Values{}
	params.Set("action", "data")
	params.Set("index", "1")
	params.Set("range", "100")
	params.Set("type", "["+status+"]")

	data, err := c.Get(path, params)
	if err != nil {
		return nil, err
	}

	return parseSprintsResponse(data)
}

// GetSprint retrieves a specific sprint by ID
func (c *Client) GetSprint(teamID, projectID, sprintID string) (*Sprint, error) {
	// List all sprints and find the one with matching ID
	sprints, err := c.ListSprints(teamID, projectID)
	if err != nil {
		return nil, err
	}
	for _, s := range sprints {
		if s.ID == sprintID {
			return &s, nil
		}
	}
	return nil, fmt.Errorf("sprint not found: %s", sprintID)
}

// CreateSprint creates a new sprint
// startDate and endDate should be in ISO format: yyyy-MM-dd'T'HH:mm:ssZ (e.g., 2024-01-15T00:00:00+10:00)
func (c *Client) CreateSprint(teamID, projectID, name, description, startDate, endDate string) (*Sprint, error) {
	var resp SprintResponse
	path := c.GetProjectPath(teamID, projectID) + "/sprints/"

	data := url.Values{}
	data.Set("name", name)
	data.Set("startdate", startDate)
	data.Set("enddate", endDate)
	if description != "" {
		data.Set("description", description)
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
func (c *Client) StartSprint(teamID, projectID, sprintID string) error {
	path := c.GetProjectPath(teamID, projectID) + "/sprints/" + sprintID + "/"
	data := url.Values{}
	data.Set("action", "startsprint")

	_, err := c.Post(path, data)
	return err
}

// CompleteSprint completes a sprint
func (c *Client) CompleteSprint(teamID, projectID, sprintID string) error {
	path := c.GetProjectPath(teamID, projectID) + "/sprints/" + sprintID + "/"
	data := url.Values{}
	data.Set("action", "completesprint")

	_, err := c.Post(path, data)
	return err
}

// DeleteSprint deletes a sprint
func (c *Client) DeleteSprint(teamID, projectID, sprintID string) error {
	path := c.GetProjectPath(teamID, projectID) + "/sprints/" + sprintID + "/"
	_, err := c.Delete(path)
	return err
}

// GetActiveSprint retrieves the currently active sprint
func (c *Client) GetActiveSprint(teamID, projectID string) (*Sprint, error) {
	// Status type 2 = Active sprint
	sprints, err := c.ListSprintsByStatus(teamID, projectID, "2")
	if err != nil {
		return nil, err
	}
	if len(sprints) == 0 {
		return nil, nil
	}
	return &sprints[0], nil
}
