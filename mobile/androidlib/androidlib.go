package androidlib

import (
	"context"
	"fmt"
	"os"
	"sync"

	"local-apk-builder/internal/config"
	"local-apk-builder/internal/database"
	"local-apk-builder/internal/server"
)

const DefaultPort = 18791

var (
	mu      sync.Mutex
	srv     *server.Server
	cancel  context.CancelFunc
	running bool
)

// Start launches the builder server. Returns immediately.
// port <= 0 uses DefaultPort.
func Start(dataDir string, port int) error {
	mu.Lock()
	defer mu.Unlock()
	if running {
		return fmt.Errorf("server already running")
	}
	if port <= 0 {
		port = DefaultPort
	}
	os.Setenv("KIN_DATA_DIR", dataDir)
	os.Setenv("PORT", fmt.Sprintf("%d", port))
	os.Setenv("KIN_DISABLE_LOCAL_SIGNING", "1")

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("config: %w", err)
	}
	if err := database.Init(cfg.DataDir); err != nil {
		return fmt.Errorf("database: %w", err)
	}
	s, err := server.New(cfg)
	if err != nil {
		return fmt.Errorf("server init: %w", err)
	}
	srv = s
	ctx, c := context.WithCancel(context.Background())
	cancel = c
	go func() { _ = srv.Run(ctx) }()
	running = true
	return nil
}

// Stop shuts the builder server down.
func Stop() error {
	mu.Lock()
	defer mu.Unlock()
	if !running {
		return nil
	}
	if cancel != nil {
		cancel()
	}
	running = false
	return nil
}
