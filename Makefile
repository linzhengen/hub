.PHONY: install build cli dev tunnel migrate generate pre-commit-install lint test

# Code-generation tools are pinned: with @latest, `make gen` produced different
# output depending on when `make init` last ran, and a one-line proto change
# could rewrite the whole of server/pb.
# Keep these in step with server/go.mod - PROTOC_GEN_GO_VERSION with
# google.golang.org/protobuf, and the gateway plugins with
# github.com/grpc-ecosystem/grpc-gateway/v2 - and re-run `make gen` when you
# move one.
#
# protoc-gen-go comes from google.golang.org/protobuf: github.com/golang/protobuf
# is the retired pre-modules repository.
SQLC_VERSION ?= v1.30.0
MIGRATE_VERSION ?= v4.19.1
GRPC_GATEWAY_VERSION ?= v2.29.0
PROTOC_GEN_GO_VERSION ?= v1.36.11
PROTOC_GEN_GO_GRPC_VERSION ?= v1.6.1

init: pre-commit-install
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@$(SQLC_VERSION)
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@$(MIGRATE_VERSION)
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@$(GRPC_GATEWAY_VERSION)
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@$(GRPC_GATEWAY_VERSION)
	go install google.golang.org/protobuf/cmd/protoc-gen-go@$(PROTOC_GEN_GO_VERSION)
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@$(PROTOC_GEN_GO_GRPC_VERSION)

dev:
	docker compose up -d
	@echo "Waiting for Keycloak to be healthy..."
	@while [ "$$(docker compose ps keycloak --format "{{.Health}}")" != "healthy" ]; do \
		sleep 2; \
	done
	docker compose exec keycloak /opt/keycloak/bin/kcadm.sh config credentials --server http://localhost:8080 --realm master --user admin --password admin
	docker compose exec keycloak /opt/keycloak/bin/kcadm.sh update realms/master -s sslRequired=NONE
	cd ./infra/tf/dev && terraform init && terraform apply -auto-approve
	make dev-seed
	docker compose logs -f

dev-seed: gen
	docker compose exec hub ./tmp/cli seed
	docker compose exec hub ./tmp/cli resource-import

migrate:
	migrate -path server/db/migrations/postgres -database "postgres://postgres:password#123@localhost:56836/hub?sslmode=disable" up

gen:
	cd server && \
	sqlc generate && \
	go run cmd/gen-adapter/main.go && \
	buf generate && \
	go run cmd/openapi223/main.go && \
	go run cmd/proto2yaml/proto_to_yaml.go -input=./proto -out=./internal/infrastructure/persistence/yaml/proto/services.yaml && \
	go run ./cmd/gen-web-client && \
	go run ./cmd/hub api docs --out ../.agents/skills/hub-api/references/api-reference.md

# Installs the `hub` API client on the PATH. It is a plain HTTP client, so it
# needs an endpoint and a token but no database.
cli:
	cd server && go install ./cmd/hub

lint:
	cd server && golangci-lint run ./...

test:
	cd server && go test ./...

pre-commit-install:
	@echo "Installing pre-commit..."
	@pre-commit install
	@pre-commit install --hook-type commit-msg

IMAGE_NAME ?= ghcr.io/$(shell gh api user -q .login)/hub
IMAGE_TAG ?= latest
PLATFORMS ?= linux/amd64,linux/arm64

docker-login:
	@if ! gh auth status --active | grep -q "write:packages"; then \
		echo "Missing 'write:packages' scope. Refreshing..."; \
		gh auth refresh -s write:packages; \
	fi
	gh auth token | docker login ghcr.io -u $(shell gh api user -q .login) --password-stdin

docker-build:
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE_NAME):$(IMAGE_TAG) .

docker-push: docker-login
	docker buildx build --platform $(PLATFORMS) -t $(IMAGE_NAME):$(IMAGE_TAG) --push .

# Bitnami image push commands
POSTGRES_SOURCE_IMAGE ?= bitnami/postgresql:latest
POSTGRES_TARGET_IMAGE ?= ghcr.io/$(shell gh api user -q .login)/postgresql:18.3.0-debian-12-r13

push-postgres: docker-login
	docker pull $(POSTGRES_SOURCE_IMAGE)
	docker tag $(POSTGRES_SOURCE_IMAGE) $(POSTGRES_TARGET_IMAGE)
	docker push $(POSTGRES_TARGET_IMAGE)

push-bitnami-images: push-postgres
