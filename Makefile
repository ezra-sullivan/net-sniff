# 项目信息
PROJECT_NAME := net-sniff
MAIN_PATH := ./main.go
BUILD_DIR := ./build
VERSION := 0.1.1

# Go 命令
GO := go
GOBUILD := $(GO) build
GOCLEAN := $(GO) clean
GOTEST := $(GO) test
GOGET := $(GO) get

# Supported operating systems and architectures
PLATFORMS := windows linux darwin
ARCHITECTURES := amd64 arm64

# 兼容 Win 构建
# Detect operating system
ifeq ($(OS),Windows_NT)
    # Windows command fixes
    RM = powershell -NoProfile -ExecutionPolicy Bypass -Command "if (Test-Path '$(BUILD_DIR)') { Remove-Item -Recurse -Force '$(BUILD_DIR)' }"
    MKDIR = powershell -NoProfile -ExecutionPolicy Bypass -Command "if (-not (Test-Path '$(BUILD_DIR)')) { New-Item -ItemType Directory -Path '$(BUILD_DIR)' }"
    MKDIR_RELEASE = powershell -NoProfile -ExecutionPolicy Bypass -Command "if (-not (Test-Path '$(BUILD_DIR)\release')) { New-Item -ItemType Directory -Path '$(BUILD_DIR)\release' }"
    CHECK_FILE = if exist
    ZIP = powershell -NoProfile -ExecutionPolicy Bypass -Command "Compress-Archive -Force -Path"
    ZIP_ARGS = -DestinationPath
    # Use cmd to output text, avoid PowerShell encoding issues
    ECHO = cmd /c echo
    # Set PowerShell execution environment, allow interruption
    PS_ENV = powershell -NoProfile -ExecutionPolicy Bypass -Command
