package generator

import (
    "encoding/json"
    "fmt"
    "os"
    "path/filepath"
    "strings"
)

// ProjectProfile describes a detected project
type ProjectProfile struct {
    Type      string // "static" | "node" | "laravel" | "unknown"
    Framework string // "none" | "react" | "vue"
    Builder   string // "none" | "vite"
    EntryDir  string // relative path to index.html or src/
    OutputDir string // relative path to build output (e.g., "dist/")
}

// NeedsBuild returns true if npm build is required
func (p ProjectProfile) NeedsBuild() bool {
    return p.Type == "node"
}

// ProjectDetector detects project type from source directory
type ProjectDetector struct {
    SourceDir string
}

// Detect inspects the source directory and returns a ProjectProfile
func (d *ProjectDetector) Detect() (*ProjectProfile, error) {
    // Normalize wrapper directory
    sourceDir, err := d.normalizeWrapperDir()
    if err != nil {
        return nil, err
    }
    d.SourceDir = sourceDir
    
    // Check for package.json
    pkgPath := filepath.Join(sourceDir, "package.json")
    if _, err := os.Stat(pkgPath); err == nil {
        return d.detectNodeProject(sourceDir)
    }
    
    // Check for index.html
    indexPath := filepath.Join(sourceDir, "index.html")
    if _, err := os.Stat(indexPath); err == nil {
        return &ProjectProfile{
            Type:      "static",
            Framework: "none",
            Builder:   "none",
            EntryDir:  "",
            OutputDir: "",
        }, nil
    }
    
    // Check for Laravel
    artisanPath := filepath.Join(sourceDir, "artisan")
    composerPath := filepath.Join(sourceDir, "composer.json")
    if _, err := os.Stat(artisanPath); err == nil {
        if _, err := os.Stat(composerPath); err == nil {
            return &ProjectProfile{
                Type:      "laravel",
                Framework: "none",
                Builder:   "none",
                EntryDir:  "",
                OutputDir: "",
            }, nil
        }
    }
    
    return &ProjectProfile{
        Type:      "unknown",
        Framework: "none",
        Builder:   "none",
        EntryDir:  "",
        OutputDir: "",
    }, nil
}

// detectNodeProject inspects package.json and config files
func (d *ProjectDetector) detectNodeProject(sourceDir string) (*ProjectProfile, error) {
    profile := &ProjectProfile{
        Type:      "node",
        Framework: "none",
        Builder:   "none",
        OutputDir: "dist",
    }
    
    // Read package.json
    pkgData, err := os.ReadFile(filepath.Join(sourceDir, "package.json"))
    if err != nil {
        return nil, err
    }
    
    var pkg struct {
        Dependencies    map[string]string `json:"dependencies"`
        DevDependencies map[string]string `json:"devDependencies"`
        Scripts         map[string]string `json:"scripts"`
    }
    if err := json.Unmarshal(pkgData, &pkg); err != nil {
        return nil, fmt.Errorf("invalid package.json: %w", err)
    }
    
    // Detect builder (vite)
    if hasViteConfig(sourceDir) {
        profile.Builder = "vite"
    }
    
    // Detect framework
    if _, ok := pkg.Dependencies["react"]; ok {
        profile.Framework = "react"
    }
    if _, ok := pkg.DevDependencies["react"]; ok {
        profile.Framework = "react"
    }
    if _, ok := pkg.Dependencies["vue"]; ok {
        profile.Framework = "vue"
    }
    if _, ok := pkg.DevDependencies["vue"]; ok {
        profile.Framework = "vue"
    }
    
    // Verify build script exists
    if _, ok := pkg.Scripts["build"]; !ok {
        return nil, fmt.Errorf("%w: package.json missing 'build' script", ErrInvalidNode)
    }
    
    // Set entry dir
    if profile.Builder == "vite" {
        profile.EntryDir = "src"
    } else {
        profile.EntryDir = ""
    }
    
    return profile, nil
}

// hasViteConfig checks for vite.config.* files
func hasViteConfig(sourceDir string) bool {
    for _, name := range []string{
        "vite.config.js",
        "vite.config.ts",
        "vite.config.mjs",
        "vite.config.cjs",
    } {
        if _, err := os.Stat(filepath.Join(sourceDir, name)); err == nil {
            return true
        }
    }
    return false
}

// normalizeWrapperDir collapses a single top-level directory to source root
func (d *ProjectDetector) normalizeWrapperDir() (string, error) {
    entries, err := os.ReadDir(d.SourceDir)
    if err != nil {
        return "", err
    }
    
    // Filter out hidden files
    visible := make([]os.DirEntry, 0, len(entries))
    for _, e := range entries {
        if !strings.HasPrefix(e.Name(), ".") {
            visible = append(visible, e)
        }
    }
    
    // If exactly 1 directory and 0 files, use it as root
    if len(visible) == 1 && visible[0].IsDir() {
        return filepath.Join(d.SourceDir, visible[0].Name()), nil
    }
    
    return d.SourceDir, nil
}
