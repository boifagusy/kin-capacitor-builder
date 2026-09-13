package generator

import (
    "encoding/json"
    "strings"
)

// PluginMap maps feature keys to Capacitor packages
var PluginMap = map[string]string{
    "app":                 "@capacitor/app",
    "browser":             "@capacitor/browser",
    "camera":              "@capacitor/camera",
    "geolocation":         "@capacitor/geolocation",
    "push-notifications":  "@capacitor/push-notifications",
    "filesystem":          "@capacitor/filesystem",
    "share":               "@capacitor/share",
    "preferences":         "@capacitor/preferences",
    "device":              "@capacitor/device",
    "network":             "@capacitor/network",
    "haptics":             "@capacitor/haptics",
    "clipboard":           "@capacitor/clipboard",
    "status-bar":          "@capacitor/status-bar",
    "splash-screen":       "@capacitor/splash-screen",
    "local-notifications": "@capacitor/local-notifications",
    "screen-reader":       "@capacitor/screen-reader",
    "text-zoom":           "@capacitor/text-zoom",
    "calendar":            "@ebarooni/capacitor-calendar",
}

// GetPluginDependencies returns map of package: version for selected features
func GetPluginDependencies(featuresConfig string) map[string]string {
    deps := map[string]string{
        "@capacitor/core":             "^8.5.1",
        "@capacitor/android":          "^8.5.1",
        "@capacitor/ios":              "^8.5.1",
        "@capacitor/cli":              "^8.5.1",
        // Default plugins (always included)
        "@capacitor/splash-screen":    "^8.0.2",
        "@capacitor/app":              "^8.0.2",
        "@capacitor/preferences":      "^8.0.1",
        "@capacitor/network":          "^8.0.1",
    }

    features := ParseFeatures(featuresConfig)
    for _, f := range features {
        if pkg, ok := PluginMap[f]; ok {
            deps[pkg] = "^8.0.0"
        }
    }
    return deps
}

// ParseFeatures parses features_config into a slice of feature keys
func ParseFeatures(featuresConfig string) []string {
    if featuresConfig == "" || featuresConfig == "[]" {
        return nil
    }
    var arr []string
    if err := json.Unmarshal([]byte(featuresConfig), &arr); err == nil {
        return arr
    }
    parts := []string{}
    for _, p := range splitAndTrim(featuresConfig, ",") {
        if p != "" {
            parts = append(parts, p)
        }
    }
    return parts
}

func splitAndTrim(s, sep string) []string {
    var result []string
    start := 0
    for i := 0; i <= len(s); i++ {
        if i == len(s) || s[i] == sep[0] {
            part := s[start:i]
            part = trimSpace(part)
            result = append(result, part)
            start = i + 1
        }
    }
    return result
}

func trimSpace(s string) string {
    return strings.TrimSpace(s)
}
