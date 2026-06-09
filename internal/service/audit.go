package service

import "taskboard-api/internal/audit"

type auditPublisher interface {
	Publish(event audit.Event)
}
