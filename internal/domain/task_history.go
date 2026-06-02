package domain

import "time"

type TaskHistoryID int64

type TaskHistory struct {
	ID        TaskHistoryID
	TaskID    TaskID
	OldStatus TaskStatus
	NewStatus TaskStatus
	ChangedAt time.Time
}
