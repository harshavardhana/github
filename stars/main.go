package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/harshavardhana/github"
	chart "github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

var gh *github.GitHub

// Compare repos
var repos string
var mode string

var defaultProjects = []string{
	"minio/minio",
	"mongodb/mongo",
	"kubernetes/kubernetes",
	"apache/cassandra",
	"apache/kafka",
	"cockroachdb/cockroach",
	"elastic/elasticsearch",
}

func init() {
	flag.StringVar(&repos, "repos", strings.Join(defaultProjects, ","), "provide list of repos to compare with each other")
	flag.StringVar(&mode, "mode", "info", "prints only higher level info for all repos. Supported modes are [info,file,service]")

	gh = &github.GitHub{
		Token:    os.Getenv("GITHUB_TOKEN"),
		PageSize: 100, // Any other value doesn't work well.
	}
}

// byStargazerCount is a collection satisfying sort.Interface.
type byStargazerCount []github.RepoInfo

func (d byStargazerCount) Len() int           { return len(d) }
func (d byStargazerCount) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }
func (d byStargazerCount) Less(i, j int) bool { return d[i].StargazersCount > d[j].StargazersCount }

func getRepoInfos() (rs []github.RepoInfo, err error) {
	for _, repo := range strings.Split(repos, ",") {
		r1, err := gh.RepoInfo(repo)
		if err != nil {
			return nil, err
		}
		rs = append(rs, r1)
	}

	sort.Sort(byStargazerCount(rs))
	return rs, nil
}

func printRepoInfos() error {
	repoInfos, err := getRepoInfos()
	if err != nil {
		return err
	}

	var maxRepoName = 0
	for _, r := range repoInfos {
		if len(r.FullName) > maxRepoName {
			maxRepoName = len(r.FullName)
		}
	}

	for _, r := range repoInfos {
		fullName := fmt.Sprintf("%-*.*s :", maxRepoName, maxRepoName, r.FullName)
		fmt.Printf("%s %d\n", fullName, r.StargazersCount)
	}

	return nil
}

// Saves repo comparison chart locally as a file depending on the
// file type, currently only supports SVG and PNG.
func saveRepoComparison(filename string) error {
	repoInfos, err := getRepoInfos()
	if err != nil {
		return err
	}
	graph, err := generateGraph(repoInfos)
	if err != nil {
		return err
	}
	w, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer w.Close()
	isSVG := strings.Contains(mime.TypeByExtension(filepath.Ext(filename)), "svg")
	isPNG := strings.Contains(mime.TypeByExtension(filepath.Ext(filename)), "png")
	switch {
	case isSVG:
		err = graph.Render(chart.SVG, w)
	case isPNG:
		err = graph.Render(chart.PNG, w)
	default:
		err = errors.New("Unrecognized file type")
	}
	return err
}

// Starts a http service to view the comparison chart on a browser.
func startService() error {
	http.HandleFunc("/output.svg", drawChart)
	http.HandleFunc("/favico.ico", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte{})
	})
	log.Println("Started listening on :8080, visit http://localhost:8080/output.svg")
	return http.ListenAndServe(":8080", nil)
}

func main() {
	flag.Parse()

	var err error
	switch mode {
	case "info":
		err = printRepoInfos()
	case "file":
		err = saveRepoComparison(flag.Arg(0))
	case "service":
		err = startService()
	default:
		err = fmt.Errorf("Unknown mode requested mode:(%s)", mode)
	}
	if err != nil {
		log.Fatalln(err)
	}
}

// Mixing random colors with white (255, 255, 255) creates neutral
// pastels by increasing the lightness while keeping the hue of the
// original color. These randomly generated pastels usually go well
// together, especially in large numbers.
func generateRandomColor(mix *drawing.Color) drawing.Color {
	red := rand.Intn(256)
	green := rand.Intn(256)
	blue := rand.Intn(256)

	if mix != nil {
		red = (red + int(mix.R)) / 2
		green = (green + int(mix.G)) / 2
		blue = (blue + int(mix.B)) / 2
	}

	return drawing.Color{
		R: uint8(red),
		G: uint8(green),
		B: uint8(blue),
		A: 150,
	}
}

func generateGraph(rs []github.RepoInfo) (graph chart.Chart, err error) {
	var chartSeries = make([][]github.Stargazer, len(rs))
	var wg sync.WaitGroup
	for i, r := range rs {
		wg.Add(1)
		go func(i int, r github.RepoInfo) {
			defer wg.Done()
			st, rerr := gh.Stargazers(r)
			if rerr != nil {
				fmt.Println(rerr, r)
				return
			}
			chartSeries[i] = st
		}(i, r)
	}
	wg.Wait()

	var timeSeriesList []chart.Series
	for i, st := range chartSeries {
		timeSeries := chart.TimeSeries{
			Name: rs[i].FullName,
			Style: chart.Style{
				Show:        true,
				StrokeColor: generateRandomColor(nil),
			},
		}
		t := time.Date(2013, time.January, 10, 23, 0, 0, 0, time.UTC)
		for j, star := range st {
			if star.StarredAt.After(t) {
				timeSeries.XValues = append(timeSeries.XValues, star.StarredAt)
				timeSeries.YValues = append(timeSeries.YValues, float64(j))
			}
		}
		timeSeriesList = append(timeSeriesList, timeSeries)
	}

	graph = chart.Chart{
		XAxis: chart.XAxis{
			Name:      "Time",
			NameStyle: chart.StyleShow(),
			Style: chart.Style{
				Show:        true,
				StrokeWidth: 1,
				StrokeColor: drawing.Color{
					R: 85,
					G: 85,
					B: 85,
					A: 180,
				},
			},
		},
		YAxis: chart.YAxis{
			Name:      "Stargazers",
			NameStyle: chart.StyleShow(),
			Style: chart.Style{
				Show:        true,
				StrokeWidth: 1,
				StrokeColor: drawing.Color{
					R: 85,
					G: 85,
					B: 85,
					A: 180,
				},
			},
		},
		Series: timeSeriesList,
	}
	graph.Elements = []chart.Renderable{
		chart.Legend(&graph),
	}
	return graph, nil
}

func drawChart(w http.ResponseWriter, req *http.Request) {
	repoInfos, err := getRepoInfos()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	graph, err := generateGraph(repoInfos)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "image/svg+xml")
	graph.Render(chart.SVG, w)
}
