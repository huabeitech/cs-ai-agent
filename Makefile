.PHONY: run run-server run-web build build-all build-assets build-web build-widget \
	package-current package-platform clean-dist clean-temp test tidy generator enums migration testdata

DIST_DIR ?= dist
TMP_DIR := $(DIST_DIR)/.tmp
APP_NAME ?= cs-agent
APP_ENTRY ?= ./cmd/server
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64
CURRENT_GOOS := $(shell go env GOOS)
CURRENT_GOARCH := $(shell go env GOARCH)
CURRENT_PLATFORM := $(CURRENT_GOOS)-$(CURRENT_GOARCH)
PLATFORM ?= $(CURRENT_PLATFORM)

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

build: clean-dist build-assets package-current clean-temp

build-all: clean-dist build-assets
	@for platform in $(PLATFORMS); do \
		$(MAKE) package-platform PLATFORM=$$platform; \
	done
	@$(MAKE) clean-temp

build-assets: build-web build-widget

build-web:
	cd web && pnpm build

build-widget:
	cd widget && pnpm build:sdk && pnpm build

package-current:
	@$(MAKE) package-platform PLATFORM=$(CURRENT_PLATFORM)

package-platform:
	@platform="$(PLATFORM)"; \
	goos=$${platform%-*}; \
	goarch=$${platform#*-}; \
	stage_dir="$(TMP_DIR)/$$platform"; \
	archive_base="$$(cd "$(DIST_DIR)" && pwd)/$(APP_NAME)-$$platform"; \
	binary_name="$(APP_NAME)"; \
	if [ "$$goos" = "windows" ]; then binary_name="$(APP_NAME).exe"; fi; \
	rm -rf "$$stage_dir"; \
	mkdir -p "$$stage_dir/web" "$$stage_dir/widget" "$$stage_dir/config" "$(DIST_DIR)"; \
	GOOS=$$goos GOARCH=$$goarch go build -o "$$stage_dir/$$binary_name" $(APP_ENTRY); \
	cp -R web/out "$$stage_dir/web/out"; \
	cp -R widget/out "$$stage_dir/widget/out"; \
	cp config/config.yaml "$$stage_dir/config/config.yaml"; \
	cp config/config.example.yaml "$$stage_dir/config/config.example.yaml"; \
	if [ "$$goos" = "windows" ]; then \
		rm -f "$$archive_base.zip"; \
		cd "$$stage_dir" && zip -rq "$$archive_base.zip" .; \
	else \
		rm -f "$$archive_base.tar.gz"; \
		tar -C "$$stage_dir" -czf "$$archive_base.tar.gz" .; \
	fi; \
	rm -rf "$$stage_dir"

clean-dist:
	rm -rf $(DIST_DIR)

clean-temp:
	rm -rf $(TMP_DIR)

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
