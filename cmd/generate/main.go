package main

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"

    "local-apk-builder/internal/database"
    "local-apk-builder/internal/generator"
)

type BuildPayload struct {
    BuildID int64                  `json:"build_id"`
    Config  map[string]interface{} `json:"config"`
}

func main() {
    payloadPath := os.Getenv("BUILD_PAYLOAD_FILE")
    if payloadPath == "" {
        payloadPath = "/tmp/build_payload.json"
    }

    data, err := os.ReadFile(payloadPath)
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to read payload: %v\n", err)
        os.Exit(1)
    }

    var payload BuildPayload
    if err := json.Unmarshal(data, &payload); err != nil {
        fmt.Fprintf(os.Stderr, "invalid JSON payload: %v\n", err)
        os.Exit(1)
    }

    // Convert map to database.Project
    project, err := projectFromConfig(payload.Config)
    if err != nil {
        fmt.Fprintf(os.Stderr, "invalid config: %v\n", err)
        os.Exit(1)
    }
    project.ID = payload.BuildID // not strictly necessary; generator uses ID for appId
    if project.ID == 0 {
        project.ID = 1
    }

    gen := generator.NewGenerator(project)

    capConfig, err := gen.GenerateCapacitorConfig()
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to generate capacitor config: %v\n", err)
        os.Exit(1)
    }
    pkgJSON, err := gen.GeneratePackageJSON()
    if err != nil {
        fmt.Fprintf(os.Stderr, "failed to generate package.json: %v\n", err)
        os.Exit(1)
    }
    webIndex := gen.GenerateWebIndex()
    androidManifest := gen.GenerateAndroidManifest()
    iosPlist := gen.GenerateIOSInfoPlist()

    outDir := "generated"
    dirs := []string{
        outDir,
        filepath.Join(outDir, "www"),
        filepath.Join(outDir, "android", "app", "src", "main"),
        filepath.Join(outDir, "ios", "App", "App"),
    }
    for _, d := range dirs {
        if err := os.MkdirAll(d, 0755); err != nil {
            fmt.Fprintf(os.Stderr, "failed to create dir %s: %v\n", d, err)
            os.Exit(1)
        }
    }

    writeFile(filepath.Join(outDir, "capacitor.config.json"), capConfig)
    writeFile(filepath.Join(outDir, "package.json"), pkgJSON)
    writeFile(filepath.Join(outDir, "www", "index.html"), webIndex)
    writeFile(filepath.Join(outDir, "android", "app", "src", "main", "AndroidManifest.xml"), androidManifest)
    writeFile(filepath.Join(outDir, "ios", "App", "App", "Info.plist"), iosPlist)

    fmt.Println("Generated project in", outDir)
}

func writeFile(path, content string) {
    if err := os.WriteFile(path, []byte(content), 0644); err != nil {
        fmt.Fprintf(os.Stderr, "failed to write %s: %v\n", path, err)
        os.Exit(1)
    }
}

// projectFromConfig builds a database.Project from a map
func projectFromConfig(m map[string]interface{}) (*database.Project, error) {
    url, _ := m["url"].(string)
    appName, _ := m["app_name"].(string)
    if url == "" || appName == "" {
        return nil, fmt.Errorf("missing required fields: url and app_name")
    }

    project := database.NewProject()
    project.URL = url
    project.AppName = appName
    project.ProjectName = appName

    if v, ok := m["logo_path"].(string); ok {
        project.LogoPath = v
    }
    if v, ok := m["primary_color"].(string); ok && v != "" {
        project.PrimaryColor = v
    }
    if v, ok := m["features"].([]interface{}); ok {
        features := []string{}
        for _, f := range v {
            if s, ok := f.(string); ok {
                features = append(features, s)
            }
        }
        project.FeaturesConfig = toJSON(features)
    } else if v, ok := m["features"].(string); ok {
        project.FeaturesConfig = v
    }
    if v, ok := m["splash"].(map[string]interface{}); ok {
        project.SplashConfig = toJSON(v)
    }
    if v, ok := m["navigation"].(map[string]interface{}); ok {
        project.OnboardingConfig = toJSON(v)
    }

    return project, nil
}

func toJSON(v interface{}) string {
    data, _ := json.Marshal(v)
    return string(data)
}
