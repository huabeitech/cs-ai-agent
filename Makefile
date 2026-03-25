.PHONY: run run-server run-web build test tidy generator enums migration

BIN_DIR ?= bin
APP_NAME ?= cs-agent
APP_ENTRY ?= ./cmd/server

run:
	@server_pid=0; \
	web_pid=0; \
	trap 'kill $$server_pid $$web_pid 2>/dev/null || true' INT TERM EXIT; \
	$(MAKE) run-server & \
	server_pid=$$!; \
	$(MAKE) run-web & \
	web_pid=$$!; \
	while kill -0 $$server_pid 2>/dev/null && kill -0 $$web_pid 2>/dev/null; do \
		sleep 1; \
	done; \
	status=0; \
	wait $$server_pid || status=$$?; \
	wait $$web_pid || status=$$?; \
	kill $$server_pid $$web_pid 2>/dev/null || true; \
	exit $$status

run-server:
	go run ./cmd/server

run-web:
	cd web && pnpm dev

build:
	mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) $(APP_ENTRY)

test:
	go test ./...

tidy:
	go mod tidy

generator:
	go run ./cmd/generator/generator.go

enums:
	go run ./cmd/enums/generator.go

migration:
	go run ./cmd/migration

testdata:
	go run ./cmd/testdata