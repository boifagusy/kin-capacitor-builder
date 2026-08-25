# Local APK Builder

A fully local, APK-only app builder that runs entirely on-device.

## Requirements

- Go 1.21 or later
- Git
- curl (for initial setup)

## Setup (Termux)

Clone repository:
  git clone <repository-url>
  cd local-apk-builder

Download dependencies:
  go mod download

## Run

Run server:
  go run ./cmd/app

Or using make:
  make dev

The server will start at: http://127.0.0.1:8080

## Test

Run tests:
  go test ./...

Run vet:
  go vet ./...

Build:
  go build ./cmd/app

## Verify

Check server is running:
  curl http://127.0.0.1:8080/

Check API:
  curl http://127.0.0.1:8080/api/project-state

## Architecture

Go application
  -> HTTP server (127.0.0.1)
  -> HTMX + Alpine.js UI
  -> SQLite
  -> Android WebView

## Database

SQLite database stored at: ~/.local-apk-builder/builder.db

## Security

- Server binds to 127.0.0.1 only
- Input validation on all user inputs
- Parameterized SQL queries
- No external dependencies
