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
