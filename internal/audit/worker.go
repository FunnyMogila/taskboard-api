package audit

import (
	"context"
	"log"
	"sync"
)

type Event struct {
	Type    string
	Payload string
}

type Worker struct {
	events chan Event

	mu   sync.Mutex
	logs []Event
}

func NewWorker(bufferSize int) *Worker {
	return &Worker{
		events: make(chan Event, bufferSize),
		logs:   make([]Event, 0),
	}
}

func (w *Worker) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case event := <-w.events:
			w.handle(event)
		}
	}
}

func (w *Worker) Publish(event Event) {
	select {
	case w.events <- event:
	default:
		log.Printf("audit event dropped: %s", event.Type)
	}
}

func (w *Worker) handle(event Event) {
	w.mu.Lock()
	w.logs = append(w.logs, event)
	w.mu.Unlock()

	log.Printf("audit event: type=%s payload=%s", event.Type, event.Payload)
}

func (w *Worker) Logs() []Event {
	w.mu.Lock()
	defer w.mu.Unlock()

	result := make([]Event, len(w.logs))
	copy(result, w.logs)

	return result
}
