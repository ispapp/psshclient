build:
	go build -o bin/
build-prod:
	go build -ldflags "-s -w" -o bin/ -tags prod
run:
	go run .
tests:
	go test ./...
clean:
	rm -rf bin/

build-mac: clean
	mkdir bin
	fyne package -os darwin -icon ./Icon.png --release --tags prod
	mv psshclient.app bin/

build-linux: clean
	mkdir bin
	fyne package -os linux -icon ./Icon.png --release --tags prod
	mv psshclient.tar.xz bin/

build-win: clean
	mkdir bin
	fyne package -os windows -icon ./Icon.png --release --tags prod --appID co.ispps.psshclient
	mv psshclient.exe bin/

# Release management targets
.PHONY: help release-create release-delete release-list release-push release-delete-remote

# Show help information
help:
	@echo "Available targets:"
	@echo ""
	@echo "Build targets:"
	@echo "  build          - Build the application"
	@echo "  build-prod     - Build with production flags"
	@echo "  build-mac      - Build macOS app package"
	@echo "  build-linux    - Build Linux package"
	@echo "  build-win      - Build Windows executable"
	@echo "  run            - Run the application"
	@echo "  tests          - Run tests"
	@echo "  clean          - Clean build artifacts"
	@echo ""
	@echo "Release management:"
	@echo "  release VERSION=v1.0.0           - Create and push a new release"
	@echo "  release-create VERSION=v1.0.0    - Create a local tag"
	@echo "  release-push VERSION=v1.0.0      - Push tag to remote (triggers CI)"
	@echo "  release-delete VERSION=v1.0.0    - Delete local tag"
	@echo "  release-delete-remote VERSION=v1.0.0 - Delete remote tag"
	@echo "  release-cleanup VERSION=v1.0.0   - Delete both local and remote tag"
	@echo "  release-list                     - List all version tags"
	@echo ""
	@echo "Examples:"
	@echo "  make release VERSION=v1.2.3      - Create and release v1.2.3"
	@echo "  make release-list                - Show all existing releases"
	@echo "  make release-cleanup VERSION=v1.2.3 - Remove v1.2.3 completely"

# Create a new version tag locally
# Usage: make release-create VERSION=v1.0.0
release-create:
ifndef VERSION
	$(error VERSION is not set. Usage: make release-create VERSION=v1.0.0)
endif
	@echo "Creating local tag $(VERSION)..."
	git tag $(VERSION)
	@echo "Tag $(VERSION) created locally."
	@echo "To push to remote, run: make release-push VERSION=$(VERSION)"

# Delete a version tag locally
# Usage: make release-delete VERSION=v1.0.0
release-delete:
ifndef VERSION
	$(error VERSION is not set. Usage: make release-delete VERSION=v1.0.0)
endif
	@echo "Deleting local tag $(VERSION)..."
	git tag -d $(VERSION)
	@echo "Local tag $(VERSION) deleted."

# Push a version tag to remote (triggers GitHub Actions)
# Usage: make release-push VERSION=v1.0.0
release-push:
ifndef VERSION
	$(error VERSION is not set. Usage: make release-push VERSION=v1.0.0)
endif
	@echo "Pushing tag $(VERSION) to remote..."
	git push origin $(VERSION)
	@echo "Tag $(VERSION) pushed to remote. GitHub Actions release workflow will start."

# Delete a version tag from remote
# Usage: make release-delete-remote VERSION=v1.0.0
release-delete-remote:
ifndef VERSION
	$(error VERSION is not set. Usage: make release-delete-remote VERSION=v1.0.0)
endif
	@echo "Deleting remote tag $(VERSION)..."
	git push --delete origin $(VERSION)
	@echo "Remote tag $(VERSION) deleted."

# List all version tags
release-list:
	@echo "Local tags:"
	@git tag -l "v*" | sort -V
	@echo ""
	@echo "Remote tags:"
	@git ls-remote --tags origin | grep "refs/tags/v" | sed 's/.*refs\/tags\///' | sort -V

# Complete release workflow: create, push, and trigger GitHub Actions
# Usage: make release VERSION=v1.0.0
release:
ifndef VERSION
	$(error VERSION is not set. Usage: make release VERSION=v1.0.0)
endif
	@echo "Creating and pushing release $(VERSION)..."
	$(MAKE) release-create VERSION=$(VERSION)
	$(MAKE) release-push VERSION=$(VERSION)
	@echo "Release $(VERSION) created and pushed. Check GitHub Actions for build progress."

# Clean up both local and remote tags
# Usage: make release-cleanup VERSION=v1.0.0
release-cleanup:
ifndef VERSION
	$(error VERSION is not set. Usage: make release-cleanup VERSION=v1.0.0)
endif
	@echo "Cleaning up tag $(VERSION) from both local and remote..."
	-$(MAKE) release-delete VERSION=$(VERSION)
	-$(MAKE) release-delete-remote VERSION=$(VERSION)
	@echo "Tag $(VERSION) cleaned up from both local and remote."