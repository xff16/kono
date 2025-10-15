GOOS=linux
GOARCH=amd64
PLUGIN_OUT=build/plugins

.PHONY: all build plugins clean

all: build plugins

build:
	mkdir -p build
	CGO_ENABLED=1 go build -o build/kairyu ./cmd/main.go

plugins:
	mkdir -p $(PLUGIN_OUT)
	CGO_ENABLED=1 go build -buildmode=plugin -o $(PLUGIN_OUT)/logger.so ./plugins/builtin/logger/plugin.go

clean:
	rm -rf build