.PHONY: pre-commit-install

pre-commit-install:
	@echo "Installing pre-commit..."
	@pre-commit install
	@pre-commit install --hook-type commit-msg
