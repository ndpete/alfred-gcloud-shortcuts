package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"strings"

	aw "github.com/deanishe/awgo"
	"github.com/deanishe/awgo/update"
	"go.deanishe.net/fuzzy"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/cloudresourcemanager/v1"
	"google.golang.org/api/option"
)

const (
	cacheGoogleProjects = "google-projects"
	scopeListProjects   = "https://www.googleapis.com/auth/cloudplatformprojects.readonly"
	repoOwnerName       = "ndpete/alfred-gcloud-shortcuts"
	updateJobName       = "checkForUpdate"
)

var (
	logger = log.New(os.Stderr, "[projects] ", log.LstdFlags)
	wf     *aw.Workflow

	argRefreshProjects bool
	argCheckUpdate     bool
)

type ProjectDescription struct {
	Name   string `json:"name"`
	ID     string `json:"id"`
	Number int64  `json:"number"`
}

func init() {
	sopts := []fuzzy.Option{
		fuzzy.AdjacencyBonus(10.0),
		fuzzy.LeadingLetterPenalty(-0.1),
		fuzzy.MaxLeadingLetterPenalty(-3.0),
		fuzzy.UnmatchedLetterPenalty(-0.5),
	}
	wf = aw.New(
		aw.SortOptions(sopts...),
		update.GitHub(repoOwnerName),
	)

	flag.CommandLine.Init(os.Args[0], flag.ContinueOnError)
	flag.CommandLine.SetOutput(io.Discard)
	flag.BoolVar(&argRefreshProjects, "refresh", false, "refresh authenticated projects")
	flag.BoolVar(&argCheckUpdate, "check", false, "check for workflow updates")
}

func FetchGoogleProjects(ctx context.Context) ([]ProjectDescription, error) {
	goog, err := google.DefaultClient(ctx, scopeListProjects)
	if err != nil {
		return nil, fmt.Errorf("failed to create google client: %w", err)
	}

	s, err := cloudresourcemanager.NewService(ctx, option.WithHTTPClient(goog))
	if err != nil {
		return nil, fmt.Errorf("google resource manager: %w", err)
	}

	allProjects := make([]ProjectDescription, 0)

	nextPageToken := ""
	for {
		res, err := s.Projects.List().Context(ctx).PageToken(nextPageToken).PageSize(200).Do()
		if err != nil {
			return nil, err
		}
		nextPageToken = res.NextPageToken
		for _, project := range res.Projects {
			allProjects = append(allProjects, ProjectDescription{
				Name:   project.Name,
				ID:     project.ProjectId,
				Number: project.ProjectNumber,
			})
		}
		if nextPageToken == "" {
			break
		}
	}

	return allProjects, nil
}

func run() {
	args := wf.Args()
	_ = flag.CommandLine.Parse(os.Args[1:])
	ctx := context.Background()

	if argCheckUpdate {
		wf.Configure(aw.TextErrors(true))
		logger.Printf("checking for updates...")
		if err := wf.CheckForUpdate(); err != nil {
			wf.FatalError(err)
		}
		return
	}

	if argRefreshProjects {
		wf.Configure(aw.TextErrors(true))
		logger.Printf("refreshing projects")
		projects, err := FetchGoogleProjects(ctx)
		if err != nil {
			wf.FatalError(err)
		}
		err = wf.Data.StoreJSON(cacheGoogleProjects, projects)
		if err != nil {
			wf.FatalError(err)
		}
		logger.Printf("refresh done")
		wf.SendFeedback()
		return
	}

	// Trigger background check if due (> 24h) and not already running
	if wf.UpdateCheckDue() && !wf.IsRunning(updateJobName) {
		cmd := exec.Command(os.Args[0], "-check")
		_ = wf.RunInBackground(updateJobName, cmd)
	}

	var query string
	if len(args) > 0 {
		query = strings.TrimSpace(args[0])
	}

	qLower := strings.ToLower(query)
	if strings.HasPrefix(qLower, "-") ||
		strings.HasPrefix(qLower, "update") ||
		strings.HasPrefix(qLower, "version") ||
		qLower == "refresh" {

		wf.NewItem("Refresh projects").
			Subtitle("Update cached GCP projects").
			Arg("-refresh").Autocomplete("-refresh").Valid(false)

		updateTitle := "Check for updates"
		updateSub := "Check GitHub for new workflow releases"
		if wf.UpdateAvailable() {
			updateTitle = "🚀 Update available!"
			updateSub = "Press ⏎ to open release on GitHub"
		}
		wf.NewItem(updateTitle).
			Subtitle(updateSub).
			Arg(fmt.Sprintf("https://github.com/%s/releases/latest", repoOwnerName)).
			Autocomplete("-update").Valid(true)

		wf.NewItem(fmt.Sprintf("Workflow version: %s", wf.Version())).
			Subtitle("Google Cloud Shortcuts (forked by Nathan Peterson)").
			Arg(fmt.Sprintf("https://github.com/%s", repoOwnerName)).
			Autocomplete("-version").Valid(true)

		if qLower != "-" && qLower != "" {
			wf.Filter(query)
		}

		wf.SendFeedback()
		return
	}

	logger.Printf("query=%s", query)
	var projects []ProjectDescription
	if !wf.Data.Exists(cacheGoogleProjects) {
		wf.Fatal(`No projects cached, please run "-refresh"`)
	}
	if err := wf.Data.LoadJSON(cacheGoogleProjects, &projects); err != nil {
		wf.FatalError(err)
	}

	// Show update banner at top of empty search when update is available
	if query == "" && wf.UpdateAvailable() {
		wf.Configure(aw.SuppressUIDs(true))
		wf.NewItem("🚀 Update available!").
			Subtitle("Press ⏎ to open release on GitHub").
			Arg(fmt.Sprintf("https://github.com/%s/releases/latest", repoOwnerName)).
			Valid(true)
	}

	for _, p := range projects {
		wf.NewItem(p.Name).Arg(p.ID).Subtitle(p.ID).UID(p.ID).Valid(true)
	}

	if query != "" {
		wf.Filter(query)
	}

	wf.SendFeedback()
}

func main() {
	wf.Run(run)
}
