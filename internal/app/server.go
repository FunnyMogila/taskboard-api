package app

import (
	"context"
	"net/http"
	"taskboard-api/internal/audit"
	"time"

	"taskboard-api/internal/config"
	"taskboard-api/internal/db"
	"taskboard-api/internal/handler"
	"taskboard-api/internal/repository"
	"taskboard-api/internal/service"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"golang.org/x/sync/errgroup"
)

type Server struct {
	router chi.Router
	db     *pgxpool.Pool
	cancel context.CancelFunc
	group  *errgroup.Group
}

func NewServer(cfg config.Config) *Server {
	ctx, cancel := context.WithCancel(context.Background())
	group, ctx := errgroup.WithContext(ctx)

	pool, err := db.NewPostgresPool(ctx, cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}

	auditWorker := audit.NewWorker(100)

	group.Go(func() error {
		return auditWorker.Run(ctx)
	})

	auditWorker.Publish(audit.Event{
		Type:    "server_started",
		Payload: "taskboard api started",
	})

	userRepository := repository.NewUserRepository(pool)
	projectRepository := repository.NewProjectRepository(pool)
	taskRepository := repository.NewTaskRepository(pool)
	commentRepository := repository.NewCommentRepository(pool)

	userService := service.NewUserService(userRepository, auditWorker)
	projectService := service.NewProjectService(projectRepository, auditWorker)
	taskService := service.NewTaskService(
		taskRepository,
		commentRepository,
		projectRepository,
		auditWorker,
	)

	userHandler := handler.NewUserHandler(userService)
	projectHandler := handler.NewProjectHandler(projectService)
	taskHandler := handler.NewTaskHandler(taskService)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(jsonContentType)

	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api/v1", func(r chi.Router) {
		r.Post("/users", userHandler.Create)
		r.Get("/users/{userID}", userHandler.GetByID)

		r.Post("/projects", projectHandler.Create)
		r.Get("/projects", projectHandler.List)
		r.Get("/projects/{projectID}", projectHandler.GetByID)
		r.Post("/projects/{projectID}/members", projectHandler.AddMember)
		r.Patch("/projects/{projectID}/close", projectHandler.Close)

		r.Post("/tasks", taskHandler.Create)
		r.Get("/tasks", taskHandler.List)
		r.Get("/tasks/{taskID}", taskHandler.GetByID)
		r.Patch("/tasks/{taskID}/status", taskHandler.UpdateStatus)
		r.Delete("/tasks/{taskID}", taskHandler.Delete)

		r.Post("/tasks/{taskID}/comments", taskHandler.AddComment)
		r.Get("/tasks/{taskID}/comments", taskHandler.ListComments)
	})

	return &Server{
		router: r,
		db:     pool,
		cancel: cancel,
		group:  group,
	}
}

func (s *Server) Router() http.Handler {
	return s.router
}

func jsonContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Content-Type-Options", "nosniff")
		next.ServeHTTP(w, r)
	})
}

func (s *Server) Close() {
	if s.cancel != nil {
		s.cancel()
	}

	if s.group != nil {
		_ = s.group.Wait()
	}

	if s.db != nil {
		s.db.Close()
	}
}
