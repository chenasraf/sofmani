.PHONY: pklgen
pklgen:
	rm -rf appconfig/
	pkl-gen-go pkl/AppConfig.pkl --base-path github.com/chenasraf/sofmani

.PHONY: build
build: pklgen
	go build

run:
	./sofmani
