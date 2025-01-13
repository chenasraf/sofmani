.PHONY: build
build:
	go build

.PHONY: run
run:
	./sofmani

.PHONY: test
test:
	go test -v ./...
