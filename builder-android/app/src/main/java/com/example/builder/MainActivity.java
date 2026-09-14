package com.example.builder;

import android.content.Intent;
import android.os.Build;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import android.view.View;
import android.webkit.WebSettings;
import android.webkit.WebView;
import android.webkit.WebViewClient;
import android.widget.LinearLayout;
import android.widget.TextView;

import androidx.appcompat.app.AppCompatActivity;
import androidx.core.content.ContextCompat;

import org.json.JSONObject;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.net.HttpURLConnection;
import java.net.URL;

public class MainActivity extends AppCompatActivity {
    private static final String BASE_URL    = "http://127.0.0.1:18791";
    private static final String BUILDER_URL = BASE_URL + "/dashboard";
    private static final String WELCOME_URL = BASE_URL + "/welcome";
    private static final String HEALTH_URL  = BASE_URL + "/api/health";
    private static final String STATUS_URL  = BASE_URL + "/api/github/status";
    private WebView webView;
    private LinearLayout overlay;
    private TextView loadingText;
    private TextView loadingHint;

    @Override
    protected void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        setContentView(R.layout.activity_main);

        webView = findViewById(R.id.webview);
        overlay = findViewById(R.id.loading_overlay);
        loadingText = findViewById(R.id.loading_text);
        loadingHint = findViewById(R.id.loading_hint);

        WebSettings ws = webView.getSettings();
        ws.setJavaScriptEnabled(true);
        ws.setDomStorageEnabled(true);
        webView.setWebViewClient(new WebViewClient() {
            @Override
            public void onPageFinished(WebView v, String url) {
                super.onPageFinished(v, url);
                overlay.setVisibility(View.GONE);
            }
        });

        Intent svc = new Intent(this, BuilderService.class);
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            ContextCompat.startForegroundService(this, svc);
        } else {
            startService(svc);
        }

        waitForServer(40);
    }

    private void waitForServer(final int attemptsLeft) {
        if (attemptsLeft <= 0) {
            loadingText.setText("Server failed to start");
            loadingHint.setText("Please close and reopen the app.");
            return;
        }
        loadingText.setText("Starting server...");
        loadingHint.setText("Attempt " + (41 - attemptsLeft) + " / 40");
        new Thread(new Runnable() {
            public void run() {
                boolean ok = false;
                try {
                    HttpURLConnection c = (HttpURLConnection) new URL(HEALTH_URL).openConnection();
                    c.setConnectTimeout(500);
                    c.setReadTimeout(500);
                    ok = (c.getResponseCode() == 200);
                } catch (Throwable ignored) {}
                final boolean success = ok;
                new Handler(Looper.getMainLooper()).post(new Runnable() {
                    public void run() {
                        if (success) {
                            loadingText.setText("Checking credentials...");
                            loadingHint.setText("");
                            checkCredentialsAndLoad();
                        } else {
                            new Handler(Looper.getMainLooper()).postDelayed(new Runnable() {
                                public void run() { waitForServer(attemptsLeft - 1); }
                            }, 500);
                        }
                    }
                });
            }
        }).start();
    }

    private void checkCredentialsAndLoad() {
        new Thread(new Runnable() {
            public void run() {
                String target = BUILDER_URL;
                try {
                    HttpURLConnection c = (HttpURLConnection) new URL(STATUS_URL).openConnection();
                    c.setConnectTimeout(1000);
                    c.setReadTimeout(1000);
                    if (c.getResponseCode() == 200) {
                        BufferedReader r = new BufferedReader(new InputStreamReader(c.getInputStream()));
                        StringBuilder sb = new StringBuilder();
                        String line;
                        while ((line = r.readLine()) != null) sb.append(line);
                        r.close();
                        JSONObject obj = new JSONObject(sb.toString());
                        boolean connected = obj.optBoolean("connected", false);
                        if (!connected) {
                            target = WELCOME_URL;
                        }
                    }
                } catch (Throwable ignored) {}
                final String finalTarget = target;
                new Handler(Looper.getMainLooper()).post(new Runnable() {
                    public void run() {
                        loadingText.setText("Loading...");
                        webView.loadUrl(finalTarget);
                    }
                });
            }
        }).start();
    }

    @Override
    public void onBackPressed() {
        if (webView != null && webView.canGoBack()) {
            webView.goBack();
        } else {
            super.onBackPressed();
        }
    }
}
