package generator

import (
    "strconv"
    "encoding/json"
    "fmt"
    "strings"
)

type SplashConfig struct {
    BackgroundColor string      `json:"background_color"`
    ShowLogo        bool        `json:"show_logo"`
    LoadingText     string      `json:"loading_text"`
    Duration        FlexibleInt `json:"duration"`
}

// FlexibleInt handles both number and string JSON values
type FlexibleInt int

func (f *FlexibleInt) UnmarshalJSON(data []byte) error {
    // Try number first
    var n int
    if err := json.Unmarshal(data, &n); err == nil {
        *f = FlexibleInt(n)
        return nil
    }
    // Try string
    var s string
    if err := json.Unmarshal(data, &s); err == nil {
        if n, err := strconv.Atoi(s); err == nil {
            *f = FlexibleInt(n)
            return nil
        }
    }
    *f = 0
    return nil
}

type NavItem struct {
    Icon      string `json:"icon"`
    URL       string `json:"url"`
    Animation string `json:"animation,omitempty"`
}

type NavConfig struct {
    Style  string    `json:"style"`
    Template string  `json:"template,omitempty"`
    Items  []NavItem `json:"items"`
}

func (g *Generator) GenerateWebIndex() string {
	// For upload projects, return the boot shell (navigates to ./app/index.html)
	if g.source != nil && g.source.SourceType == "upload" {
		return g.GenerateBootShell()
	}

    splash := SplashConfig{
        BackgroundColor: g.project.PrimaryColor,
        ShowLogo:        true,
        LoadingText:     "Loading...",
        Duration:        1500,
    }
    if g.project.SplashConfig != "" {
        json.Unmarshal([]byte(g.project.SplashConfig), &splash)
    }
    if int(splash.Duration) < 500 {
        splash.Duration = FlexibleInt(1500)
    }

    nav := NavConfig{
        Style: "bottom",
        Items: []NavItem{},
    }
    if g.project.OnboardingConfig != "" {
        json.Unmarshal([]byte(g.project.OnboardingConfig), &nav)
    }

    logoHTML := ""
    if splash.ShowLogo {
        if g.project.LogoPath != "" {
            // Use base64 logo directly in HTML for splash
            logoHTML = fmt.Sprintf(`<div class="splash-icon"><img src="%s" alt="Logo" style="width:96px;height:96px;border-radius:24px;object-fit:contain;"></div>`, g.project.LogoPath)
        } else {
            logoHTML = `<div class="splash-icon">📱</div>`
        }
    }

    navHTML := generateNavHTML(nav, g.project.PrimaryColor)

    return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no">
    <meta name="mobile-web-app-capable" content="yes">
    <meta name="apple-mobile-web-app-capable" content="yes">
    <title>%s</title>
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/font-awesome/6.5.1/css/all.min.css">
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        html, body { width: 100%%; height: 100%%; overflow: hidden; }
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; }

        #splash {
            position: fixed;
            top: 0; left: 0; right: 0; bottom: 0;
            z-index: 1000;
            display: flex;
            align-items: center;
            justify-content: center;
            flex-direction: column;
            text-align: center;
            background: %s;
            color: white;
            transition: opacity 0.5s;
        }
        #splash.hidden { opacity: 0; pointer-events: none; }
        .splash-icon { font-size: 4rem; margin-bottom: 1rem; }
        .splash-title { font-size: 1.5rem; font-weight: 700; }
        .splash-loading { margin-top: 1.5rem; font-size: 0.875rem; opacity: 0.8; }
        .spinner { width: 32px; height: 32px; border: 3px solid rgba(255,255,255,0.3); border-top-color: white; border-radius: 50%%; animation: spin 1s linear infinite; margin: 0 auto; }
        .skeleton-box { height: 20px; background: linear-gradient(90deg, #e0e0e0 25%%, #f0f0f0 50%%, #e0e0e0 75%%); background-size: 200px 100%%; animation: skeleton 1.5s infinite; border-radius: 4px; margin: 4px 0; }
        .dots { display: flex; gap: 6px; }
        .dots span { width: 8px; height: 8px; border-radius: 50%%; background: #6366f1; animation: pulse 1s infinite; }
        .dots span:nth-child(2) { animation-delay: 0.2s; }
        .dots span:nth-child(3) { animation-delay: 0.4s; }
        .progress-bar { width: 100%%; height: 4px; background: rgba(0,0,0,0.1); border-radius: 2px; overflow: hidden; }
        .progress-fill { width: 40%%; height: 100%%; background: #6366f1; animation: progress 1.5s ease-in-out infinite; }
        .fade-box { width: 32px; height: 32px; background: #6366f1; border-radius: 8px; animation: fadeInOut 1.5s ease-in-out infinite; }
        @keyframes progress { 0%% { transform: translateX(-100%%); } 100%% { transform: translateX(250%%); } }
        @keyframes fadeInOut { 0%%, 100%% { opacity: 0.3; } 50%% { opacity: 1; } }
        @keyframes skeleton { 0%% { background-position: -200px 0; } 100%% { background-position: 200px 0; } }
        @keyframes pulse { 0%%, 100%% { opacity: 1; } 50%% { opacity: 0.5; } }
        @keyframes spin { to { transform: rotate(360deg); } }
        .loading-overlay { position: fixed; top: 0; left: 0; right: 0; bottom: 0; z-index: 999; display: flex; align-items: center; justify-content: center; background: rgba(0,0,0,0.5); backdrop-filter: blur(4px); }
        .skeleton-box { width: 80%%; height: 20px; background: linear-gradient(90deg, #e0e0e0 25%%, #f0f0f0 50%%, #e0e0e0 75%%); background-size: 200px 100%%; animation: skeleton 1.5s infinite; border-radius: 4px; margin: 4px 0; }
        .dots { display: flex; gap: 6px; }
        .dots span { width: 8px; height: 8px; border-radius: 50%%; background: white; animation: pulse 1s infinite; }
        .dots span:nth-child(2) { animation-delay: 0.2s; }
        .dots span:nth-child(3) { animation-delay: 0.4s; }
        .progress-bar { width: 70%%; height: 4px; background: rgba(255,255,255,0.2); border-radius: 2px; overflow: hidden; }
        .progress-fill { width: 40%%; height: 100%%; background: white; animation: progress 1.5s ease-in-out infinite; }
        @keyframes progress { 0%% { transform: translateX(-100%%); } 100%% { transform: translateX(250%%); } }
        @keyframes spin { to { transform: rotate(360deg); } }


        %s
    </style>
</head>
<body>
    <!-- Navigation overlay -->
    %s

    <!-- Loading overlay -->
    <div id="loading-overlay" style="display:none;position:fixed;top:0;left:0;right:0;bottom:0;z-index:9999;background:%s;backdrop-filter:blur(4px);align-items:center;justify-content:center;flex-direction:column;">
        <div class="anim-spinner" style="display:none;"><div class="spinner"></div></div>
        <div class="anim-skeleton" style="display:none;width:80%%;"><div class="skeleton-box"></div><div class="skeleton-box" style="width:70%%;"></div><div class="skeleton-box" style="width:50%%;"></div></div>
        <div class="anim-dots" style="display:none;"><div class="dots"><span></span><span></span><span></span></div></div>
        <div class="anim-progress" style="display:none;width:70%%;"><div class="progress-bar"><div class="progress-fill"></div></div></div>
        <div class="anim-fade" style="display:none;"><div class="fade-box"></div></div>
    </div>

    <!-- Splash overlay -->
    <div id="splash">
        %s
        <div class="splash-title">%s</div>
        <div class="splash-loading"><div class="spinner"></div><p style="margin-top:0.75rem;">%s</p></div>
    </div>

    <script src="app_bridge.js"></script>
    <script>
        function showLoading(anim, url) {
            var overlay = document.getElementById('loading-overlay');
            overlay.style.display = 'flex';
            var types = ['spinner','skeleton','dots','progress','fade'];
            for (var i = 0; i < types.length; i++) {
                document.querySelector('.anim-' + types[i]).style.display = 'none';
            }
            var target = document.querySelector('.anim-' + (anim || 'spinner'));
            if (target) target.style.display = 'block';
            setTimeout(function() {
                window.location.href = url;
            }, 800);
            return false;
        }

        // Show loading indicator and navigate
        function navigateTo(url) {
            // Show loading overlay
            var overlay = document.getElementById('loading-overlay');
            overlay.style.display = 'flex';
            var spinner = document.querySelector('.anim-spinner');
            if(spinner) spinner.style.display = 'block';
            
            // Add timeout to detect loading failure
            var timeout = setTimeout(function() {
                // Still loading after 15 seconds - might be stuck
                var msg = document.querySelector('.anim-spinner');
                if(msg) msg.innerHTML = '<div style="text-align:center;"><p style="color:#666;font-size:14px;">Warning️ Still loading...</p><p style="color:#999;font-size:12px;margin-top:10px;">Check your connection</p></div>';
            }, 15000);
            
            // Navigate
            window.location.href = url;
        }
        
        setTimeout(function() {
            document.getElementById('splash').classList.add('hidden');
            navigateTo('%s');
        }, %d);
    </script>
</body>
</html>`, g.project.AppName, splash.BackgroundColor, navStyles(), navHTML, splash.BackgroundColor, logoHTML, g.project.AppName, splash.LoadingText, g.project.URL, splash.Duration)
}

func generateNavHTML(nav NavConfig, primaryColor string) string {
    if len(nav.Items) == 0 {
        return ""
    }

    if nav.Style == "bottom" {
        items := ""
        for _, item := range nav.Items {
            anim := item.Animation
            if anim == "" {
                anim = "spinner"
            }
            items += fmt.Sprintf(`<a href="%s" onclick="return showLoading('%s','%s');" style="flex:1;text-align:center;color:rgba(255,255,255,0.75);text-decoration:none;padding:0.5rem 0;transition:color 0.2s;"><i class="fas %s" style="font-size:1.25rem;color:%s;"></i><span style="display:block;font-size:0.65rem;margin-top:0.25rem;font-weight:500;">%s</span></a>`, item.URL, anim, item.URL, item.Icon, primaryColor, labelFromIcon(item.Icon))
        }
        return fmt.Sprintf(`<nav style="position:fixed;bottom:0;left:0;right:0;background:rgba(17,24,39,0.92);backdrop-filter:blur(12px);-webkit-backdrop-filter:blur(12px);display:flex;padding:0.5rem 0;z-index:2;border-top:1px solid rgba(255,255,255,0.1);box-shadow:0 -2px 10px rgba(0,0,0,0.15);">%s</nav>`, items)
    }

    // hamburger/drawer
    items := ""
    for _, item := range nav.Items {
        anim := item.Animation
        if anim == "" {
            anim = "spinner"
        }
        items += fmt.Sprintf(`<a href="%s" onclick="return showLoading('%s','%s');" style="display:flex;align-items:center;gap:0.75rem;padding:0.9rem 1.25rem;color:rgba(255,255,255,0.85);text-decoration:none;border-bottom:1px solid rgba(255,255,255,0.06);"><i class="fas %s" style="width:20px;text-align:center;color:%s;"></i> <span style="font-size:0.9rem;">%s</span></a>`, item.URL, anim, item.URL, item.Icon, primaryColor, labelFromIcon(item.Icon))
    }
    return fmt.Sprintf(`<div style="position:fixed;top:0;right:0;bottom:0;width:270px;background:rgba(17,24,39,0.96);backdrop-filter:blur(12px);-webkit-backdrop-filter:blur(12px);transform:translateX(100%%);transition:transform 0.3s cubic-bezier(0.4,0,0.2,1);z-index:2;box-shadow:-2px 0 15px rgba(0,0,0,0.25);" id="drawer">%s</div><button onclick="document.getElementById('drawer').style.transform='translateX(0)'" style="position:fixed;top:0.75rem;left:0.75rem;background:rgba(17,24,39,0.85);backdrop-filter:blur(8px);border:1px solid rgba(255,255,255,0.15);border-radius:50%%;color:white;font-size:1.1rem;z-index:2;width:40px;height:40px;display:flex;align-items:center;justify-content:center;cursor:pointer;box-shadow:0 2px 8px rgba(0,0,0,0.3);">Menu</button>`, items)
}
func labelFromIcon(icon string) string {
    parts := strings.Split(icon, "-")
    if len(parts) > 1 {
        return parts[len(parts)-1]
    }
    return icon
}

func navStyles() string {
    return ""
}

func generateLoadingAnimation(animationType string) string {
    switch animationType {
    case "spinner":
        return `<div class="loading-overlay"><div class="spinner"></div></div>`
    case "skeleton":
        return `<div class="loading-overlay"><div class="skeleton-box"></div><div class="skeleton-box"></div><div class="skeleton-box"></div></div>`
    case "dots":
        return `<div class="loading-overlay"><div class="dots"><span></span><span></span><span></span></div></div>`
    case "progress":
        return `<div class="loading-overlay"><div class="progress-bar"><div class="progress-fill"></div></div></div>`
    case "fade":
        return `<div class="loading-overlay" style="animation:fadeIn 0.5s;"></div>`
    default:
        return `<div class="loading-overlay"><div class="spinner"></div></div>`
    }
}



// InjectAppShell adds viewport, app bridge, and safe-area CSS to uploaded HTML
func InjectAppShell(html, appName string) string {
    // 1. Inject viewport meta if missing
    if !strings.Contains(html, `name="viewport"`) {
        viewportMeta := `<meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">`
        if strings.Contains(html, "<head>") {
            html = strings.Replace(html, "<head>", "<head>\n    "+viewportMeta, 1)
        } else if strings.Contains(html, "<html>") {
            html = strings.Replace(html, "<html>", "<html>\n<head>\n    "+viewportMeta+"\n</head>", 1)
        }
    }

    // 2. Inject app_bridge.js before </body> if missing
    if !strings.Contains(html, "app_bridge.js") {
        scriptTag := `<script src="app_bridge.js"></script>`
        if strings.Contains(html, "</body>") {
            html = strings.Replace(html, "</body>", "    "+scriptTag+"\n</body>", 1)
        } else {
            html += "\n" + scriptTag
        }
    }

    // 3. Add safe-area CSS
    safeAreaCSS := "<style>body{padding-top:env(safe-area-inset-top);padding-bottom:env(safe-area-inset-bottom);}html,body{overscroll-behavior:none;-webkit-tap-highlight-color:transparent;}</style>"
    if strings.Contains(html, "</head>") {
        html = strings.Replace(html, "</head>", safeAreaCSS+"\n</head>", 1)
    }

    return html
}
