
PREFIX := /usr/local/bin

build:
	go build ./dyniptoroute53.go

install: build
	sudo mv dyniptoroute53 $(PREFIX)/
