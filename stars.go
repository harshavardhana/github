package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
)

var errNoMorePages = errors.New("no more pages to get")

// Stargazer is a star on a project at a given time
type Stargazer struct {
	StarredAt time.Time `json:"starred_at"`
	User      struct {
		Name string `json:"login"`
		Type string `json:"type"`
	} `json:"user"`
}

// byStargazers is a collection satisfying sort.Interface.
type byStargazers []Stargazer

func (d byStargazers) Len() int           { return len(d) }
func (d byStargazers) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d byStargazers) Less(i, j int) bool { return d[i].StarredAt.Before(d[j].StarredAt) }

// Stargazers returns all the stargazers of a given repo
// the date and time when they were starred at. Additionally
// also provides user name as well.
func (gh *GitHub) Stargazers(repo RepoInfo) (stars []Stargazer, err error) {
	sem := make(chan bool, 10)
	var g errgroup.Group
	var lock sync.Mutex
	for page := 1; page <= gh.lastPage(repo); page++ {
		sem <- true
		page := page
		g.Go(func() error {
			defer func() { <-sem }()
			result, err := gh.getStargazersPage(repo, page)
			if err == errNoMorePages {
				return nil
			}
			if err != nil {
				return err
			}
			lock.Lock()
			defer lock.Unlock()
			stars = append(stars, result...)
			return nil
		})
	}
	err = g.Wait()
	sort.Sort(byStargazers(stars))
	return
}

func (gh *GitHub) getStargazersPage(repo RepoInfo, page int) (stars []Stargazer, err error) {
	var url = fmt.Sprintf(
		"%s/repos/%s/stargazers?page=%d&per_page=%d",
		githubAPIURL,
		repo.FullName,
		page,
		gh.PageSize,
	)
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return stars, err
	}
	req.Header.Add("Accept", "application/vnd.github.v3.star+json")
	if gh.Token != "" {
		req.Header.Add("Authorization", fmt.Sprintf("Token %s", gh.Token))
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return stars, err
	}
	defer resp.Body.Close()

	// Rate limit hit, wait for 10secs and try again.
	if resp.StatusCode == http.StatusForbidden {
		log.Println("Rate limit hit, waiting 10s before trying again.")
		time.Sleep(10 * time.Second)
		return gh.getStargazersPage(repo, page)
	}
	if resp.StatusCode != http.StatusOK {
		bts, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return stars, err
		}
		return stars, fmt.Errorf("failed to get stargazers from github api: %v", string(bts))
	}
	err = json.NewDecoder(resp.Body).Decode(&stars)
	if len(stars) == 0 {
		return stars, errNoMorePages
	}
	return
}

func (gh *GitHub) lastPage(repo RepoInfo) int {
	return (repo.StargazersCount / gh.PageSize) + 1
}
