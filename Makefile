.PHONY: install build dev tunnel migrate generate pre-commit-install lint test

init: pre-commit-install
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/golang/protobuf/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

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
	go run cmd/proto2yaml/proto_to_yaml.go -input=./proto -out=./internal/infrastructure/persistence/yaml/proto/services.yaml

lint:
	cd server && golangci-lint run ./...

test:
	cd server && go test ./...

pre-commit-install:
	@echo "Installing pre-commit..."
	@pre-commit install
	@pre-commit install --hook-type commit-msg
