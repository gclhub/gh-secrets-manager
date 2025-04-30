package api

import (
	"fmt"
	"github.com/google/go-github/v45/github"
)

// ListRepositoriesByProperty returns all repositories in an organization that have a specific property value
func (c *Client) ListRepositoriesByProperty(org, propertyName, propertyValue string) ([]*github.Repository, error) {
	var matchingRepos []*github.Repository
	opts := &github.RepositoryListByOrgOptions{
		ListOptions: github.ListOptions{PerPage: 100},
	}

	for {
		repos, resp, err := c.github.Repositories.ListByOrg(c.ctx, org, opts)
		if err != nil {
			return nil, fmt.Errorf("failed to list repositories: %w", err)
		}

		for _, repo := range repos {
			matches, err := c.hasProperty(repo, propertyName, propertyValue)
			if err != nil {
				// Log the error but continue processing other repositories
				fmt.Printf("Warning: Failed to check property for %s: %v\n", repo.GetName(), err)
				continue
			}
			if matches {
				matchingRepos = append(matchingRepos, repo)
			}
		}

		if resp.NextPage == 0 {
			break
		}
		opts.Page = resp.NextPage
	}

	return matchingRepos, nil
}

// hasProperty checks if a repository has the specified property with the given value
func (c *Client) hasProperty(repo *github.Repository, propertyName, propertyValue string) (bool, error) {
	// First check common properties
	switch propertyName {
	case "name":
		return repo.GetName() == propertyValue, nil
	case "description":
		return repo.GetDescription() == propertyValue, nil
	case "language":
		return repo.GetLanguage() == propertyValue, nil
	case "visibility":
		return repo.GetVisibility() == propertyValue, nil
	case "is_private":
		return repo.GetPrivate() == (propertyValue == "true"), nil
	case "has_issues":
		return repo.GetHasIssues() == (propertyValue == "true"), nil
	case "has_wiki":
		return repo.GetHasWiki() == (propertyValue == "true"), nil
	case "archived":
		return repo.GetArchived() == (propertyValue == "true"), nil
	case "disabled":
		return repo.GetDisabled() == (propertyValue == "true"), nil
	case "topic":
		// Make a direct API call to get repository topics
		req, err := c.github.NewRequest("GET", fmt.Sprintf("repos/%s/%s/topics", repo.GetOwner().GetLogin(), repo.GetName()), nil)
		if err != nil {
			return false, fmt.Errorf("failed to create request: %w", err)
		}

		var response struct {
			Names []string `json:"names"`
		}
		_, err = c.github.Do(c.ctx, req, &response)
		if err != nil {
			return false, fmt.Errorf("failed to get topics: %w", err)
		}

		for _, topic := range response.Names {
			if topic == propertyValue {
				return true, nil
			}
		}
		return false, nil
	}

	return false, nil
}