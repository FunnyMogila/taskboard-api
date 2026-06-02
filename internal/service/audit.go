package service

import "taskboard-api/internal/audit"

type AuditPublisher interface {
	Publish(event audit.Event)
}
