package api

import (
	"fmt"

	"github.com/google/go-github/v45/github"
)

// ListRepositoriesByProperty returns all repositories in an organization that have a specific custom property value
func (c *Client) ListRepositoriesByProperty(org, propertyName, propertyValue string) ([]*github.Repository, error) {
	if err := c.ensureValidToken(); err != nil {
		return nil, err
	}

	// Validate required parameters
	if org == "" {
		return nil, fmt.Errorf("organization name cannot be empty")
	}
	if propertyName == "" {
		return nil, fmt.Errorf("property_name cannot be empty")
	}
	if propertyValue == "" {
		return nil, fmt.Errorf("property value cannot be empty")
	}

	var matchingRepos []*github.Repository
	page := 1

	// Use the custom properties API to get repositories with the specific property value
	for {
		url := fmt.Sprintf("orgs/%s/properties/values?property_name=%s&value=%s&page=%d&per_page=100",
			org, propertyName, propertyValue, page)

		req, err := c.github.NewRequest("GET", url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		var response struct {
			Repositories []struct {
				Name string `json:"name"`
			} `json:"repositories"`
			HasNextPage bool `json:"has_next_page"`
		}
		_, err = c.github.Do(c.ctx, req, &response)
		if err != nil {
			return nil, fmt.Errorf("failed to get repositories by property: %w", err)
		}

		// Get full repository objects for each matching repository
		for _, repoInfo := range response.Repositories {
			repo, _, err := c.github.Repositories.Get(c.ctx, org, repoInfo.Name)
			if err != nil {
				return nil, fmt.Errorf("failed to get repository %s: %w", repoInfo.Name, err)
			}
			matchingRepos = append(matchingRepos, repo)
		}

		if !response.HasNextPage {
			break
		}
		page++
	}

	return matchingRepos, nil
}
