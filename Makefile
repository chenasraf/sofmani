.PHONY: build
build:
	go build

.PHONY: run
run: build
	./sofmani

.PHONY: test
test:
	go test -v ./...

.PHONY: install
install: build
	cp sofmani ~/.local/bin

.PHONY: uninstall
uninstall:
	rm -f ~/.local/bin/sofmani

.PHONY: precommit-install
precommit-install:
	@echo "Installing pre-commit hooks..."
	@echo "#!/bin/sh\n\nmake precommit" > .git/hooks/pre-commit
	@chmod +x .git/hooks/pre-commit
	@echo "Pre-commit hooks installed."

.PHONY: precommit
precommit:
	@STAGED_FILES=$$(git diff --cached --name-only --diff-filter=ACM | grep -E '\.go$$'); \
	if [ -z "$$STAGED_FILES" ]; then \
		echo "No staged Go files to check."; \
	else \
		echo "Running pre-commit checks..."; \
		echo "go fmt"; \
		go fmt ./...; \
		echo "go vet"; \
		go vet ./...; \
		echo "golangci-lint"; \
		golangci-lint run ./...; \
		echo "go test"; \
		go test -v ./...; \
	fi
