package github

// GitHub client struct
type GitHub struct {
	Token    string
	PageSize int
}

const githubAPIURL = "https://api.github.com"
