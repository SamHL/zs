package api

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
)

// ItemType represents an item type in Zoho Sprints (Story, Bug, Task, etc.)
type ItemType struct {
	ID   string `json:"proj_itemtype_id"`
	Name string `json:"proj_itemtype_name"`
}

// Priority represents a priority level in Zoho Sprints
type Priority struct {
	ID   string `json:"proj_priority_id"`
	Name string `json:"proj_priority_name"`
}

// ItemTypesRawResponse represents the raw API response for item types
type ItemTypesRawResponse struct {
	Status       string                   `json:"status"`
	ItemTypeJObj map[string][]interface{} `json:"projItemTypeJObj"`
	ItemTypeIDs  []string                 `json:"projItemTypeIds"`
}

// PrioritiesRawResponse represents the raw API response for priorities
type PrioritiesRawResponse struct {
	Status       string                   `json:"status"`
	PriorityJObj map[string][]interface{} `json:"projPriorityJObj"`
	PriorityIDs  []string                 `json:"projPriorityIds"`
}

// ListItemTypes retrieves all item types for a project
func (c *Client) ListItemTypes(teamID, projectID string) ([]ItemType, error) {
	path := c.GetProjectPath(teamID, projectID) + "/itemtype/"
	params := url.Values{}
	params.Set("action", "data")

	data, err := c.Get(path, params)
	if err != nil {
		return nil, err
	}

	var rawResp ItemTypesRawResponse
	if err := json.Unmarshal(data, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to parse item types response: %w", err)
	}

	// Array indices from projItemType_prop:
	// 0=baseTypeId, 1=itemTypeName, 2=isDefault, 3=sequence, 4=baseType, 5=prefix,
	// 6=itemTypeImage, 7=itemTypeDescription, 8=createdBy, 9=createdTime
	var itemTypes []ItemType
	for _, id := range rawResp.ItemTypeIDs {
		if arr, ok := rawResp.ItemTypeJObj[id]; ok && len(arr) >= 2 {
			itemTypes = append(itemTypes, ItemType{
				ID:   id,
				Name: toString(arr[1]), // itemTypeName is at index 1
			})
		}
	}

	return itemTypes, nil
}

// ListPriorities retrieves all priorities for a project
func (c *Client) ListPriorities(teamID, projectID string) ([]Priority, error) {
	path := c.GetProjectPath(teamID, projectID) + "/priority/"
	params := url.Values{}
	params.Set("action", "data")

	data, err := c.Get(path, params)
	if err != nil {
		return nil, err
	}

	var rawResp PrioritiesRawResponse
	if err := json.Unmarshal(data, &rawResp); err != nil {
		return nil, fmt.Errorf("failed to parse priorities response: %w", err)
	}

	// Array indices from projPriority_prop:
	// 0=priorityName, 1=isDefault, 2=priorityId, 3=priorityDescription, 4=colorCode, 5=sequence
	var priorities []Priority
	for _, id := range rawResp.PriorityIDs {
		if arr, ok := rawResp.PriorityJObj[id]; ok && len(arr) >= 1 {
			priorities = append(priorities, Priority{
				ID:   id,
				Name: toString(arr[0]), // priorityName is at index 0
			})
		}
	}

	return priorities, nil
}

// ResolveItemType finds an item type ID by name (case-insensitive)
func (c *Client) ResolveItemType(teamID, projectID, name string) (string, error) {
	itemTypes, err := c.ListItemTypes(teamID, projectID)
	if err != nil {
		return "", err
	}

	nameLower := strings.ToLower(name)
	for _, it := range itemTypes {
		if strings.ToLower(it.Name) == nameLower {
			return it.ID, nil
		}
	}

	// Build list of valid types for error message
	var validTypes []string
	for _, it := range itemTypes {
		validTypes = append(validTypes, it.Name)
	}
	return "", fmt.Errorf("unknown item type '%s', valid types: %s", name, strings.Join(validTypes, ", "))
}

// ResolvePriority finds a priority ID by name (case-insensitive)
func (c *Client) ResolvePriority(teamID, projectID, name string) (string, error) {
	priorities, err := c.ListPriorities(teamID, projectID)
	if err != nil {
		return "", err
	}

	nameLower := strings.ToLower(name)
	for _, p := range priorities {
		if strings.ToLower(p.Name) == nameLower {
			return p.ID, nil
		}
	}

	// Build list of valid priorities for error message
	var validPriorities []string
	for _, p := range priorities {
		validPriorities = append(validPriorities, p.Name)
	}
	return "", fmt.Errorf("unknown priority '%s', valid priorities: %s", name, strings.Join(validPriorities, ", "))
}
