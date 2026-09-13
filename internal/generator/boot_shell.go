package generator

import (
    "encoding/json"
    "fmt"
)

// GenerateBootShell produces the boot HTML that Capacitor loads first.
// It shows Splash 2 (loading animation from splash_config), then navigates
// to ./app/index.html (the user's prepared application).
func (g *Generator) GenerateBootShell() string {
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

    logoHTML := ""
    if splash.ShowLogo {
        if g.project.LogoPath != "" {
            logoHTML = `<img src="app/logo.png" alt="Logo" style="width:96px;height:96px;border-radius:24px;object-fit:contain;">`
        } else {
            logoHTML = `<div style="font-size:4rem;">📱</div>`
        }
    }

    return fmt.Sprintf(`<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no, viewport-fit=cover">
    <title>%s</title>
    <style>
        * { margin: 0; padding: 0; box-sizing: border-box; }
        html, body { width: 100%%; height: 100%%; overflow: hidden; }
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; }
        #splash {
            position: fixed; top: 0; left: 0; right: 0; bottom: 0;
            z-index: 1000;
            display: flex; align-items: center; justify-content: center;
            flex-direction: column; text-align: center;
            background: %s; color: white;
            transition: opacity 0.3s;
        }
        .splash-icon { margin-bottom: 1rem; }
        .splash-title { font-size: 1.5rem; font-weight: 700; }
        .splash-loading { margin-top: 1.5rem; font-size: 0.875rem; opacity: 0.8; }
        .spinner { width: 32px; height: 32px; border: 3px solid rgba(255,255,255,0.3); border-top-color: white; border-radius: 50%%; animation: spin 1s linear infinite; margin: 0 auto; }
        @keyframes spin { to { transform: rotate(360deg); } }
    </style>
</head>
<body>
    <div id="splash">
        <div class="splash-icon">%s</div>
        <div class="splash-title">%s</div>
        <div class="splash-loading">
            <div class="spinner"></div>
            <p style="margin-top:0.75rem;">%s</p>
        </div>
    </div>
    <script>
        try { fetch('./app/index.html', { credentials: 'same-origin' }); } catch (e) {}
        var minDuration = %d;
        setTimeout(function() {
            document.getElementById('splash').style.opacity = '0';
            setTimeout(function() {
                window.location.replace('./app/index.html');
            }, 200);
        }, minDuration);
    </script>
</body>
</html>`, g.project.AppName, splash.BackgroundColor, logoHTML, g.project.AppName, splash.LoadingText, int(splash.Duration))
}
