# Project settings
BINARY_NAME=server

# Build the project
build:
	cd ./server && cargo build

# Run the project
run:
	cd ./server && cargo run

# Format code using rustfmt
fmt:
	cd ./server && cargo fmt --all

# Check formatting without applying changes
fmt-check:
	cd ./server && cargo fmt --all -- --check

# Lint code using clippy
lint:
	cd ./server && cargo clippy --all-targets --all-features -- -D warnings

# Run tests
test:
	cargo test

# Clean target directory
clean:
	cargo clean

# Rebuild from scratch
rebuild: clean build

# Help message
help:
	@echo "Usage:"
	@echo "  make build      - Compile the project"
	@echo "  make run        - Run the project"
	@echo "  make fmt        - Format the code"
	@echo "  make fmt-check  - Check formatting"
	@echo "  make lint       - Lint the code with clippy"
	@echo "
