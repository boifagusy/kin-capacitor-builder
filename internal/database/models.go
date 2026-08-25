package database

import "time"

type Project struct {
    ID               int64     `json:"id"`
    URL              string    `json:"url"`
    AppName          string    `json:"app_name"`
    LogoPath         string    `json:"logo_path"`
    PrimaryColor     string    `json:"primary_color"`
    OnboardingConfig string    `json:"onboarding_config"`
    CurrentStep      int       `json:"current_step"`
    CreatedAt        time.Time `json:"created_at"`
    UpdatedAt        time.Time `json:"updated_at"`
}

func NewProject() *Project {
    return &Project{
        URL:              "",
        AppName:          "",
        LogoPath:         "",
        PrimaryColor:     "#3B82F6",
        OnboardingConfig: `{"has_splash_screen":true,"splash_color":"#FFFFFF","orientation":"portrait"}`,
        CurrentStep:      1,
        CreatedAt:        time.Now(),
        UpdatedAt:        time.Now(),
    }
}
