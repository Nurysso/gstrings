package main

import (
	"encoding/json"
	"os"
	"time"
)

//  Domain model

type Priority string

const (
	PriorityHigh   Priority = "high"
	PriorityMedium Priority = "medium"
	PriorityLow    Priority = "low"
)

type Status string

const (
	StatusTodo       Status = "todo"
	StatusInProgress Status = "in-progress"
	StatusDone       Status = "done"
)

type Task struct {
	ID        int      `json:"id"`
	Title     string   `json:"title"`
	Project   string   `json:"project"`
	Priority  Priority `json:"priority"`
	Status    Status   `json:"status"`
	DueDate   string   `json:"due_date"` // "YYYY-MM-DD" or ""
	CreatedAt string   `json:"created_at"`
	Notes     string   `json:"notes"`
}

//  Store

type Store struct {
	Tasks  []*Task `json:"tasks"`
	NextID int     `json:"next_id"`
	file   string
}

func LoadStore(path string) (*Store, error) {
	s := &Store{file: path, NextID: 1}
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return s, nil
	}
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(data, s); err != nil {
		return nil, err
	}
	return s, nil
}

func (s *Store) Save() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.file, data, 0644)
}

func (s *Store) Add(title, project string, priority Priority, dueDate, notes string) *Task {
	t := &Task{
		ID:        s.NextID,
		Title:     title,
		Project:   project,
		Priority:  priority,
		Status:    StatusTodo,
		DueDate:   dueDate,
		CreatedAt: time.Now().Format("2006-01-02"),
		Notes:     notes,
	}
	s.Tasks = append(s.Tasks, t)
	s.NextID++
	return t
}

func (s *Store) Find(id int) *Task {
	for _, t := range s.Tasks {
		if t.ID == id {
			return t
		}
	}
	return nil
}

func (s *Store) Delete(id int) bool {
	for i, t := range s.Tasks {
		if t.ID == id {
			s.Tasks = append(s.Tasks[:i], s.Tasks[i+1:]...)
			return true
		}
	}
	return false
}

func (s *Store) FilterByStatus(status Status) []*Task {
	var out []*Task
	for _, t := range s.Tasks {
		if t.Status == status {
			out = append(out, t)
		}
	}
	return out
}

func (s *Store) FilterByProject(project string) []*Task {
	var out []*Task
	for _, t := range s.Tasks {
		if t.Project == project {
			out = append(out, t)
		}
	}
	return out
}

func (s *Store) Projects() []string {
	seen := map[string]bool{}
	var projects []string
	for _, t := range s.Tasks {
		if t.Project != "" && !seen[t.Project] {
			seen[t.Project] = true
			projects = append(projects, t.Project)
		}
	}
	return projects
}

func (s *Store) Stats() (todo, inProgress, done int) {
	for _, t := range s.Tasks {
		switch t.Status {
		case StatusTodo:
			todo++
		case StatusInProgress:
			inProgress++
		case StatusDone:
			done++
		}
	}
	return
}
