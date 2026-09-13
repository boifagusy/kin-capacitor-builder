package generator

import "fmt"

// GenerateRuntimeMainActivityJava builds a plain Activity that starts the
// Go runtime, exposes the auth token to JS, and loads the WebView at
// http://127.0.0.1:PORT/. Used only when project.RuntimeType == "go-sqlite".
func (g *Generator) GenerateRuntimeMainActivityJava() string {
appID := fmt.Sprintf("com.localapkbuilder.app%d", g.project.ID)

return fmt.Sprintf(`package %s;

import android.app.Activity;
import android.os.Bundle;
import android.util.Log;
import android.webkit.ConsoleMessage;
import android.webkit.JavascriptInterface;
import android.webkit.WebChromeClient;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;

import runtime.Runtime;
import runtime.Server;

public class MainActivity extends Activity {
    private WebView webView;
    private Server srv;
    private String authToken;

    public class TokenBridge {
        @JavascriptInterface
        public String getToken() {
            return authToken == null ? "" : authToken;
        }
    }

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        Log.i("Runtime", "onCreate reached");

        String dbPath = getFilesDir().getAbsolutePath() + "/runtime.db";
        Log.i("Runtime", "dbPath=" + dbPath);

        try {
            srv = Runtime.newServer(dbPath);
            Log.i("Runtime", "server created");
        } catch (Throwable t) {
            Log.e("Runtime", "Runtime.newServer failed", t);
            return;
        }

        long port = srv.start();
        Log.i("Runtime", "port=" + port);

        if (port <= 0) {
            Log.e("Runtime", "server did not start");
            return;
        }

        authToken = srv.token();
        Log.i("Runtime", "token_len=" + (authToken == null ? 0 : authToken.length()));

        webView = new WebView(this);
        webView.setWebViewClient(new WebViewClient());
        webView.setWebChromeClient(new WebChromeClient() {
            @Override
            public boolean onConsoleMessage(ConsoleMessage msg) {
                Log.i("Runtime", msg.message());
                return true;
            }
        });
        webView.addJavascriptInterface(new TokenBridge(), "AndroidBridge");

        WebSettings s = webView.getSettings();
        s.setJavaScriptEnabled(true);
        s.setDomStorageEnabled(true);
        s.setAllowFileAccess(true);
        s.setMixedContentMode(WebSettings.MIXED_CONTENT_ALWAYS_ALLOW);

        webView.loadUrl("http://127.0.0.1:" + port + "/");
        setContentView(webView);
    }

    @Override
    protected void onDestroy() {
        super.onDestroy();
        if (srv != null) {
            srv.stop();
            Log.i("Runtime", "server stopped");
        }
    }
}
`, appID)
}

// GenerateRuntimeAppBridgeJS returns an extended AppBridge that talks to
// the local Go runtime over HTTP. Used only when project.RuntimeType == "go-sqlite".
func (g *Generator) GenerateRuntimeAppBridgeJS() string {
return `window.AppBridge = {
    config: null,
    authToken: null,

    async init() {
        try {
            const res = await fetch('/app_config.json');
            this.config = await res.json();
        } catch(e) {}

        if (window.AndroidBridge && window.AndroidBridge.getToken) {
            try {
                this.authToken = window.AndroidBridge.getToken();
            } catch(e) {}
        }
    },

    async api(path, opts) {
        opts = opts || {};
        const headers = Object.assign({}, opts.headers || {});
        headers['Content-Type'] = headers['Content-Type'] || 'application/json';
        if (this.authToken) {
            headers['Authorization'] = 'Bearer ' + this.authToken;
        }
        const res = await fetch(path, Object.assign({}, opts, { headers }));
        const ct = res.headers.get('content-type') || '';
        if (ct.indexOf('application/json') >= 0) {
            return res.json();
        }
        return res.text();
    },

    async get(path) {
        return this.api(path, { method: 'GET' });
    },

    async post(path, body) {
        return this.api(path, { method: 'POST', body: JSON.stringify(body) });
    },

    async del(path) {
        return this.api(path, { method: 'DELETE' });
    }
};

if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', () => window.AppBridge.init());
} else {
    window.AppBridge.init();
}
`
}
