package config

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LoadEnvFile loads KEY=VALUE pairs from .env into environment
func LoadEnvFile(path string) {
	f, err := os.Open(path)
	if err != nil {
		return
	}
	defer f.Close()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) == 2 {
			key := strings.TrimSpace(parts[0])
			value := strings.TrimSpace(parts[1])
			if os.Getenv(key) == "" {
				os.Setenv(key, value)
			}
		}
	}
}

type Config struct {
	Host    string
	Port    int
	DataDir string
}

func Load() (*Config, error) {
	dataDir := os.Getenv("KIN_DATA_DIR")
	if dataDir == "" {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, fmt.Errorf("failed to get home directory: %w", err)
		}
		dataDir = filepath.Join(homeDir, ".local-apk-builder")
	}

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

// GitHubConfig holds GitHub Actions provider settings
type GitHubConfig struct {
	Token        string
	Repo         string
	WorkflowID   string
	ArtifactName string
}

// LoadGitHubConfig from environment
func LoadGitHubConfig() *GitHubConfig {
	return &GitHubConfig{
		Token:        os.Getenv("GITHUB_TOKEN"),
		Repo:         os.Getenv("GITHUB_REPO"),
		WorkflowID:   os.Getenv("GITHUB_WORKFLOW_ID"),
		ArtifactName: os.Getenv("GITHUB_ARTIFACT_NAME"),
	}
}

// OAuthConfig holds GitHub OAuth app settings
type OAuthConfig struct {
	ClientID     string
	ClientSecret string
	CallbackURL  string
}

// LoadOAuthConfig from environment
func LoadOAuthConfig() *OAuthConfig {
	return &OAuthConfig{
		ClientID:     os.Getenv("GITHUB_CLIENT_ID"),
		ClientSecret: os.Getenv("GITHUB_CLIENT_SECRET"),
		CallbackURL:  os.Getenv("GITHUB_CALLBACK_URL"),
	}
}
