package main

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/function61/gokit/app/aws/lambdautils"
	"github.com/function61/gokit/net/http/httputils"
	"github.com/function61/gokit/os/osutil"
	"github.com/gorilla/feeds"
	"github.com/gorilla/mux"
)

func main() {
	if lambdautils.InLambda() {
		lambda.StartHandler(lambdautils.NewLambdaHttpHandlerAdapter(
			makeHandler()))
		return
	}

	osutil.ExitIfError(
		logic(osutil.CancelOnInterruptOrTerminate(nil)))
}

func logic(ctx context.Context) error {
	srv := &http.Server{
		Addr:    ":80",
		Handler: makeHandler(),
	}

	return httputils.CancelableServer(ctx, srv, func() error { return srv.ListenAndServe() })
}

func makeHandler() http.Handler {
	routes := mux.NewRouter()

	routes.HandleFunc("/todoist-to-rss/api/project/{project}/tasks.xml", func(w http.ResponseWriter, r *http.Request) {
		projectId, err := func() (int64, error) {
			num, err := strconv.Atoi(mux.Vars(r)["project"])
			return int64(num), err
		}()
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		token := r.URL.Query().Get("token")
		if token == "" {
			http.Error(w, "need token", http.StatusBadRequest)
			return
		}

		todoistForUser := Todoist{token}

		// need to fetch project metadata for naming RSS feed
		project, err := todoistForUser.Project(r.Context(), projectId)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		now := time.Now()

		tasks, err := todoistForUser.TasksByProject(r.Context(), projectId, now)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		feed, err := tasksToRSS(tasks, *project, now).ToRss()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/rss+xml")
		_, _ = w.Write([]byte(feed))
	})

	return routes
}

func tasksToRSS(tasks []Task, project Project, now time.Time) *feeds.Feed {
	rssItems := []*feeds.Item{}

	for _, task := range tasks {
		titleWithAlarm := func() string {
			overdueAmount := task.OverdueAmount(now)

			if overdueAmount > 0 {
				return fmt.Sprintf("🚨 (OVER=%dd) %s", durationDays(overdueAmount), task.Content)
			} else {
				return task.Content
			}
		}()

		rssItems = append(rssItems, &feeds.Item{
			Id: intToGuid(task.Id),
			Title: func() string {
				if task.Completed {
					return fmt.Sprintf("✅ %s", titleWithAlarm)
				} else {
					return titleWithAlarm
				}
			}(),
			Link:    &feeds.Link{Href: task.Url},
			Created: task.Created,
		})
	}

	return &feeds.Feed{
		Title:       project.Name,
		Description: fmt.Sprintf("Todoist tasks for project %s", project.Name),
		Items:       rssItems,

		// crashes without this. is it required per RSS spec?
		Link: &feeds.Link{Href: project.URL},

		Created: time.Now().UTC(),
	}
}

func durationDays(dur time.Duration) int {
	return int(dur.Hours()) / 24 // naive approach
}
