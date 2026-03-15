package api

import (
	"encoding/json"
	"fmt"
	"net/url"
)

// Item represents a work item (story, bug, task) in Zoho Sprints
type Item struct {
	ID          string `json:"item_id"`
	ItemNo      string `json:"item_no"`
	Name        string `json:"item_name"`
	Points      int    `json:"points"`
	Duration    string `json:"duration"`
	CreatedBy   string `json:"created_by"`
	CreatedTime string `json:"created_time"`
	StatusID    string `json:"status_id"`
	TypeID      string `json:"type_id"`
	PriorityID  string `json:"priority_id"`
	SprintID    string `json:"sprint_id"`
}

// ItemsRawResponse represents the raw API response for listing items
type ItemsRawResponse struct {
	Status          string                   `json:"status"`
	ItemJObj        map[string][]interface{} `json:"itemJObj"`
	ItemIDs         []string                 `json:"itemIds"`
	UserDisplayName map[string]string        `json:"userDisplayName"`
}

// ItemResponse represents the API response for a single item
type ItemResponse struct {
	Status string `json:"status"`
	Item   Item   `json:"itemJObj"`
}

// parseItemsResponse parses the raw items response into Item objects
func parseItemsResponse(data []byte) ([]Item, error) {
	var rawResp ItemsRawResponse
	if err := json.Unmarshal(data, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to parse items response: %w", err)
	}

	// Array indices from item_prop:
	// 0=itemName, 1=itemNo, 2=createdBy, 3=duration, 12=sequence, 13=points,
	// 16=createdTime, 28=statusId, 29=projItemTypeId, 30=projPriorityId, 32=sprintId
	var items []Item
	for _, itemID := range rawResp.ItemIDs {
		if arr, ok := rawResp.ItemJObj[itemID]; ok && len(arr) >= 33 {
			createdByID := toString(arr[2])
			createdByName := rawResp.UserDisplayName[createdByID]

			points := 0
			if p, ok := arr[13].(float64); ok {
				points = int(p)
			}

			items = append(items, Item{
				ID:          itemID,
				Name:        toString(arr[0]),
				ItemNo:      toString(arr[1]),
				Duration:    toString(arr[3]),
				Points:      points,
				CreatedBy:   createdByName,
				CreatedTime: toString(arr[16]),
				StatusID:    toString(arr[28]),
				TypeID:      toString(arr[29]),
				PriorityID:  toString(arr[30]),
				SprintID:    toString(arr[32]),
			})
		}
	}

	return items, nil
}

// ListItems retrieves all items in a sprint or backlog
// sprintOrBacklogID should be a sprint ID or backlog ID
func (c *Client) ListItems(teamID, projectID, sprintOrBacklogID string, extraParams url.Values) ([]Item, error) {
	path := c.GetProjectPath(teamID, projectID) + "/sprints/" + sprintOrBacklogID + "/item/"
	params := url.Values{}
	params.Set("action", "data")
	params.Set("index", "1")
	params.Set("range", "100")
	// Merge extra params
	for k, v := range extraParams {
		params[k] = v
	}

	data, err := c.Get(path, params)
	if err != nil {
		return nil, err
	}

	return parseItemsResponse(data)
}

// ListSprintItems retrieves all items in a sprint
func (c *Client) ListSprintItems(teamID, projectID, sprintID string) ([]Item, error) {
	return c.ListItems(teamID, projectID, sprintID, nil)
}

// ListBacklogItems retrieves all backlog items
// Note: You need to get the backlog ID from the project details first
func (c *Client) ListBacklogItems(teamID, projectID, backlogID string) ([]Item, error) {
	return c.ListItems(teamID, projectID, backlogID, nil)
}

// ListItemsByStatus retrieves items filtered by status
func (c *Client) ListItemsByStatus(teamID, projectID, sprintID, statusID string) ([]Item, error) {
	params := url.Values{}
	params.Set("status_id", statusID)
	return c.ListItems(teamID, projectID, sprintID, params)
}

// ListItemsByType retrieves items filtered by type
func (c *Client) ListItemsByType(teamID, projectID, sprintID, itemType string) ([]Item, error) {
	params := url.Values{}
	params.Set("item_type", itemType)
	return c.ListItems(teamID, projectID, sprintID, params)
}

