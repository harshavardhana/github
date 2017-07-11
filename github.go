package github

// GitHub client struct
type GitHub struct {
	token    string
	pageSize int
}

const githubAPIURL = "https://api.github.com"

// New github client
func New(token string, pageSize int) *GitHub {
	if pageSize <= 0 {
		pageSize = 100 // Default page size is 100.
	}
	return &GitHub{
		token:    token,
		pageSize: pageSize,
	}
}
