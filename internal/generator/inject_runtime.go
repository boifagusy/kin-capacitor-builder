package generator

import "strings"

// InjectRuntimeHooks injects required runtime hooks into the user's HTML
// without replacing or modifying the body content.
func InjectRuntimeHooks(html string) string {
    // 1. Inject viewport meta if missing
    if !strings.Contains(html, `name="viewport"`) {
        viewport := `<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">`
        idx := strings.Index(html, "<head>")
        if idx >= 0 {
            pos := idx + len("<head>")
            html = html[:pos] + "\n    " + viewport + html[pos:]
        }
    }

    // 2. Inject safe-area CSS before </head>
    if !strings.Contains(html, "env(safe-area-inset-top)") {
        safeCSS := `<style id="kin-safe-area">body{padding-top:env(safe-area-inset-top);padding-bottom:env(safe-area-inset-bottom);}html,body{overscroll-behavior:none;-webkit-tap-highlight-color:transparent;}</style>`
        idx := strings.Index(html, "</head>")
        if idx >= 0 {
            html = html[:idx] + safeCSS + "\n" + html[idx:]
        }
    }

    // 3. Inject app_bridge.js before </body> if missing
    if !strings.Contains(html, "app_bridge.js") {
        bridge := `<script src="../app_bridge.js"></script>`
        idx := strings.Index(html, "</body>")
        if idx >= 0 {
            html = html[:idx] + "    " + bridge + "\n" + html[idx:]
        } else {
            html = html + "\n" + bridge
        }
    }

    return html
}
