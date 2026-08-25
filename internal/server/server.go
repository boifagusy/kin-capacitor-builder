package server

import (
    "context"
    "log"
    "net/http"
    "os"
    "os/signal"
    "syscall"
    "time"
    
    "local-apk-builder/internal/config"
    "local-apk-builder/internal/web"
)

type Server struct {
    httpServer *http.Server
    config     *config.Config
    handler    *Handler
}

func New(cfg *config.Config) (*Server, error) {
    templates, err := web.ParseTemplates()
    if err != nil {
        return nil, err
    }
    
    handler := NewHandler(templates)
    
    mux := http.NewServeMux()
    
    mux.HandleFunc("/", handler.IndexHandler)
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
    mux.HandleFunc("/api/validate-url", handler.ValidateURLHandler)
    mux.HandleFunc("/api/project-state", handler.ProjectStateHandler)
    
    // Serve static files
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
    }, nil
}

func (s *Server) Start() error {
    log.Printf("Starting server on %s", s.config.Addr())
    log.Printf("Server URL: %s", s.config.URL())
    
    go func() {
        if err := s.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            log.Fatalf("Server failed: %v", err)
        }
    }()
    
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit
    
    log.Println("Shutting down server...")
    
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()
    
    if err := s.httpServer.Shutdown(ctx); err != nil {
        log.Printf("Server forced to shutdown: %v", err)
        return err
    }
    
    log.Println("Server exited gracefully")
    return nil
}
