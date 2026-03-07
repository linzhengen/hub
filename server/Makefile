.PHONY: install build dev tunnel migrate generate pre-commit-install lint test

init:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'mysql' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/golang/protobuf/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

dev:
	docker compose up

dev-seed: gen
	docker compose exec hub ./tmp/cli seed
	docker compose exec hub ./tmp/cli resource-import

migrate:
	migrate -path db/migrations/mysql -database "mysql://root:password#123@tcp(localhost:56836)/hub" up

gen:
	sqlc generate
	buf generate
	go run cmd/openapi223/main.go
	go run ./cmd/proto2yaml/proto_to_yaml.go -input=./proto -out=./internal/infrastructure/persistence/yaml/proto/services.yaml

pre-commit-install:
	@echo "Installing pre-commit..."
	@pre-commit install
	@pre-commit install --hook-type commit-msg

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

test:
	@echo "Running tests..."
	@go test ./...
