
# Makefile for Distributed Regional Check Agent

.PHONY: all build clean deb install

# Variables
BINARY_NAME=distributed-regional-check-agent
VERSION=1.0.0
BUILD_DIR=build
PKG_DIR=$(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)
DEB_ARCH ?= amd64
GO_ARCH ?= amd64

# Go build flags
GO_FLAGS=-a -installsuffix cgo -ldflags '-w -s'
CGO_ENABLED=0
GOOS=linux

all: build

build:
	@echo "Building $(BINARY_NAME) for $(GOOS)/$(GO_ARCH)..."
	@mkdir -p $(BUILD_DIR)
	CGO_ENABLED=$(CGO_ENABLED) GOOS=$(GOOS) GOARCH=$(GO_ARCH) go build $(GO_FLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

build-arm64:
	@$(MAKE) build GO_ARCH=arm64 DEB_ARCH=arm64

build-amd64:
	@$(MAKE) build GO_ARCH=amd64 DEB_ARCH=amd64

deb: build
	@echo "Creating .deb package for $(DEB_ARCH)..."
	@mkdir -p $(PKG_DIR)/DEBIAN
	@mkdir -p $(PKG_DIR)/usr/bin
	@mkdir -p $(PKG_DIR)/etc/systemd/system
	@mkdir -p $(PKG_DIR)/etc/$(BINARY_NAME)
	
	# Copy binary with correct name
	cp $(BUILD_DIR)/$(BINARY_NAME) $(PKG_DIR)/usr/bin/$(BINARY_NAME)
	chmod +x $(PKG_DIR)/usr/bin/$(BINARY_NAME)
	
	# Copy systemd service
	cp packaging/regional-check-agent.service $(PKG_DIR)/etc/systemd/system/
	
	# Copy configuration
	cp packaging/regional-check-agent.conf $(PKG_DIR)/etc/$(BINARY_NAME)/
	
	# Copy control file and scripts
	sed 's/Architecture: amd64/Architecture: $(DEB_ARCH)/' packaging/control > $(PKG_DIR)/DEBIAN/control
	cp packaging/postinst $(PKG_DIR)/DEBIAN/
	cp packaging/prerm $(PKG_DIR)/DEBIAN/
	cp packaging/postrm $(PKG_DIR)/DEBIAN/
	
	# Set permissions
	chmod 755 $(PKG_DIR)/DEBIAN/postinst
	chmod 755 $(PKG_DIR)/DEBIAN/prerm
	chmod 755 $(PKG_DIR)/DEBIAN/postrm
	chmod 755 $(PKG_DIR)/usr/bin/$(BINARY_NAME)
	
	# Build package
	dpkg-deb --build $(PKG_DIR) $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_$(DEB_ARCH).deb
	@echo "Package created: $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_$(DEB_ARCH).deb"

deb-arm64:
	@$(MAKE) deb GO_ARCH=arm64 DEB_ARCH=arm64

deb-amd64:
	@$(MAKE) deb GO_ARCH=amd64 DEB_ARCH=amd64

deb-all: deb-amd64 deb-arm64

clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)

install: deb
	@echo "Installing package..."
	sudo dpkg -i $(BUILD_DIR)/$(BINARY_NAME)_$(VERSION)_$(DEB_ARCH).deb

test:
	@echo "Running tests..."
	go test ./...

help:
	@echo "Available targets:"
	@echo "  build       - Build binary for current architecture"
	@echo "  build-amd64 - Build binary for AMD64"
	@echo "  build-arm64 - Build binary for ARM64"
	@echo "  deb         - Create .deb package for current architecture"
	@echo "  deb-amd64   - Create .deb package for AMD64"
	@echo "  deb-arm64   - Create .deb package for ARM64"
	@echo "  deb-all     - Create .deb packages for both architectures"
	@echo "  install     - Install the package"
	@echo "  clean       - Clean build artifacts"
	@echo "  test        - Run tests"
	@echo "  help        - Show this help"