else
    RM := rm -rf $(BUILD_DIR)
    MKDIR := mkdir -p $(BUILD_DIR)
    MKDIR_RELEASE := mkdir -p $(BUILD_DIR)/release
    CHECK_FILE := if [ -f
    ZIP := zip -j
    ZIP_ARGS :=
    ECHO := echo
    PS_ENV := 
endif

# Default target
.PHONY: all
all: clean build

# Clean build directory
.PHONY: clean
clean:
	@$(ECHO) "Cleaning build directory..."
	-@$(RM)
	-@$(MKDIR)
	@$(GOCLEAN)

# Build executable for current platform
.PHONY: build
build: clean
	@$(ECHO) "Building executable for current platform..."
	$(eval OS := $(shell go env GOOS))
	$(eval ARCH := $(shell go env GOARCH))
	$(eval SUFFIX := $(if $(findstring windows,$(OS)),.exe,))
	@$(GOBUILD) -o $(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_$(OS)_$(ARCH)$(SUFFIX) $(MAIN_PATH)
	@$(ECHO) "Build completed: $(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_$(OS)_$(ARCH)$(SUFFIX)"

# Build executables for all platforms
.PHONY: build-all
build-all: clean
	@$(ECHO) "Building executables for all platforms..."
	-@$(MKDIR)
	@$(MAKE) build-windows-amd64
	@$(MAKE) build-windows-arm64
	@$(MAKE) build-linux-amd64
	@$(MAKE) build-linux-arm64
	@$(MAKE) build-darwin-amd64
	@$(MAKE) build-darwin-arm64
	
# Windows platform builds
.PHONY: build-windows-amd64
build-windows-amd64:
	@$(ECHO) "Building windows/amd64..."
	@$(PS_ENV) "$$env:GOOS='windows'; $$env:GOARCH='amd64'; go build -o '$(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_windows_amd64.exe' $(MAIN_PATH)"

.PHONY: build-windows-arm64
build-windows-arm64:
	@$(ECHO) "Building windows/arm64..."
	@$(PS_ENV) "$$env:GOOS='windows'; $$env:GOARCH='arm64'; go build -o '$(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_windows_arm64.exe' $(MAIN_PATH)"

# Linux platform builds
.PHONY: build-linux-amd64
build-linux-amd64:
	@$(ECHO) "Building linux/amd64..."
	@$(PS_ENV) "$$env:GOOS='linux'; $$env:GOARCH='amd64'; go build -o '$(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_linux_amd64' $(MAIN_PATH)"

.PHONY: build-linux-arm64
build-linux-arm64:
	@$(ECHO) "Building linux/arm64..."
	@$(PS_ENV) "$$env:GOOS='linux'; $$env:GOARCH='arm64'; go build -o '$(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_linux_arm64' $(MAIN_PATH)"

# Darwin platform builds
.PHONY: build-darwin-amd64
build-darwin-amd64:
	@$(ECHO) "Building darwin/amd64..."
	@$(PS_ENV) "$$env:GOOS='darwin'; $$env:GOARCH='amd64'; go build -o '$(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_darwin_amd64' $(MAIN_PATH)"

.PHONY: build-darwin-arm64
build-darwin-arm64:
	@$(ECHO) "Building darwin/arm64..."
	@$(PS_ENV) "$$env:GOOS='darwin'; $$env:GOARCH='arm64'; go build -o '$(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_darwin_arm64' $(MAIN_PATH)"

# Build and package release versions
.PHONY: release
release: build-all
	@$(ECHO) "Packaging release versions..."
	-@$(MKDIR_RELEASE)
ifeq ($(OS),Windows_NT)
	@$(PS_ENV) "Get-ChildItem -Path '$(BUILD_DIR)' -File | ForEach-Object { Compress-Archive -Force -Path $$_.FullName,'README.md','LICENSE' -DestinationPath '$(BUILD_DIR)\release\$$($$_.BaseName).zip' }"
else
	$(foreach PLATFORM,$(PLATFORMS),\
		$(foreach ARCH,$(ARCHITECTURES),\
			$(eval GOOS := $(PLATFORM))\
			$(eval GOARCH := $(ARCH))\
			$(eval SUFFIX := $(if $(findstring windows,$(GOOS)),.exe,))\
			$(eval BINARY := $(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_$(GOOS)_$(GOARCH)$(SUFFIX))\
			$(eval PACKAGE := $(BUILD_DIR)/release/$(PROJECT_NAME)_$(VERSION)_$(GOOS)_$(GOARCH).zip)\
			$(CHECK_FILE) $(BINARY) ]; then \
				echo "Packaging $(GOOS)/$(GOARCH)..." && \
				$(ZIP) $(PACKAGE) $(BINARY) README.md LICENSE || true ; \
			fi ; \
		)\
	)
endif

# Build and package tar.gz archives for all platforms
.PHONY: build-all-tgz
build-all-tgz: build-all
	@$(ECHO) "create tar.gz file..."
ifeq ($(OS),Windows_NT)
	@$(PS_ENV) "Get-ChildItem -Path '$(BUILD_DIR)' -File | ForEach-Object { $$fileName = $$_.FullName; $$outName = $$fileName -replace '\.exe$$','' ; $$tempDir = Join-Path '$(BUILD_DIR)' ('temp_' + [System.IO.Path]::GetRandomFileName()); New-Item -ItemType Directory -Path $$tempDir; Copy-Item -Path $$_.FullName,'LICENSE' -Destination $$tempDir; tar -czf \"$$outName.tar.gz\" -C $$tempDir .; Remove-Item -Recurse -Force $$tempDir }"
else
	$(foreach PLATFORM,$(PLATFORMS),\
		$(foreach ARCH,$(ARCHITECTURES),\
			$(eval GOOS := $(PLATFORM))\
			$(eval GOARCH := $(ARCH))\
			$(eval SUFFIX := $(if $(findstring windows,$(GOOS)),.exe,))\
			$(eval BINARY := $(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_$(GOOS)_$(GOARCH)$(SUFFIX))\
			$(eval PACKAGE := $(BUILD_DIR)/$(PROJECT_NAME)_$(VERSION)_$(GOOS)_$(GOARCH).tar.gz)\
			$(CHECK_FILE) $(BINARY) ]; then \
				echo "Packaging $(GOOS)/$(GOARCH) to tar.gz..." && \
				(mkdir -p $(BUILD_DIR)/temp && \
				cp $(BINARY) README.md LICENSE $(BUILD_DIR)/temp/ && \
				tar -czf $(PACKAGE) -C $(BUILD_DIR)/temp . && \
				rm -rf $(BUILD_DIR)/temp) || true ; \
			fi ; \
		)\
	)
endif

# Run tests
.PHONY: test
test:
	@$(ECHO) "Running tests..."
	@$(GOTEST) -v ./...

# Help information
.PHONY: help
help:
	@$(ECHO) "Available commands:"
	@$(ECHO) "  make              - Clean and build executable for current platform"
	@$(ECHO) "  make build        - Build executable for current platform"
	@$(ECHO) "  make build-all    - Build executables for all platforms"
	@$(ECHO) "  make build-all-tgz - Build executables and package as tar.gz archives"
	@$(ECHO) "  make clean        - Clean build directory"
	@$(ECHO) "  make test         - Run tests"
	@$(ECHO) "  make release      - Build and package all platform releases"
	@$(ECHO) "  make help         - Show help information"