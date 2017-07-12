package main

import (
	"errors"
	"flag"
	"log"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/harshavardhana/github"
	chart "github.com/wcharczuk/go-chart"
	"github.com/wcharczuk/go-chart/drawing"
)

var gh *github.GitHub

// Compare repos
var (
	repo1 string
	repo2 string
)

func init() {
	flag.StringVar(&repo1, "repo1", "minio/minio", "provide custom repo name")
	flag.StringVar(&repo2, "repo2", "ceph/ceph", "provide custom repo name")
	pageSize, err := strconv.Atoi(os.Getenv("GITHUB_PAGE_SIZE"))
	if err != nil {
		pageSize = 100
	}
	gh = &github.GitHub{
		Token:    os.Getenv("GITHUB_TOKEN"),
		PageSize: pageSize,
	}
}

func main() {
	flag.Parse()

	if len(os.Args) == 2 {
		graph, err := generateGraph()
		if err != nil {
			log.Fatalln(err)
		}
		w, err := os.Create(os.Args[1])
		if err != nil {
			log.Fatalln(err)
		}
		defer w.Close()
		isSVG := strings.Contains(mime.TypeByExtension(filepath.Ext(os.Args[1])), "svg")
		isPNG := strings.Contains(mime.TypeByExtension(filepath.Ext(os.Args[1])), "png")
		switch {
		case isSVG:
			err = graph.Render(chart.SVG, w)
		case isPNG:
			err = graph.Render(chart.PNG, w)
		default:
			err = errors.New("Unrecognized file type")
		}
		if err != nil {
			log.Fatalln(err)
		}
		os.Exit(0)
	}
	http.HandleFunc("/", drawChart)
	http.HandleFunc("/favico.ico", func(res http.ResponseWriter, req *http.Request) {
		res.Write([]byte{})
	})
	log.Println("Started listening on :8080, visit http://localhost:8080")
	http.ListenAndServe(":8080", nil)
}

func generateGraph() (graph chart.Chart, err error) {
	r1, err := gh.RepoInfo(repo1)
	if err != nil {
		return graph, err
	}
	r2, err := gh.RepoInfo(repo2)
	if err != nil {
		return graph, err
	}

	st1, err := gh.Stargazers(r1)
	if err != nil {
		return graph, err
	}

	st2, err := gh.Stargazers(r2)
	if err != nil {
		return graph, err
	}

	ts1 := chart.TimeSeries{
		Name: repo1,
		Style: chart.Style{
			Show: true,
			StrokeColor: drawing.Color{
				R: 129,
				G: 199,
				B: 239,
				A: 150,
			},
		},
	}

	ts2 := chart.TimeSeries{
		Name: repo2,
		Style: chart.Style{
			Show:        true,
			StrokeColor: chart.GetDefaultColor(1),
		},
	}

	for i, star := range st1 {
		ts1.XValues = append(ts1.XValues, star.StarredAt)
		ts1.YValues = append(ts1.YValues, float64(i))
	}

	for i, star := range st2 {
		ts2.XValues = append(ts2.XValues, star.StarredAt)
		ts2.YValues = append(ts2.YValues, float64(i))
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
		Series: []chart.Series{ts1, ts2},
	}
	return graph, nil
}

func drawChart(w http.ResponseWriter, req *http.Request) {
	graph, err := generateGraph()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Add("Content-Type", "image/svg+xml")
	graph.Render(chart.SVG, w)
}
