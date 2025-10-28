GOOS=linux
GOARCH=amd64
PLUGIN_OUT=build/plugins
MIDDLEWARE_OUT=build/middlewares

.PHONY: all build plugins clean lint

all: build plugins

build:
	mkdir -p build
	CGO_ENABLED=1 go build -o build/bravka ./cmd/main.go

plugins:
	mkdir -p $(PLUGIN_OUT)
	CGO_ENABLED=1 go build -buildmode=plugin -o $(PLUGIN_OUT)/logger.so ./builtin/plugins/logger/plugin.go
	CGO_ENABLED=1 go build -buildmode=plugin -o $(MIDDLEWARE_OUT)/recoverer.so ./builtin/middlewares/recoverer/middleware.go

clean:
	rm -rf build

lint:
	golangci-lint run