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

import java.net.HttpURLConnection;
import java.net.URL;

public class MainActivity extends AppCompatActivity {
    private static final String BUILDER_URL = "http://127.0.0.1:18791/";
    private static final String HEALTH_URL  = "http://127.0.0.1:18791/api/health";
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
                            loadingText.setText("Loading dashboard...");
                            loadingHint.setText("");
                            webView.loadUrl(BUILDER_URL);
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

    @Override
    public void onBackPressed() {
        if (webView != null && webView.canGoBack()) {
            webView.goBack();
        } else {
            super.onBackPressed();
        }
    }
}
