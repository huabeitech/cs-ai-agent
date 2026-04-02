.PHONY: run run-server run-web build build-server build-web build-widget dist clean-dist test tidy generator enums migration

DIST_DIR ?= dist
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

build: build-web build-widget build-server

build-server:
	mkdir -p $(DIST_DIR)
	go build -o $(DIST_DIR)/$(APP_NAME) $(APP_ENTRY)

build-web:
	cd web && pnpm build

build-widget:
	cd widget && pnpm build:sdk && pnpm build

dist: clean-dist build
	mkdir -p $(DIST_DIR)/config
	mkdir -p $(DIST_DIR)/web
	mkdir -p $(DIST_DIR)/widget
	cp -R web/out $(DIST_DIR)/web/out
	cp -R widget/out $(DIST_DIR)/widget/out
	cp config/config.yaml $(DIST_DIR)/config/config.yaml
	cp config/config.example.yaml $(DIST_DIR)/config/config.example.yaml

clean-dist:
	rm -rf $(DIST_DIR)

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
