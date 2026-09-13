package server

import (
    "fmt"
    "net/url"
    "strings"
)

func ValidateURL(rawURL string) error {
    rawURL = strings.TrimSpace(rawURL)
    
    if rawURL == "" {
        return fmt.Errorf("URL cannot be empty")
    }
    
    if !strings.Contains(rawURL, "://") {
        rawURL = "https://" + rawURL
    }
    
    parsedURL, err := url.Parse(rawURL)
    if err != nil {
        return fmt.Errorf("invalid URL format")
    }
    
    if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
        return fmt.Errorf("URL must use http or https scheme")
    }
    
    if parsedURL.Host == "" {
        return fmt.Errorf("URL must include a domain name")
    }
    
    if !strings.Contains(parsedURL.Host, ".") {
        return fmt.Errorf("invalid domain name")
    }
    
    if strings.Contains(parsedURL.Host, "localhost") || strings.Contains(parsedURL.Host, "127.0.0.1") {
        return fmt.Errorf("localhost URLs are not allowed")
    }
    
    return nil
}

func NormalizeURL(rawURL string) string {
    rawURL = strings.TrimSpace(rawURL)
    if !strings.Contains(rawURL, "://") {
        rawURL = "https://" + rawURL
    }
    return rawURL
}

func ValidateAppName(appName string) error {
    appName = strings.TrimSpace(appName)
    
    if appName == "" {
        return fmt.Errorf("app name cannot be empty")
    }
    
    if len(appName) < 2 {
        return fmt.Errorf("app name must be at least 2 characters")
    }
    
    if len(appName) > 50 {
        return fmt.Errorf("app name must be less than 50 characters")
    }
    
    return nil
}

func ValidateColor(color string) error {
    color = strings.TrimSpace(color)
    
    if color == "" {
        return fmt.Errorf("color cannot be empty")
    }
    
    if !strings.HasPrefix(color, "#") {
        return fmt.Errorf("color must start with #")
    }
    
    if len(color) != 7 {
        return fmt.Errorf("color must be in #RRGGBB format")
    }
    
    for _, c := range color[1:] {
        if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
            return fmt.Errorf("color must contain only hex characters")
        }
    }
    
    return nil
}
