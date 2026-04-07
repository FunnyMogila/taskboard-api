package app

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"example.com/go-master-web-sample/internal/auth"
	"example.com/go-master-web-sample/internal/config"
	"example.com/go-master-web-sample/internal/handlers"
	appmw "example.com/go-master-web-sample/internal/middleware"
	"example.com/go-master-web-sample/internal/store"
)

type Server struct {
	router chi.Router
}

func NewServer(cfg config.Config) *Server {
	store := store.NewMemoryStore()
	authManager := auth.NewManager(cfg.JWTSecret)
	handler := handlers.New(store, authManager)

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(chimw.RealIP)
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.Timeout(30 * time.Second))
	r.Use(jsonContentType)

	fileServer := http.StripPrefix("/static/", http.FileServer(http.Dir("static")))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/static/", http.StatusTemporaryRedirect)
	})
	r.Handle("/static/*", fileServer)

	r.Get("/health", handler.Health)
	r.Get("/ready", handler.Ready)
	r.Post("/login", handler.Login)
	r.Get("/request-info", handler.RequestInfo)
	r.Post("/submit-form", handler.EchoForm)

	r.Route("/api/v1", func(r chi.Router) {
		r.Route("/items", func(r chi.Router) {
			r.Get("/", handler.ListItems)
			r.Get("/{itemID}", handler.GetItem)

			r.Group(func(r chi.Router) {
				r.Use(appmw.RequireAuth(authManager))
				r.Post("/", handler.CreateItem)
				r.Put("/{itemID}", handler.ReplaceItem)
				r.Patch("/{itemID}", handler.PatchItem)
				r.Delete("/{itemID}", handler.DeleteItem)
			})
		})
	})

	r.Group(func(r chi.Router) {
		r.Use(appmw.RequireAuth(authManager))
		r.Get("/me", handler.Me)
		r.Post("/upload", handler.UploadFile)
	})

	r.Route("/admin", func(r chi.Router) {
		r.Use(appmw.RequireAuth(authManager))
		r.Use(appmw.RequireRole("admin"))
		r.Get("/audit", handler.AdminAudit)
	})

	return &Server{router: r}
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
