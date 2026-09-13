package generator

import "fmt"

// GenerateNativeMainActivityJava returns a MainActivity that registers
// the AppBridge plugin and loads the WebView from Capacitor assets.
// Used when project.RuntimeType == "go-native".
func (g *Generator) GenerateNativeMainActivityJava() string {
	return `package com.localapkbuilder;

import android.os.Bundle;
import com.getcapacitor.BridgeActivity;

public class MainActivity extends BridgeActivity {
    @Override
    public void onCreate(Bundle savedInstanceState) {
        registerPlugin(AppBridgePlugin.class);
        super.onCreate(savedInstanceState);
    }
}
`
}

// GenerateNativeAppBridgePlugin returns the Java source for the Capacitor
// plugin that bridges the WebView to the Go runtime via JNI.
// Used when project.RuntimeType == "go-native".
func (g *Generator) GenerateNativeAppBridgePlugin() string {
	return `package com.localapkbuilder;

import android.util.Log;
import com.getcapacitor.JSObject;
import com.getcapacitor.Plugin;
import com.getcapacitor.PluginCall;
import com.getcapacitor.PluginMethod;
import com.getcapacitor.annotation.CapacitorPlugin;

import capgotest.AppCore;
import capgotest.Capgotest;

@CapacitorPlugin(name = "AppBridge")
public class AppBridgePlugin extends Plugin {
    private static final String TAG = "AppBridge";
    private AppCore core;

    @Override
    public void load() {
        try {
            String dbPath = getContext().getFilesDir().getAbsolutePath() + "/runtime.db";
            core = Capgotest.newAppCore(dbPath);
            Log.i(TAG, "AppCore initialized at " + dbPath);
        } catch (Exception e) {
            Log.e(TAG, "Failed to initialize AppCore", e);
        }
    }

    @PluginMethod
    public void health(PluginCall call) {
        JSObject ret = new JSObject();
        try {
            String h = (core == null) ? "no-core" : core.health();
            ret.put("value", h);
        } catch (Exception e) {
            ret.put("value", "error: " + e.getMessage());
        }
        call.resolve(ret);
    }

    @PluginMethod
    public void saveNote(PluginCall call) {
        JSObject ret = new JSObject();
        try {
            String text = call.getString("text", "");
            long id = (core == null) ? -1 : core.saveNote(text);
            ret.put("id", id);
            ret.put("text", text);
        } catch (Exception e) {
            ret.put("id", -1);
            ret.put("error", e.getMessage());
        }
        call.resolve(ret);
    }

    @PluginMethod
    public void getNotes(PluginCall call) {
        JSObject ret = new JSObject();
        try {
            String json = (core == null) ? "[]" : core.getNotes();
            ret.put("notes", json);
        } catch (Exception e) {
            ret.put("notes", "[]");
            ret.put("error", e.getMessage());
        }
        call.resolve(ret);
    }
}
`
}

// GenerateNativeAppBridgeJS returns a JavaScript AppBridge that calls
// the Capacitor AppBridge plugin. Replaces the classic app_bridge.js.
// Used when project.RuntimeType == "go-native".
func (g *Generator) GenerateNativeAppBridgeJS() string {
	appID := fmt.Sprintf("app%d", g.project.ID)
	_ = appID
	return `window.AppBridge = {
    async health() {
        try {
            const { AppBridge } = Capacitor.Plugins;
            const r = await AppBridge.health();
            return r.value || "unknown";
        } catch (e) {
            return "error: " + e.message;
        }
    },

    async saveNote(text) {
        try {
            const { AppBridge } = Capacitor.Plugins;
            const r = await AppBridge.saveNote({ text: text });
            return r;
        } catch (e) {
            return { id: -1, error: e.message };
        }
    },

    async getNotes() {
        try {
            const { AppBridge } = Capacitor.Plugins;
            const r = await AppBridge.getNotes();
            return JSON.parse(r.notes || "[]");
        } catch (e) {
            return [];
        }
    }
};
`
}
