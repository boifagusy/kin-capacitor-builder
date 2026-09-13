package build

import (
    "encoding/json"
    "local-apk-builder/internal/database"
)

// BuildConfigFromProject creates a BuildConfig snapshot from project
func BuildConfigFromProject(build *database.Build, project *database.Project) BuildConfig {
    var features []string
    if project.FeaturesConfig != "" {
        // Try JSON array, fallback comma-separated
        if err := json.Unmarshal([]byte(project.FeaturesConfig), &features); err != nil {
            features = splitCSV(project.FeaturesConfig)
        }
    }
    var splash map[string]interface{}
    json.Unmarshal([]byte(project.SplashConfig), &splash)
    var nav map[string]interface{}
    json.Unmarshal([]byte(project.OnboardingConfig), &nav)

    return BuildConfig{
        ProjectID:    project.ID,
        BuildID:      build.ID,
        URL:          project.URL,
        AppName:      project.AppName,
        LogoPath:     project.LogoPath,
        PrimaryColor: project.PrimaryColor,
        Features:     features,
        Splash:       splash,
        Navigation:   nav,
    }
}

func splitCSV(s string) []string {
    var out []string
    for _, p := range splitStr(s, ",") {
        if p != "" { out = append(out, p) }
    }
    return out
}

func splitStr(s, sep string) []string {
    var res []string
    start := 0
    for i := 0; i <= len(s); i++ {
        if i == len(s) || s[i] == sep[0] {
            res = append(res, s[start:i])
            start = i + 1
        }
    }
    return res
}
