VERSION := $(shell git describe --tags --always --dirty)
LDFLAGS := -X main.version=$(VERSION)

build:
	go build -ldflags "$(LDFLAGS)" -o tswitch

install: build
	sudo ln -sf $(CURDIR)/tswitch /usr/local/bin/tswitch

.PHONY: build install
