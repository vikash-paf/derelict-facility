# Makefile for Derelict Facility (Retro Sci-Fi Game Engine)

.PHONY: all build run test clean rerun

all: build

build:
	CGO_ENABLED=1 go build -ldflags="-s -w" -o derelict.exe ./cmd/derelict




run: build
	./derelict.exe

rerun: clean run

test:
	go test ./internal/...

clean:
	rm -f derelict.exe

