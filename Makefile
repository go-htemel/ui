.PHONY: build test lint run tailwind-safelist check-safelist tidy

ADDR ?= :8080

build:
	go build ./...

test:
	go test -race ./...

lint:
	@if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed; falling back to go vet"; \
		go vet ./...; \
	fi

# Starts the component showcase on $(ADDR) (default :8080).
run:
	go run ./cmd/showcase -addr $(ADDR)

# Writes every Tailwind class the components use, for the default theme, to tailwind.safelist.txt.
# Pass flags through SAFELIST_FLAGS, e.g. make tailwind-safelist SAFELIST_FLAGS="-primary indigo -radius lg".
tailwind-safelist:
	go run ./internal/cmd/safelist $(SAFELIST_FLAGS) > tailwind.safelist.txt

tidy:
	go mod tidy
