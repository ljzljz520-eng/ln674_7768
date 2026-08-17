package retry

import (
	"sort"
	"time"

	"campus-device-hub/internal/domain"
)

type SchedulePlan struct {
	Now      time.Time
	Attempts int
	Policy   Policy
}

func (p SchedulePlan) NextTime() time.Time {
	return p.Now.Add(p.Policy.Delay(p.Attempts + 1))
}

func (p SchedulePlan) IsTerminal() bool {
	return p.Attempts >= p.Policy.MaxAttempts
}

func (p SchedulePlan) Build(task domain.RetryTask, message string) domain.RetryTask {
	result := task
	p.Policy.Apply(&result, p.Now, message)
	return result
}

func SortTasks(tasks []domain.RetryTask) []domain.RetryTask {
	result := append([]domain.RetryTask(nil), tasks...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].NextAt.Equal(result[j].NextAt) {
			return result[i].ID < result[j].ID
		}
		return result[i].NextAt.Before(result[j].NextAt)
	})
	return result
}

func DueTasks(tasks []domain.RetryTask, now time.Time) []domain.RetryTask {
	due := make([]domain.RetryTask, 0)
	for _, task := range tasks {
		if task.IsDue(now) {
			due = append(due, task)
		}
	}
	return SortTasks(due)
}

func Summarize(tasks []domain.RetryTask) (pending, resolved int) {
	for _, task := range tasks {
		if task.Resolved {
			resolved++
		} else {
			pending++
		}
	}
	return pending, resolved
}
