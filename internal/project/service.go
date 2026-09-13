package project

import (
    "fmt"
    "strings"
    
    "local-apk-builder/internal/database"
)

type Service struct{}

func NewService() *Service {
    return &Service{}
}

func (s *Service) Create(p *database.Project) (*database.Project, error) {
    if err := database.InsertProject(p); err != nil {
        return nil, err
    }
    return p, nil
}

func (s *Service) Update(p *database.Project) (*database.Project, error) {
    if p.ID == 0 {
        return nil, fmt.Errorf("project ID required")
    }
    if err := database.UpdateProject(p); err != nil {
        return nil, err
    }
    return p, nil
}

func (s *Service) Delete(id int64) error {
    return database.DeleteProject(id)
}

func (s *Service) Get(id int64) (*database.Project, error) {
    return database.GetProject(id)
}

func (s *Service) List() ([]*database.Project, error) {
    return database.ListProjects()
}

func (s *Service) Validate(p *database.Project) error {
    if strings.TrimSpace(p.URL) == "" {
        return fmt.Errorf("URL is required")
    }
    if strings.TrimSpace(p.AppName) == "" {
        return fmt.Errorf("app name is required")
    }
    return nil
}

func (s *Service) GenerateProjectName(url string) string {
    name := strings.TrimPrefix(url, "https://")
    name = strings.TrimPrefix(name, "http://")
    name = strings.TrimPrefix(name, "www.")
    parts := strings.Split(name, ".")
    if len(parts) > 0 {
        name = parts[0]
    }
    if len(name) > 0 {
        name = strings.ToUpper(name[:1]) + name[1:]
    }
    return name
}
