.PHONY: pklgen
pklgen:
	rm -rf appconfig/
	pkl-gen-go pkl/AppConfig.pkl

.PHONY: build
build: pklgen
	go build

run:
	./sofmani
