package upload

import (
    "errors"
    "testing"
)

func TestNormalizePathValid(t *testing.T) {
    cases := []struct {
        input    string
        expected string
    }{
        {"index.html", "index.html"},
        {"css/app.css", "css/app.css"},
        {"js/main.js", "js/main.js"},
        {"images/logo.png", "images/logo.png"},
        {"frontend/index.html", "frontend/index.html"},
        {`windows\path\file.js`, "windows/path/file.js"},
        {"./index.html", "index.html"},
        {"a/./b/index.html", "a/b/index.html"},
    }
    
    for _, c := range cases {
        result, err := NormalizePath(c.input)
        if err != nil {
            t.Errorf("NormalizePath(%q) unexpected error: %v", c.input, err)
        }
        if result != c.expected {
            t.Errorf("NormalizePath(%q) = %q, expected %q", c.input, result, c.expected)
        }
    }
}

func TestNormalizePathTraversal(t *testing.T) {
    cases := []string{
        "../evil.txt",
        "a/../../evil.txt",
        "..\\evil.txt",
        "../a/../b/evil.txt",
    }
    
    for _, c := range cases {
        _, err := NormalizePath(c)
        if !errors.Is(err, ErrPathTraversal) {
            t.Errorf("NormalizePath(%q) expected ErrPathTraversal, got %v", c, err)
        }
    }
}

func TestNormalizePathAbsolute(t *testing.T) {
    cases := []string{
        "/etc/passwd",
        "/home/user/file.txt",
        "C:\\Windows\\file.txt",
        "C:/Windows/file.txt",
        "c:/windows/file.txt",
    }
    
    for _, c := range cases {
        _, err := NormalizePath(c)
        if !errors.Is(err, ErrAbsolutePath) {
            t.Errorf("NormalizePath(%q) expected ErrAbsolutePath, got %v", c, err)
        }
    }
}

func TestNormalizePathEmpty(t *testing.T) {
    cases := []string{"", ".", "./", "/"}
    
    for _, c := range cases {
        _, err := NormalizePath(c)
        if err == nil {
            t.Errorf("NormalizePath(%q) expected error", c)
        }
    }
}
