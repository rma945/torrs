#!/bin/sh

echo "Generate static"
go run ./cmd/genpages/gen_pages.go

echo "Build..."
CGO_ENABLED=${CGO_ENABLED:-0} GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH:-amd64} go build -v -ldflags='-s -w' -o ./dist/torrs ./cmd/main
