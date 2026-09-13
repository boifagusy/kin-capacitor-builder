package com.example.builder;

import android.app.Notification;
import android.app.NotificationChannel;
import android.app.NotificationManager;
import android.app.Service;
import android.content.Intent;
import android.os.Build;
import android.os.IBinder;

import mobile.androidlib.Androidlib;

public class BuilderService extends Service {
    private static final String CHANNEL_ID = "builder_service";
    private static final int NOTIFICATION_ID = 1;

    @Override
    public int onStartCommand(Intent intent, int flags, int startId) {
        createNotificationChannel();
        Notification n = new Notification.Builder(this, CHANNEL_ID)
                .setContentTitle("Local APK Builder")
                .setContentText("Builder server is running")
                .setSmallIcon(android.R.drawable.stat_sys_download)
                .build();
        startForeground(NOTIFICATION_ID, n);

        try {
            String dataDir = getFilesDir().getAbsolutePath();
            Androidlib.start(dataDir, 18791);
        } catch (Exception e) {
            e.printStackTrace();
        }
        return START_STICKY;
    }

    @Override
    public void onDestroy() {
        try {
            Androidlib.stop();
        } catch (Exception e) {
            e.printStackTrace();
        }
        super.onDestroy();
    }

    @Override
    public IBinder onBind(Intent intent) {
        return null;
    }

    private void createNotificationChannel() {
        if (Build.VERSION.SDK_INT >= Build.VERSION_CODES.O) {
            NotificationChannel ch = new NotificationChannel(
                    CHANNEL_ID,
                    "Builder",
                    NotificationManager.IMPORTANCE_LOW);
            NotificationManager nm = getSystemService(NotificationManager.class);
            if (nm != null) nm.createNotificationChannel(ch);
        }
    }
}
