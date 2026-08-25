package config

import (
    "fmt"
    "os"
    "path/filepath"
)

type Config struct {
    Host    string
    Port    int
    DataDir string
}

func Load() (*Config, error) {
    homeDir, err := os.UserHomeDir()
    if err != nil {
        return nil, fmt.Errorf("failed to get home directory: %w", err)
    }
    
    dataDir := filepath.Join(homeDir, ".local-apk-builder")
    
    port := 8080
    if portStr := os.Getenv("PORT"); portStr != "" {
        fmt.Sscanf(portStr, "%d", &port)
    }
    
    return &Config{
        Host:    "127.0.0.1",
        Port:    port,
        DataDir: dataDir,
    }, nil
}

func (c *Config) Addr() string {
    return fmt.Sprintf("%s:%d", c.Host, c.Port)
}

func (c *Config) URL() string {
    return fmt.Sprintf("http://%s", c.Addr())
}
