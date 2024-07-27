package main

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/function61/gokit/net/http/ezhttp"
)

type Project struct {
	Name string `json:"name"`
	URL  string `json:"url"`
}

type Task struct {
	Id        string    `json:"id"`
	Order     int       `json:"order"` // order within this project
	Content   string    `json:"content"`
	Completed bool      `json:"is_completed"`
	Created   time.Time `json:"created_at"`
	Url       string    `json:"url"`
	Due       *DueSpec  `json:"due"` // only present for ones that have due date
}

// returns 0 if no due date
// NOTE: returned value can be negative
func (t Task) OverdueAmount(now time.Time) time.Duration {
	if t.Due != nil {
		return t.Due.Overdue(now)
	} else {
		return time.Duration(0)
	}
}

type DueSpec struct {
	Recurring bool          `json:"is_recurring"`
	Date      JSONPlainDate `json:"date"` // looks like: 2021-01-15
}

// NOTE: returned value can be negative
func (d DueSpec) Overdue(now time.Time) time.Duration {
	return now.Sub(d.Date.Time)
}

type Todoist struct {
	token string
}

func (t *Todoist) Project(ctx context.Context, id int64) (*Project, error) {
	project := &Project{}

	if _, err := ezhttp.Get(
		ctx,
		fmt.Sprintf("https://api.todoist.com/rest/v2/projects/%d", id),
		ezhttp.AuthBearer(t.token),
		ezhttp.RespondsJSONAllowUnknownFields(project),
	); err != nil {
		return nil, fmt.Errorf("Project: %w", err)
	}

	return project, nil
}

func (t *Todoist) TasksByProject(ctx context.Context, id int64, now time.Time) ([]Task, error) {
	tasks := []Task{}

	if _, err := ezhttp.Get(
		ctx,
		fmt.Sprintf("https://api.todoist.com/rest/v2/tasks?project_id=%d", id),
		ezhttp.AuthBearer(t.token),
		ezhttp.RespondsJSONAllowUnknownFields(&tasks),
	); err != nil {
		return nil, fmt.Errorf("TasksByProject: %w", err)
	}

	// REST API has no sensible ordering, so we have to sort them.
	// NOTE: this doesn't seem to be the exact same order as Todoist UI uses. i.e. not all overdue
	//       tasks are at the top for some reason..
	sort.Slice(tasks, func(i, j int) bool {
		return multiCompare(
			func() int { // first sort by *overdue* due date (if set)
				leftOverdue := tasks[i].OverdueAmount(now) > 0
				rightOverdue := tasks[j].OverdueAmount(now) > 0

				switch {
				case !leftOverdue && !rightOverdue:
					return 0 // equal
				case leftOverdue && !rightOverdue:
					return -1
				case !leftOverdue && rightOverdue:
					return 1
				default: // both have due date
					// i <-> j in different order purposefully because larger overdue needs to sort before
					return int64Compare(
						int64(tasks[j].OverdueAmount(now)),
						int64(tasks[i].OverdueAmount(now)))
				}
			}(),
			intCompare(tasks[i].Order, tasks[j].Order),
		)
	})

	return tasks, nil
}
