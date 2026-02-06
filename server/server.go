package server

import (
	"context"
	"io/fs"
	"net/http"

	"github.com/cloud-gt/ai-sensors/command"
	"github.com/cloud-gt/ai-sensors/manager"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Server struct {
	router  chi.Router
	server  *http.Server
	manager *manager.Manager
}

func New(store *command.Store, mgr *manager.Manager) *Server {
	s := &Server{
		router:  chi.NewRouter(),
		manager: mgr,
	}

	s.router.Use(middleware.Logger)

	commandsAPI := NewCommandsAPI(store, mgr)
	s.router.Mount("/commands", commandsAPI.Router())

	return s
}

// MountDashboard registers the SPA dashboard handler at /dashboard.
func (s *Server) MountDashboard(dashFS fs.FS) {
	handler := NewDashboardHandler(dashFS)
	s.router.Mount("/dashboard", http.StripPrefix("/dashboard", handler))
	s.router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/dashboard", http.StatusTemporaryRedirect)
	})
}

func (s *Server) Router() chi.Router {
	return s.router
}

func (s *Server) ListenAndServe(addr string) error {
	s.server = &http.Server{
		Addr:    addr,
		Handler: s.router,
	}
	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	if err := s.manager.Shutdown(ctx); err != nil {
		return err
	}
	if s.server != nil {
		return s.server.Shutdown(ctx)
	}
	return nil
}
