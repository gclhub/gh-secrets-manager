package auth

var githubAPIBaseURL = "https://api.github.com"

// SetGitHubAPIBaseURL allows overriding the GitHub API base URL for testing
func SetGitHubAPIBaseURL(url string) {
	githubAPIBaseURL = url
}

// GetGitHubAPIBaseURL returns the current GitHub API base URL
func GetGitHubAPIBaseURL() string {
	return githubAPIBaseURL
}
