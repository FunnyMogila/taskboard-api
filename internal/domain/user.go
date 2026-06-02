package domain

import "time"

type UserID int64

type User struct {
	ID        UserID
	Name      string
	Email     string
	CreatedAt time.Time
}
