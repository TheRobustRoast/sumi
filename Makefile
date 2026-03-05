.PHONY: build install static clean test cover

VERSION ?= dev

# Default build: static Linux binary (committed to repo for ISO use)
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o sumi ./cmd/sumi

# Build for local dev on current OS (not committed)
dev:
	go build -ldflags="-X main.version=$(VERSION)" -o sumi ./cmd/sumi

install: build
	install -Dm755 sumi ~/.local/bin/sumi

test:
	go test ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	rm -f sumi sumi.exe coverage.out
