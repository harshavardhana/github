package github

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// RepoInfo get repository info.
type RepoInfo struct {
	FullName        string `json:"full_name"`
	StargazersCount int    `json:"stargazers_count"`
	CreatedAt       string `json:"created_at"`
}

// GetRepoInfo gets the given repository info.
func (gh *GitHub) GetRepoInfo(name string) (repo RepoInfo, err error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/repos/%s", githubAPIURL, name), nil)
	if err != nil {
		return
	}
	// Set token if provided.
	if gh.token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("token %s", gh.token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&repo)
	return
}