// ListItemsByAssignee retrieves items assigned to a specific user
func (c *Client) ListItemsByAssignee(teamID, projectID, sprintID, assigneeID string) ([]Item, error) {
	params := url.Values{}
	params.Set("assignee", assigneeID)
	return c.ListItems(teamID, projectID, sprintID, params)
}

// GetItem retrieves a specific item by ID
func (c *Client) GetItem(teamID, projectID, sprintID, itemID string) (*Item, error) {
	items, err := c.ListItems(teamID, projectID, sprintID, nil)
	if err != nil {
		return nil, err
	}
	for _, i := range items {
		if i.ID == itemID {
			return &i, nil
		}
	}
	return nil, fmt.Errorf("item not found: %s", itemID)
}

// CreateItem creates a new item in a sprint or backlog
func (c *Client) CreateItem(teamID, projectID, sprintID string, itemData map[string]string) (*Item, error) {
	var resp ItemResponse
	path := c.GetProjectPath(teamID, projectID) + "/sprints/" + sprintID + "/item/"

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
func (c *Client) CreateStory(teamID, projectID, sprintID, name, description string, points int, itemTypeID, priorityID string) (*Item, error) {
	data := map[string]string{
		"name":           name,
		"projitemtypeid": itemTypeID,
		"projpriorityid": priorityID,
	}
	if description != "" {
		data["description"] = description
	}
	if points > 0 {
		data["point"] = fmt.Sprintf("%d", points)
	}
	return c.CreateItem(teamID, projectID, sprintID, data)
}

// CreateBug creates a new bug
func (c *Client) CreateBug(teamID, projectID, sprintID, name, description, itemTypeID, priorityID string) (*Item, error) {
	data := map[string]string{
		"name":           name,
		"projitemtypeid": itemTypeID,
		"projpriorityid": priorityID,
	}
	if description != "" {
		data["description"] = description
	}
	return c.CreateItem(teamID, projectID, sprintID, data)
}

// CreateTask creates a new task
func (c *Client) CreateTask(teamID, projectID, sprintID, name, description, itemTypeID, priorityID string) (*Item, error) {
	data := map[string]string{
		"name":           name,
		"projitemtypeid": itemTypeID,
		"projpriorityid": priorityID,
	}
	if description != "" {
		data["description"] = description
	}
	return c.CreateItem(teamID, projectID, sprintID, data)
}

// UpdateItem updates an item's details
func (c *Client) UpdateItem(teamID, projectID, sprintID, itemID string, updates map[string]string) (*Item, error) {
	var resp ItemResponse
	path := c.GetProjectPath(teamID, projectID) + "/sprints/" + sprintID + "/item/" + itemID + "/"

	data := url.Values{}
	for k, v := range updates {
		data.Set(k, v)
	}

	err := c.PostJSON(path, data, &resp)
	if err != nil {
		return nil, err
	}
	return &resp.Item, nil
}

// MoveItemToSprint moves an item to a different sprint
func (c *Client) MoveItemToSprint(teamID, projectID, currentSprintID, itemID, targetSprintID string) (*Item, error) {
	return c.UpdateItem(teamID, projectID, currentSprintID, itemID, map[string]string{
		"sprint_id": targetSprintID,
	})
}

// AssignItem assigns an item to a user
func (c *Client) AssignItem(teamID, projectID, sprintID, itemID, assigneeID string) (*Item, error) {
	return c.UpdateItem(teamID, projectID, sprintID, itemID, map[string]string{
		"users": "[\"" + assigneeID + "\"]",
	})
}

// UpdateItemStatus updates an item's status
func (c *Client) UpdateItemStatus(teamID, projectID, sprintID, itemID, statusID string) (*Item, error) {
	return c.UpdateItem(teamID, projectID, sprintID, itemID, map[string]string{
		"statusid": statusID,
	})
}

// DeleteItem deletes an item
func (c *Client) DeleteItem(teamID, projectID, sprintID, itemID string) error {
	path := c.GetProjectPath(teamID, projectID) + "/sprints/" + sprintID + "/item/" + itemID + "/"
	_, err := c.Delete(path)
	return err
}

// SearchItems searches for items by name or description within a sprint
func (c *Client) SearchItems(teamID, projectID, sprintID, query string) ([]Item, error) {
	params := url.Values{}
	params.Set("search", query)
	return c.ListItems(teamID, projectID, sprintID, params)
}
