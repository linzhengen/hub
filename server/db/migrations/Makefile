.PHONY: migrate-create
migrate-create:
	@read -p "Enter migration name: " name; \
	if [ -z "$$name" ]; then \
		echo "Migration name is required"; \
		exit 1; \
	fi; \
	migrate create -seq=true --ext=sql -dir=./mysql $$name
