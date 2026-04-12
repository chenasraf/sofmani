BIN := $(notdir $(CURDIR))

all:
	@if [ ! -f ".git/hooks/pre-commit" ]; then \
		$(MAKE) install-hooks; \
	fi
	$(MAKE) build
	$(MAKE) run

.PHONY: build
build:
	go build -o $(BIN)

.PHONY: run
run: build
	./$(BIN)

.PHONY: test
test:
	go test -v ./...

.PHONY: install
install: build
	cp $(BIN) ~/.local/bin/

.PHONY: uninstall
uninstall:
	rm -f ~/.local/bin/$(BIN)

.PHONY: lint
lint:
	golangci-lint run ./...

.PHONY: install-hooks
install-hooks:
	lefthook install
