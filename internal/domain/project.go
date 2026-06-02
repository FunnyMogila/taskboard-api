package domain

import "time"

type ProjectID int64

type ProjectStatus string

const (
	ProjectStatusActive ProjectStatus = "active"
	ProjectStatusClosed ProjectStatus = "closed"
)

type Project struct {
	ID          ProjectID
	Name        string
	Description string
	Status      ProjectStatus
	CreatedAt   time.Time
}
