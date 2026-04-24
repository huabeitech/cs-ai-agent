.PHONY: help install run run-server run-web run-widget build build-all build-assets build-web build-widget \
	package-current package-platform build-linux-amd64 clean-dist clean-temp generator enums migration testdata

DIST_DIR ?= dist
TMP_DIR := $(DIST_DIR)/.tmp
APP_NAME ?= cs-agent
ARCHIVE_NAME ?= cs-ai-agent
PACKAGE_ROOT_DIR ?= $(ARCHIVE_NAME)
APP_ENTRY ?= ./cmd/server
PLATFORMS := linux-amd64 linux-arm64 darwin-amd64 darwin-arm64 windows-amd64
CURRENT_GOOS := $(shell go env GOOS)
CURRENT_GOARCH := $(shell go env GOARCH)
CURRENT_PLATFORM := $(CURRENT_GOOS)-$(CURRENT_GOARCH)
PLATFORM ?= $(CURRENT_PLATFORM)

help:
	@echo "Available targets:"
	@echo "  make help                 Show this help message"
	@echo "  make install              Install Go, web, and widget dependencies"
	@echo "  make run                  Run server, web, and widget"
	@echo "  make run-server           Run Go server"
	@echo "  make run-web              Run web app"
	@echo "  make run-widget           Run widget app"
	@echo "  make build                Build assets and package current platform"
	@echo "  make build-all            Build and package all platforms"
	@echo "  make build-assets         Build web and widget assets"
	@echo "  make build-web.           Build web app"
	@echo "  make build-widget         Build widget sdk and app"
	@echo "  make package-current      Package current platform"
	@echo "  make package-platform     Package specified PLATFORM"
	@echo "  make build-linux-amd64    Build and package linux amd64 release"
	@echo "  make clean-dist           Remove dist directory"
	@echo "  make clean-temp           Remove temporary packaging files"
	@echo "  make generator            Run code generator"
	@echo "  make enums                Generate frontend enums"
	@echo "  make migration            Run migration command"
	@echo "  make testdata             Run testdata generator"

install:
	go mod download
	cd web && pnpm install
	cd widget && pnpm install
	@echo "[install] done"

run:
	@server_pid=0; \
	web_pid=0; \
	widget_pid=0; \
	trap 'kill $$server_pid $$web_pid $$widget_pid 2>/dev/null || true' INT TERM EXIT; \
	$(MAKE) run-server & \
	server_pid=$$!; \
	$(MAKE) run-web & \
	web_pid=$$!; \
	$(MAKE) run-widget & \
	widget_pid=$$!; \
	while kill -0 $$server_pid 2>/dev/null && kill -0 $$web_pid 2>/dev/null && kill -0 $$widget_pid 2>/dev/null; do \
		sleep 1; \
	done; \
	status=0; \
	wait $$server_pid || status=$$?; \
	wait $$web_pid || status=$$?; \
	wait $$widget_pid || status=$$?; \
	kill $$server_pid $$web_pid $$widget_pid 2>/dev/null || true; \
	exit $$status

run-server:
	go run ./cmd/server

run-web:
	cd web && pnpm dev

run-widget:
	cd widget && pnpm dev

build: clean-dist build-assets package-current clean-temp
	@echo "[build] done"

build-all: clean-dist build-assets
	@echo "[build-all] packaging platforms: $(PLATFORMS)"
	@for platform in $(PLATFORMS); do \
		echo "[build-all] packaging $$platform"; \
		$(MAKE) package-platform PLATFORM=$$platform; \
	done
	@$(MAKE) clean-temp
	@echo "[build-all] done"

build-assets: build-web build-widget
	@echo "[build-assets] done"

build-web:
	@echo "[build-web] building web app"
	cd web && pnpm build
	@echo "[build-web] done"

build-widget:
	@echo "[build-widget] building widget sdk"
	cd widget && pnpm build:sdk && pnpm build
	@echo "[build-widget] done"

package-current:
	@$(MAKE) package-platform PLATFORM=$(CURRENT_PLATFORM)

build-linux-amd64: clean-dist build-assets
	@$(MAKE) package-platform PLATFORM=linux-amd64
	@$(MAKE) clean-temp
	@echo "[build-linux-amd64] done"

package-platform:
	@platform="$(PLATFORM)"; \
	goos=$${platform%-*}; \
	goarch=$${platform#*-}; \
	stage_dir="$(TMP_DIR)/$$platform"; \
	package_dir="$$stage_dir/$(PACKAGE_ROOT_DIR)"; \
	dist_dir="$(CURDIR)/$(DIST_DIR)"; \
	archive_base="$$dist_dir/$(ARCHIVE_NAME)-$$platform"; \
	binary_name="$(APP_NAME)"; \
	if [ "$$goos" = "windows" ]; then binary_name="$(APP_NAME).exe"; fi; \
	echo "[package] start $$platform"; \
	rm -rf "$$stage_dir"; \
	mkdir -p "$$package_dir/web" "$$package_dir/widget" "$$package_dir/config" "$$dist_dir"; \
	echo "[package] go build $$platform"; \
	GOOS=$$goos GOARCH=$$goarch go build -o "$$package_dir/$$binary_name" $(APP_ENTRY); \
	echo "[package] copy assets $$platform"; \
	cp -R web/out "$$package_dir/web/out"; \
	cp -R widget/out "$$package_dir/widget/out"; \
	echo "[package] include example config only"; \
	cp config/config.example.yaml "$$package_dir/config/config.example.yaml"; \
	rm -f "$$archive_base.zip"; \
	echo "[package] archive $$archive_base.zip"; \
	cd "$$stage_dir" && zip -rXq "$$archive_base.zip" "$(PACKAGE_ROOT_DIR)" -x "*.DS_Store" "__MACOSX/*"; \
	rm -rf "$$stage_dir"; \
	echo "[package] done $$platform -> $$archive_base.zip"

clean-dist:
	rm -rf $(DIST_DIR)

clean-temp:
	rm -rf $(TMP_DIR)

generator:
	go run ./cmd/generator/generator.go

enums:
	go run ./cmd/enums/generator.go

migration:
	go run ./cmd/migration

testdata:
	go run ./cmd/testdata
