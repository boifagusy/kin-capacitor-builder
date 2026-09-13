package main

import (
    "context"
    "fmt"
    "os"

    "local-apk-builder/internal/database"
    "local-apk-builder/internal/generator"
)

func main() {
    wsRoot := "/data/data/com.termux/files/home/test-prepare-workspace"
    ws := &database.Workspace{
        Root:     wsRoot,
        Source:   wsRoot + "/source",
        Prepared: wsRoot + "/prepared",
        Build:    wsRoot + "/build",
    }
    ws.Ensure()

    project := &database.Project{ID: 99, AppName: "Prepare Test"}
    adapter := &generator.UploadSourceAdapter{Project: project, Workspace: ws}

    prepared, err := adapter.Prepare(context.Background())
    if err != nil {
        fmt.Printf("Error: %v\n", err)
        os.Exit(1)
    }
    fmt.Printf("Prepared: %+v\n\n", prepared)

    data, _ := os.ReadFile(wsRoot + "/prepared/index.html")
    fmt.Println("=== prepared/index.html ===")
    fmt.Println(string(data))
}
