package main

import (
"fmt"
"local-apk-builder/internal/generator"
)

func main() {
input := `<!DOCTYPE html>
<html>
<head>
    <title>Test Static App</title>
    <link rel="stylesheet" href="css/app.css">
</head>
<body>
    <h1>Hello from Uploaded ZIP!</h1>
    <img src="images/logo.png">
    <script src="js/app.js"></script>
</body>
</html>`

output := generator.InjectRuntimeHooks(input)
fmt.Println(output)
}
