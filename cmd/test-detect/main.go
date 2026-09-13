package main

import (
    "fmt"
    "local-apk-builder/internal/generator"
)

func main() {
    d := &generator.ProjectDetector{SourceDir: "/data/data/com.termux/files/home/.local-apk-builder/projects/7/source"}
    profile, err := d.Detect()
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        return
    }
    fmt.Printf("Type:      %s\n", profile.Type)
    fmt.Printf("Framework: %s\n", profile.Framework)
    fmt.Printf("Builder:   %s\n", profile.Builder)
    fmt.Printf("NeedsBuild: %v\n", profile.NeedsBuild())
}
