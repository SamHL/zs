package api

import (
	"fmt"
	"net/url"
)

// Item represents a work item (story, bug, task) in Zoho Sprints
type Item struct {
	ID            string `json:"item_id"`
	ItemNo        string `json:"item_no"`
	Name          string `json:"item_name"`
	Description   string `json:"item_desc"`
	Type          string `json:"item_type"`   // Story, Bug, Task
	Status        string `json:"status_name"`
	StatusID      string `json:"status_id"`
	Priority      string `json:"priority"`
	Points        int    `json:"story_points"`
	SprintID      string `json:"sprint_id"`
	SprintName    string `json:"sprint_name"`
	EpicID        string `json:"epic_id"`
	EpicName      string `json:"epic_name"`
	AssigneeID    string `json:"assignee_zsoid"`
	AssigneeName  string `json:"assignee_name"`
	ReporterID    string `json:"reporter_zsoid"`
	ReporterName  string `json:"reporter_name"`
	CreatedTime   string `json:"created_time"`
	UpdatedTime   string `json:"updated_time"`
	DueDate       string `json:"due_date"`
	StartDate     string `json:"start_date"`
	TimeLogged    int    `json:"time_logged"`
	TimeEstimate  int    `json:"time_estimate"`
	SubItemCount  int    `json:"sub_item_count"`
	CommentCount  int    `json:"comment_count"`
	AttachmentCount int  `json:"attachment_count"`
}

// ItemsResponse represents the API response for listing items
type ItemsResponse struct {
	Status string `json:"status"`
	Items  []Item `json:"itemJObj"`
}

// ItemResponse represents the API response for a single item
type ItemResponse struct {
	Status string `json:"status"`
	Item   Item   `json:"itemJObj"`
}

// ListItems retrieves all items in a project
func (c *Client) ListItems(teamID, projectID string, params url.Values) ([]Item, error) {
	var resp ItemsResponse
	path := c.GetProjectPath(teamID, projectID) + "/items/"
	err := c.GetJSON(path, params, &resp)
	if err != nil {
		return nil, err
	}
	return resp.Items, nil
}

// ListSprintItems retrieves all items in a sprint
func (c *Client) ListSprintItems(teamID, projectID, sprintID string) ([]Item, error) {
	params := url.Values{}
	params.Set("sprint_id", sprintID)
	return c.ListItems(teamID, projectID, params)
}

// ListBacklogItems retrieves all backlog items (not in any sprint)
func (c *Client) ListBacklogItems(teamID, projectID string) ([]Item, error) {
	params := url.Values{}
	params.Set("backlog", "true")
	return c.ListItems(teamID, projectID, params)
}

// ListItemsByStatus retrieves items filtered by status
func (c *Client) ListItemsByStatus(teamID, projectID, statusID string) ([]Item, error) {
	params := url.Values{}
	params.Set("status_id", statusID)
	return c.ListItems(teamID, projectID, params)
}

// ListItemsByType retrieves items filtered by type
func (c *Client) ListItemsByType(teamID, projectID, itemType string) ([]Item, error) {
	params := url.Values{}
	params.Set("item_type", itemType)
	return c.ListItems(teamID, projectID, params)
}

// ListItemsByAssignee retrieves items assigned to a specific user
func (c *Client) ListItemsByAssignee(teamID, projectID, assigneeID string) ([]Item, error) {
	params := url.Values{}
	params.Set("assignee", assigneeID)
	return c.ListItems(teamID, projectID, params)
}

// GetItem retrieves a specific item by ID
func (c *Client) GetItem(teamID, projectID, itemID string) (*Item, error) {
	var resp ItemResponse
	path := c.GetProjectPath(teamID, projectID) + "/items/" + itemID + "/"
	err := c.GetJSON(path, nil, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Item, nil
}

// CreateItem creates a new item
func (c *Client) CreateItem(teamID, projectID string, itemData map[string]string) (*Item, error) {
	var resp ItemResponse
	path := c.GetProjectPath(teamID, projectID) + "/items/"

	data := url.Values{}
	for k, v := range itemData {
		data.Set(k, v)
	}

	err := c.PostJSON(path, data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Item, nil
}

// CreateStory creates a new user story
func (c *Client) CreateStory(teamID, projectID, name, description string, points int) (*Item, error) {
	data := map[string]string{
		"name":         name,
		"item_type":    "Story",
	}
	if description != "" {
		data["description"] = description
	}
	if points > 0 {
		data["story_points"] = fmt.Sprintf("%d", points)
	}
	return c.CreateItem(teamID, projectID, data)
}

// CreateBug creates a new bug
func (c *Client) CreateBug(teamID, projectID, name, description, priority string) (*Item, error) {
	data := map[string]string{
		"name":      name,
		"item_type": "Bug",
	}
	if description != "" {
		data["description"] = description
	}
	if priority != "" {
		data["priority"] = priority
	}
	return c.CreateItem(teamID, projectID, data)
}

// CreateTask creates a new task
func (c *Client) CreateTask(teamID, projectID, name, description string) (*Item, error) {
	data := map[string]string{
		"name":      name,
		"item_type": "Task",
	}
	if description != "" {
		data["description"] = description
	}
	return c.CreateItem(teamID, projectID, data)
}

// UpdateItem updates an item's details
func (c *Client) UpdateItem(teamID, projectID, itemID string, updates map[string]string) (*Item, error) {
	var resp ItemResponse
	path := c.GetProjectPath(teamID, projectID) + "/items/" + itemID + "/"

	data := url.Values{}
	for k, v := range updates {
		data.Set(k, v)
	}

	err := c.PutJSON(path, data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Item, nil
}

// MoveItemToSprint moves an item to a sprint
func (c *Client) MoveItemToSprint(teamID, projectID, itemID, sprintID string) (*Item, error) {
	return c.UpdateItem(teamID, projectID, itemID, map[string]string{
		"sprint_id": sprintID,
	})
}

// MoveItemToBacklog moves an item to the backlog
func (c *Client) MoveItemToBacklog(teamID, projectID, itemID string) (*Item, error) {
	return c.UpdateItem(teamID, projectID, itemID, map[string]string{
		"sprint_id": "",
	})
}

// AssignItem assigns an item to a user
func (c *Client) AssignItem(teamID, projectID, itemID, assigneeID string) (*Item, error) {
	return c.UpdateItem(teamID, projectID, itemID, map[string]string{
		"assignee": assigneeID,
	})
}

// UpdateItemStatus updates an item's status
func (c *Client) UpdateItemStatus(teamID, projectID, itemID, statusID string) (*Item, error) {
	return c.UpdateItem(teamID, projectID, itemID, map[string]string{
		"status_id": statusID,
	})
}

// DeleteItem deletes an item
func (c *Client) DeleteItem(teamID, projectID, itemID string) error {
	path := c.GetProjectPath(teamID, projectID) + "/items/" + itemID + "/"
	_, err := c.Delete(path)
	return err
}

// SearchItems searches for items by name or description
func (c *Client) SearchItems(teamID, projectID, query string) ([]Item, error) {
	params := url.Values{}
	params.Set("search", query)
	return c.ListItems(teamID, projectID, params)
}
