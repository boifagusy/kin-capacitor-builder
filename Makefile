.PHONY: dev test build clean

dev:
go run ./cmd/app

test:
go test ./...

vet:
go vet ./...

build:
go build -o bin/local-apk-builder ./cmd/app

clean:
rm -rf bin/
rm -f ~/.local-apk-builder/builder.db

install:
go install ./cmd/app

run: dev
