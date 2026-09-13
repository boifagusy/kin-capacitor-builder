package web

import (
    "embed"
    "html/template"
    "io/fs"
)

//go:embed all:templates
var TemplatesFS embed.FS

//go:embed all:static
var StaticFS embed.FS

func ParseTemplates() (*template.Template, error) {
    root := template.New("root")
    
    // Parse base template
    baseContent, err := fs.ReadFile(TemplatesFS, "templates/base.html")
    if err != nil {
        return nil, err
    }
    _, err = root.Parse(string(baseContent))
    if err != nil {
        return nil, err
    }
    
    // Parse all templates
    files, err := fs.Glob(TemplatesFS, "templates/*.html")
    if err != nil {
        return nil, err
    }
    
    for _, file := range files {
        content, err := fs.ReadFile(TemplatesFS, file)
        if err != nil {
            return nil, err
        }
        _, err = root.Parse(string(content))
        if err != nil {
            return nil, err
        }
    }
    
    // Parse partials
    partialFiles, err := fs.Glob(TemplatesFS, "templates/partials/*.html")
    if err != nil {
        return nil, err
    }
    for _, file := range partialFiles {
        content, err := fs.ReadFile(TemplatesFS, file)
        if err != nil {
            return nil, err
        }
        _, err = root.Parse(string(content))
        if err != nil {
            return nil, err
        }
    }
    
    // Parse components
    componentFiles, err := fs.Glob(TemplatesFS, "templates/components/*.html")
    if err != nil {
        return nil, err
    }
    for _, file := range componentFiles {
        content, err := fs.ReadFile(TemplatesFS, file)
        if err != nil {
            return nil, err
        }
        _, err = root.Parse(string(content))
        if err != nil {
            return nil, err
        }
    }
    
    return root, nil
}

func StaticFileServer() (fs.FS, error) {
    return fs.Sub(StaticFS, "static")
}
