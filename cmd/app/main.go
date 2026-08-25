package main

import (
    "log"
    
    "local-apk-builder/internal/config"
    "local-apk-builder/internal/database"
    "local-apk-builder/internal/server"
)

func main() {
    cfg, err := config.Load()
    if err != nil {
        log.Fatalf("Failed to load configuration: %v", err)
    }
    
    if err := database.Init(cfg.DataDir); err != nil {
        log.Fatalf("Failed to initialize database: %v", err)
    }
    defer database.Close()
    
    log.Printf("Database initialized at: %s/builder.db", cfg.DataDir)
    
    srv, err := server.New(cfg)
    if err != nil {
        log.Fatalf("Failed to create server: %v", err)
    }
    
    if err := srv.Start(); err != nil {
        log.Fatalf("Server error: %v", err)
    }
}
