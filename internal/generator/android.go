package generator

import (
    "fmt"
    "strings"
)

// GeneratePackageName generates a valid Android package name
func (g *Generator) GeneratePackageName() string {
    projectName := strings.ToLower(strings.TrimSpace(g.project.ProjectName))
    projectName = strings.Map(func(r rune) rune {
        if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
            return r
        }
        return -1
    }, projectName)
    return "com.localapkbuilder." + projectName
}

// GenerateAndroidManifest generates AndroidManifest.xml with proper SDK tags
func (g *Generator) GenerateAndroidManifest() string {
    
    return fmt.Sprintf(`<?xml version="1.0" encoding="utf-8"?>
<manifest xmlns:android="http://schemas.android.com/apk/res/android"
    android:versionCode="1"
    android:versionName="%s">

    <uses-sdk
        android:minSdkVersion="24"
        android:targetSdkVersion="30" />

    <uses-permission android:name="android.permission.INTERNET" />

    <application
        android:allowBackup="true"
        android:icon="@mipmap/ic_launcher"
        android:label="%s"
        android:supportsRtl="true"
        android:theme="@style/AppTheme"
        android:background="%s"
        android:usesCleartextTraffic="true">

        <activity
            android:name=".MainActivity"
            android:label="%s"
            android:exported="true"
            android:launchMode="singleTask"
            android:configChanges="orientation|keyboardHidden|screenSize|smallestScreenSize|locale|uiMode">
            <intent-filter>
                <action android:name="android.intent.action.MAIN" />
                <category android:name="android.intent.category.LAUNCHER" />
            </intent-filter>
        </activity>
    </application>
</manifest>`, g.project.Version, g.project.AppName, g.project.PrimaryColor, g.project.AppName)
}
