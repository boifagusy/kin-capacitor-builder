package main

import (
    "fmt"
    "local-apk-builder/internal/database"
    "local-apk-builder/internal/generator"
)

func main() {
    project := &database.Project{
        ID:           3,
        AppName:      "Test App",
        PrimaryColor: "#10b981",
        SplashConfig: `{"background_color":"#10b981","duration":"2000","loading_text":"Starting up...","show_logo":true}`,
        LogoPath:     "data:image/png;base64,iVBORw0KGgo=",
    }
    gen := generator.NewGenerator(project)
    fmt.Println(gen.GenerateBootShell())
}
