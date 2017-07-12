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

// RepoInfo gets the given repository info.
func (gh *GitHub) RepoInfo(name string) (repo RepoInfo, err error) {
	req, err := http.NewRequest(http.MethodGet, fmt.Sprintf("%s/repos/%s", githubAPIURL, name), nil)
	if err != nil {
		return
	}
	// Set Token if provided.
	if gh.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", gh.Token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(&repo)
	return
}
