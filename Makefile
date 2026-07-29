.PHONY: help test test-cover lint fmt security security-install vet clean

help:
	@echo "confkit development targets:"
	@echo ""
	@echo "  make test            - Run all tests"
	@echo "  make test-cover      - Run tests with coverage report"
	@echo "  make lint            - Run golangci-lint"
	@echo "  make vet             - Run go vet"
	@echo "  make fmt             - Check and apply code formatting"
	@echo "  make security        - Run security checks (vet, gosec, govulncheck)"
	@echo "  make security-install - Install security tools (gosec, govulncheck)"
	@echo "  make clean           - Clean build artifacts and coverage files"

test:
	go test -v -race ./...

test-cover:
	go test -v -race -coverprofile=coverage.txt ./...
	go tool cover -func=coverage.txt
	@coverage=$$(go tool cover -func=coverage.txt | grep total | awk '{print $$3}' | sed 's/%//'); \
	echo ""; \
	echo "Coverage: $$coverage%"; \
	if [ "$$(echo "$$coverage < 75" | bc -l)" -eq 1 ]; then \
		echo "ERROR: Coverage is below 75%"; \
		exit 1; \
	fi

lint:
	golangci-lint run ./... --timeout=5m

vet:
	go vet ./...

fmt:
	@echo "Checking formatting..."
	@if go fmt ./... | grep -q .; then \
		echo "Code needs formatting"; \
		exit 1; \
	else \
		echo "Code is properly formatted"; \
	fi

security-install:
	@echo "Installing security tools..."
	go install github.com/securego/gosec/v2/cmd/gosec@latest
	go install golang.org/x/vuln/cmd/govulncheck@latest

security: vet
	@echo "Running security checks..."
	@echo ""
	@echo "=== Running gosec ==="
	gosec -fmt json -out gosec-results.json ./... || true
	@echo ""
	@echo "=== Running govulncheck ==="
	govulncheck ./...

clean:
	rm -f coverage.txt gosec-results.json
	go clean -testcache
