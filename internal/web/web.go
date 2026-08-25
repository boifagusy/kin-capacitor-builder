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

// ParseTemplates parses all templates from the embedded filesystem
func ParseTemplates() (*template.Template, error) {
    // Create a root template
    root := template.New("root")
    
    // Parse base.html first
    baseContent, err := fs.ReadFile(TemplatesFS, "templates/base.html")
    if err != nil {
        return nil, err
    }
    
    _, err = root.Parse(string(baseContent))
    if err != nil {
        return nil, err
    }
    
    // Parse partial templates
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
    
    // Parse component templates
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

// StaticFileServer returns an http.FileSystem for static files
func StaticFileServer() (fs.FS, error) {
    return fs.Sub(StaticFS, "static")
}
