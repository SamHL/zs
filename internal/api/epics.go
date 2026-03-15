package api

import (
	"net/url"
)

// Epic represents an epic in Zoho Sprints
type Epic struct {
	ID            string `json:"epic_id"`
	Name          string `json:"epic_name"`
	Description   string `json:"epic_desc"`
	Status        string `json:"epic_status"`
	Color         string `json:"epic_color"`
	OwnerID       string `json:"owner_zsoid"`
	OwnerName     string `json:"owner_name"`
	StartDate     string `json:"start_date"`
	EndDate       string `json:"end_date"`
	ItemCount     int    `json:"item_count"`
	CompletedCount int   `json:"completed_count"`
	TotalPoints   int    `json:"total_points"`
	CompletedPoints int  `json:"completed_points"`
	CreatedTime   string `json:"created_time"`
	UpdatedTime   string `json:"updated_time"`
}

// EpicsResponse represents the API response for listing epics
type EpicsResponse struct {
	Status string `json:"status"`
	Epics  []Epic `json:"epicJObj"`
}

// EpicResponse represents the API response for a single epic
type EpicResponse struct {
	Status string `json:"status"`
	Epic   Epic   `json:"epicJObj"`
}

// ListEpics retrieves all epics in a project
func (c *Client) ListEpics(teamID, projectID string) ([]Epic, error) {
	var resp EpicsResponse
	path := c.GetProjectPath(teamID, projectID) + "/epics/"
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Epics, nil
}

// GetEpic retrieves a specific epic by ID
func (c *Client) GetEpic(teamID, projectID, epicID string) (*Epic, error) {
	var resp EpicResponse
	path := c.GetProjectPath(teamID, projectID) + "/epics/" + epicID + "/"
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Epic, nil
}

// CreateEpic creates a new epic
func (c *Client) CreateEpic(teamID, projectID, name, description, color string) (*Epic, error) {
	var resp EpicResponse
	path := c.GetProjectPath(teamID, projectID) + "/epics/"

	data := url.Values{}
	data.Set("name", name)
	if description != "" {
		data.Set("description", description)
	}
	if color != "" {
		data.Set("color", color)
	}

	err := c.PostJSON(path, data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Epic, nil
}

// UpdateEpic updates an epic's details
func (c *Client) UpdateEpic(teamID, projectID, epicID string, updates map[string]string) (*Epic, error) {
	var resp EpicResponse
	path := c.GetProjectPath(teamID, projectID) + "/epics/" + epicID + "/"

	data := url.Values{}
	for k, v := range updates {
		data.Set(k, v)
	}

	err := c.PutJSON(path, data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Epic, nil
}

// DeleteEpic deletes an epic
func (c *Client) DeleteEpic(teamID, projectID, epicID string) error {
	path := c.GetProjectPath(teamID, projectID) + "/epics/" + epicID + "/"
	_, err := c.Delete(path)
	return err
}

// ListEpicItems retrieves all items linked to an epic
func (c *Client) ListEpicItems(teamID, projectID, epicID string) ([]Item, error) {
	params := url.Values{}
	params.Set("epic_id", epicID)
	return c.ListItems(teamID, projectID, params)
}

// LinkItemToEpic links an item to an epic
func (c *Client) LinkItemToEpic(teamID, projectID, itemID, epicID string) (*Item, error) {
	return c.UpdateItem(teamID, projectID, itemID, map[string]string{
		"epic_id": epicID,
	})
}

// UnlinkItemFromEpic removes an item's epic link
func (c *Client) UnlinkItemFromEpic(teamID, projectID, itemID string) (*Item, error) {
	return c.UpdateItem(teamID, projectID, itemID, map[string]string{
		"epic_id": "",
	})
}
