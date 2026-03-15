package api

import (
	"net/url"
)

// BacklogItem represents an item in the backlog with priority info
type BacklogItem struct {
	Item
	Priority int `json:"backlog_priority"`
}

// BacklogResponse represents the API response for backlog operations
type BacklogResponse struct {
	Status  string        `json:"status"`
	Items   []BacklogItem `json:"itemJObj"`
}

// ListBacklog retrieves the prioritized backlog
func (c *Client) ListBacklog(teamID, projectID string) ([]BacklogItem, error) {
	var resp BacklogResponse
	path := c.GetProjectPath(teamID, projectID) + "/backlog/"
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// AddToBacklog adds an item to the backlog by moving it from a sprint
// Note: Requires the current sprint ID and backlog ID
func (c *Client) AddToBacklog(teamID, projectID, currentSprintID, itemID, backlogID string) (*Item, error) {
	return c.MoveItemToSprint(teamID, projectID, currentSprintID, itemID, backlogID)
}

// PrioritizeBacklogItem sets the priority/position of a backlog item
func (c *Client) PrioritizeBacklogItem(teamID, projectID, itemID string, position int) (*Item, error) {
	path := c.GetProjectPath(teamID, projectID) + "/backlog/" + itemID + "/prioritize/"

	data := url.Values{}
	data.Set("position", string(rune(position)))

	var resp ItemResponse
	err := c.PostJSON(path, data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Item, nil
}

// MoveBacklogItemsToSprint moves multiple backlog items to a sprint
func (c *Client) MoveBacklogItemsToSprint(teamID, projectID, sprintID string, itemIDs []string) error {
	path := c.GetProjectPath(teamID, projectID) + "/backlog/move-to-sprint/"

	data := url.Values{}
	data.Set("sprint_id", sprintID)
	for _, id := range itemIDs {
		data.Add("item_ids", id)
	}

	_, err := c.Post(path, data)
	return err
}

// GetBacklogStats retrieves statistics about the backlog
func (c *Client) GetBacklogStats(teamID, projectID string) (*BacklogStats, error) {
	var resp struct {
		Status string       `json:"status"`
		Stats  BacklogStats `json:"stats"`
	}
	path := c.GetProjectPath(teamID, projectID) + "/backlog/stats/"
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Stats, nil
}

// BacklogStats contains backlog statistics
type BacklogStats struct {
	TotalItems      int `json:"total_items"`
	TotalPoints     int `json:"total_points"`
	StoriesCount    int `json:"stories_count"`
	BugsCount       int `json:"bugs_count"`
	TasksCount      int `json:"tasks_count"`
	UnestimatedCount int `json:"unestimated_count"`
}
