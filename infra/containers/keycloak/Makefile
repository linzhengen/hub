IMAGE_NAME ?= ghcr.io/$(shell gh api user -q .login)/keyclaok
IMAGE_TAG ?= 26.6.1-debian-12
PLATFORMS ?= linux/amd64,linux/arm64

docker-login:
	@if ! gh auth status --active | grep -q "write:packages"; then \
		echo "Missing 'write:packages' scope. Refreshing..."; \
		gh auth refresh -s write:packages; \
	fi
	gh auth token | docker login ghcr.io -u $(shell gh api user -q .login) --password-stdin

docker-build:
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE_NAME):$(IMAGE_TAG) ./26/debian-12

docker-push: docker-login
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE_NAME):$(IMAGE_TAG) --push ./26/debian-12
