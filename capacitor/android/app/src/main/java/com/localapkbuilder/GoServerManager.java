package com.localapkbuilder;

import android.content.Context;
import android.util.Log;
import java.io.File;
import java.io.FileOutputStream;
import java.io.InputStream;
import java.io.IOException;

public class GoServerManager {
    private static final String TAG = "GoServerManager";
    private static Process goProcess;
    private static File goBinary;
    
    public static void startServer(Context context) {
        if (goProcess != null) {
            Log.i(TAG, "Server already running");
            return;
        }
        
        try {
            goBinary = extractBinary(context);
            
            if (goBinary == null || !goBinary.exists()) {
                Log.e(TAG, "Failed to extract Go binary");
                return;
            }
            
            goBinary.setExecutable(true);
            
            File dataDir = new File(context.getFilesDir(), "data");
            dataDir.mkdirs();
            
            ProcessBuilder pb = new ProcessBuilder(
                goBinary.getAbsolutePath(),
                "--port", "8080",
                "--data-dir", dataDir.getAbsolutePath()
            );
            pb.redirectErrorStream(true);
            goProcess = pb.start();
            
            Log.i(TAG, "Go server started with PID: " + goProcess.pid());
            
            // Wait for server to be ready
            Thread.sleep(1500);
            
        } catch (Exception e) {
            Log.e(TAG, "Failed to start Go server", e);
        }
    }
    
    public static void stopServer() {
        if (goProcess != null) {
            goProcess.destroy();
            goProcess = null;
            Log.i(TAG, "Go server stopped");
        }
    }
    
    private static File extractBinary(Context context) {
        try {
            File binaryFile = new File(context.getFilesDir(), "local-apk-builder");
            
            if (binaryFile.exists() && binaryFile.length() > 0) {
                return binaryFile;
            }
            
            InputStream is = context.getAssets().open("go/local-apk-builder");
            FileOutputStream os = new FileOutputStream(binaryFile);
            
            byte[] buffer = new byte[8192];
            int length;
            while ((length = is.read(buffer)) > 0) {
                os.write(buffer, 0, length);
            }
            
            os.close();
            is.close();
            
            Log.i(TAG, "Binary extracted: " + binaryFile.length() + " bytes");
            return binaryFile;
            
        } catch (IOException e) {
            Log.e(TAG, "Failed to extract binary", e);
            return null;
        }
    }
}
