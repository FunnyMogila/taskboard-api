package domain

import "time"

type CommentID int64

type Comment struct {
	ID        CommentID
	TaskID    TaskID
	AuthorID  UserID
	Text      string
	CreatedAt time.Time
}
