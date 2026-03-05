.PHONY: build install release static clean test cover

VERSION ?= dev

build:
	go build -ldflags="-X main.version=$(VERSION)" -o sumi ./cmd/sumi

install: build
	install -Dm755 sumi ~/.local/bin/sumi

release:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o sumi ./cmd/sumi

static:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o sumi-static ./cmd/sumi

test:
	go test ./...

cover:
	go test -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

clean:
	rm -f sumi sumi-static coverage.out
