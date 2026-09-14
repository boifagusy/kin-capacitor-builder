package server

import (
	"context"
	"fmt"
	"local-apk-builder/internal/database"
	"log"
	"net/http"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"local-apk-builder/internal/auth"
	"local-apk-builder/internal/config"
	"local-apk-builder/internal/dashboard"
	"local-apk-builder/internal/web"
)

type Server struct {
	httpServer *http.Server
	config     *config.Config
	handler    *Handler
	dashboard  *dashboard.Handler
}

func New(cfg *config.Config) (*Server, error) {
	templates, err := web.ParseTemplates()
	if err != nil {
		return nil, err
	}

	handler := NewHandler(templates)
	dashHandler := dashboard.NewHandler(templates)

	mux := http.NewServeMux()

	// Auth routes
	oauthCfg := config.LoadOAuthConfig()
	authHandler := auth.NewHandler(oauthCfg, templates)
	mux.HandleFunc("/auth/github", authHandler.Login)
	mux.HandleFunc("/auth/github/callback", authHandler.Callback)
	mux.HandleFunc("/api/github/status", authHandler.StatusHandler)
	mux.HandleFunc("/guide/backend-bridge", func(w http.ResponseWriter, r *http.Request) {
		data, err := web.StaticFS.ReadFile("static/guide/backend-bridge.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Write(data)
	})

	mux.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"status":"ok","time":"%s"}`, time.Now().Format(time.RFC3339))
	})
	mux.HandleFunc("/settings/github/pat", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Redirect(w, r, "/settings/github", http.StatusFound)
			return
		}
		patToken := r.FormValue("pat_token")
		if patToken == "" {
			http.Redirect(w, r, "/settings/github?error=empty", http.StatusFound)
			return
		}
		// Save PAT to database
		err := database.UpdateAccessToken(patToken)
		if err != nil {
			http.Error(w, "Failed to save token: "+err.Error(), http.StatusInternalServerError)
			return
		}
		http.Redirect(w, r, "/settings/github?saved=pat", http.StatusFound)
	})

	mux.HandleFunc("/settings/github/pat/validate", authHandler.PATValidateHandler)

	mux.HandleFunc("/settings/github", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			authHandler.SettingsPageHandler(w, r)
		case http.MethodPost:
			authHandler.SaveSettingsHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Dashboard routes
	mux.HandleFunc("/dashboard", dashHandler.DashboardHandler)
	mux.HandleFunc("/welcome", handler.WelcomeHandler)

	// Test route for Alpine debugging
	mux.HandleFunc("/test-alpine", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		templates.ExecuteTemplate(w, "test_alpine", nil)
	})

	// Project routes
	mux.HandleFunc("/projects/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/build-aab") {
			dashHandler.EnqueueAABBuildHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/build") {
			dashHandler.EnqueueBuildHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/builds") {
			dashHandler.ListBuildsHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/build-status") {
			dashHandler.BuildStatusHTMLHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/source/upload") {
			dashHandler.UploadSourceHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/download") {
			dashHandler.DownloadBuildHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/preview") {
			handler.PreviewHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/generate") {
			dashHandler.GenerateProjectHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/delete") {
			dashHandler.DeleteProjectHandler(w, r)
			return
		}
		dashHandler.ProjectDetailHandler(w, r)
	})

	// Root
	mux.HandleFunc("/", handler.IndexHandler)

	// Wizard routes - each registered exactly once
	mux.HandleFunc("/wizard/1", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.Screen1Handler(w, r)
		case http.MethodPost:
			handler.Screen1PostHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/wizard/2", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.Screen2Handler(w, r)
		case http.MethodPost:
			handler.Screen2PostHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/wizard/3", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.Screen3Handler(w, r)
		case http.MethodPost:
			handler.Screen3PostHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/wizard/4", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.Screen4Handler(w, r)
		case http.MethodPost:
			handler.Screen4PostHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/wizard/5", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.Screen5Handler(w, r)
		case http.MethodPost:
			handler.Screen5PostHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/wizard/6", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodGet:
			handler.Screen6Handler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// API routes
	mux.HandleFunc("/api/validate-url", handler.ValidateURLHandler)
	mux.HandleFunc("/api/project-state", handler.ProjectStateHandler)
	mux.HandleFunc("/api/projects", dashHandler.APIListProjects)
	mux.HandleFunc("/api/projects/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/google-services") && r.Method == http.MethodPost {
			dashHandler.UploadGoogleServicesHandler(w, r)
			return
		}
		if strings.HasSuffix(r.URL.Path, "/navigation") {
			dashHandler.SaveNavigationHandler(w, r)
			return
		}
		dashHandler.APIGetProject(w, r)
	})

	// Static files
	staticFS, err := web.StaticFileServer()
	if err != nil {
		return nil, err
	}
	staticHandler := http.FileServer(http.FS(staticFS))
	mux.Handle("/static/", http.StripPrefix("/static/", staticHandler))

	var finalHandler http.Handler = mux
	finalHandler = loggingMiddleware(finalHandler)
	finalHandler = recoveryMiddleware(finalHandler)

	httpServer := &http.Server{
		Addr:         cfg.Addr(),
		Handler:      finalHandler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return &Server{
		httpServer: httpServer,
		config:     cfg,
		handler:    handler,
		dashboard:  dashHandler,
	}, nil
}

func (s *Server) Start() error {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	return s.Run(ctx)
}

func (s *Server) Run(ctx context.Context) error {
	log.Printf("Starting server on %s", s.config.Addr())
	log.Printf("Server URL: %s", s.config.URL())

	errCh := make(chan error, 1)
	go func() {
		if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errCh <- err
		}
	}()

	select {
	case <-ctx.Done():
		log.Println("Shutting down server (context cancelled)...")
	case err := <-errCh:
		return err
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := s.httpServer.Shutdown(shutdownCtx); err != nil {
		return err
	}
	log.Println("Server exited gracefully")
	return nil
}
