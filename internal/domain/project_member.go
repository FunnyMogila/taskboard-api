package domain

import "time"

type ProjectRole string

const (
	ProjectRoleOwner  ProjectRole = "owner"
	ProjectRoleMember ProjectRole = "member"
)

type ProjectMember struct {
	ProjectID ProjectID
	UserID    UserID
	Role      ProjectRole
	CreatedAt time.Time
}
