package api

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/BlueBeard63/archon-node/internal/config"
	"github.com/BlueBeard63/archon-node/internal/docker"
	"github.com/BlueBeard63/archon-node/internal/proxy"
	"github.com/BlueBeard63/archon-node/internal/ssl"
)

type Server struct {
	config       *config.Config
	router       *chi.Mux
	server       *http.Server
	dockerClient *docker.Client
	proxyManager proxy.ProxyManager
	sslManager   *ssl.Manager
}

func NewServer(cfg *config.Config) (*Server, error) {
	// Create Docker client
	dockerClient, err := docker.NewClient(cfg.Docker.Host, cfg.Docker.Network)
	if err != nil {
		return nil, fmt.Errorf("failed to create docker client: %w", err)
	}

	// Create proxy manager
	proxyManager, err := proxy.NewProxyManager(&cfg.Proxy, &cfg.SSL)
	if err != nil {
		return nil, fmt.Errorf("failed to create proxy manager: %w", err)
	}

	// Create SSL manager
	sslManager := ssl.NewManager(&cfg.SSL)

	// Create router
	r := chi.NewRouter()

	// Middleware
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(LoggingMiddleware)

	// Create handlers
	handlers := NewHandlers(dockerClient, proxyManager, sslManager, cfg.Server.DataDir)

	// Public routes (no auth required)
	r.Get("/health", handlers.HandleHealth)

	// Protected routes (require API key)
	r.Group(func(r chi.Router) {
		r.Use(AuthMiddleware(cfg.Server.APIKey))

		// Site management
		r.Post("/api/v1/sites/deploy", handlers.HandleDeploySite)
		r.Get("/api/v1/sites/{siteID}/status", handlers.HandleGetSiteStatus)
		r.Post("/api/v1/sites/{siteID}/stop", handlers.HandleStopSite)
		r.Post("/api/v1/sites/{siteID}/restart", handlers.HandleRestartSite)
		r.Delete("/api/v1/sites/{siteID}", handlers.HandleDeleteSite)
		r.Get("/api/v1/sites/{siteID}/logs", handlers.HandleGetLogs)
	})

	// Create HTTP server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	httpServer := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		config:       cfg,
		router:       r,
		server:       httpServer,
		dockerClient: dockerClient,
		proxyManager: proxyManager,
		sslManager:   sslManager,
	}, nil
}

func (s *Server) Start() error {
	// Ensure Docker network exists
	ctx := context.Background()
	if err := s.dockerClient.EnsureNetwork(ctx); err != nil {
		return fmt.Errorf("failed to ensure Docker network: %w", err)
	}

	fmt.Printf("Starting Archon Node server on %s\n", s.server.Addr)
	fmt.Printf("Proxy type: %s\n", s.config.Proxy.Type)
	fmt.Printf("SSL mode: %s\n", s.config.SSL.Mode)

	return s.server.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	fmt.Println("Shutting down server...")

	// Close Docker client
	if err := s.dockerClient.Close(); err != nil {
		fmt.Printf("Warning: failed to close Docker client: %v\n", err)
	}

	return s.server.Shutdown(ctx)
}
