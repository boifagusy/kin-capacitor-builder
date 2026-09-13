package generator

import (
    "os"
    "path/filepath"
    "testing"
)

func createTestProject(t *testing.T, files map[string]string) string {
    t.Helper()
    dir := t.TempDir()
    for name, content := range files {
        path := filepath.Join(dir, filepath.FromSlash(name))
        os.MkdirAll(filepath.Dir(path), 0755)
        os.WriteFile(path, []byte(content), 0644)
    }
    return dir
}

func TestDetectStatic(t *testing.T) {
    dir := createTestProject(t, map[string]string{
        "index.html":       "<html>Test</html>",
        "css/app.css":      "body{}",
        "js/app.js":        "console.log('hi')",
    })
    
    d := &ProjectDetector{SourceDir: dir}
    profile, err := d.Detect()
    if err != nil {
        t.Fatalf("Detect failed: %v", err)
    }
    
    if profile.Type != "static" {
        t.Errorf("Expected static, got %s", profile.Type)
    }
}

func TestDetectViteReact(t *testing.T) {
    dir := createTestProject(t, map[string]string{
        "package.json": `{
            "dependencies": {"react": "^18.0.0"},
            "scripts": {"build": "vite build"}
        }`,
        "vite.config.js":   "export default {}",
        "src/main.jsx":     "console.log('react')",
        "index.html":       "<html>Vite</html>",
    })
    
    d := &ProjectDetector{SourceDir: dir}
    profile, err := d.Detect()
    if err != nil {
        t.Fatalf("Detect failed: %v", err)
    }
    
    if profile.Type != "node" {
        t.Errorf("Expected node, got %s", profile.Type)
    }
    if profile.Builder != "vite" {
        t.Errorf("Expected vite builder, got %s", profile.Builder)
    }
    if profile.Framework != "react" {
        t.Errorf("Expected react framework, got %s", profile.Framework)
    }
}

func TestDetectViteVue(t *testing.T) {
    dir := createTestProject(t, map[string]string{
        "package.json": `{
            "dependencies": {"vue": "^3.0.0"},
            "scripts": {"build": "vite build"}
        }`,
        "vite.config.ts":   "export default {}",
        "src/main.ts":      "console.log('vue')",
    })
    
    d := &ProjectDetector{SourceDir: dir}
    profile, err := d.Detect()
    if err != nil {
        t.Fatalf("Detect failed: %v", err)
    }
    
    if profile.Builder != "vite" {
        t.Errorf("Expected vite builder, got %s", profile.Builder)
    }
    if profile.Framework != "vue" {
        t.Errorf("Expected vue framework, got %s", profile.Framework)
    }
}

func TestDetectWrapperDirectory(t *testing.T) {
    dir := createTestProject(t, map[string]string{
        "my-website/index.html":  "<html>Wrapper</html>",
        "my-website/css/app.css": "body{}",
    })
    
    d := &ProjectDetector{SourceDir: dir}
    profile, err := d.Detect()
    if err != nil {
        t.Fatalf("Detect failed: %v", err)
    }
    
    if profile.Type != "static" {
        t.Errorf("Expected static, got %s", profile.Type)
    }
}

func TestDetectLaravel(t *testing.T) {
    dir := createTestProject(t, map[string]string{
        "artisan":       "#!/usr/bin/env php",
        "composer.json": `{"require": {"laravel/framework": "^10"}}`,
    })
    
    d := &ProjectDetector{SourceDir: dir}
    profile, err := d.Detect()
    if err != nil {
        t.Fatalf("Detect failed: %v", err)
    }
    
    if profile.Type != "laravel" {
        t.Errorf("Expected laravel, got %s", profile.Type)
    }
}

func TestDetectUnknown(t *testing.T) {
    dir := createTestProject(t, map[string]string{
        "README.md": "# Nothing here",
    })
    
    d := &ProjectDetector{SourceDir: dir}
    profile, err := d.Detect()
    if err != nil {
        t.Fatalf("Detect failed: %v", err)
    }
    
    if profile.Type != "unknown" {
        t.Errorf("Expected unknown, got %s", profile.Type)
    }
}
