GOCACHE ?= $(CURDIR)/.cache/go-build
GOMODCACHE ?= $(CURDIR)/.cache/go-mod

export GOCACHE
export GOMODCACHE

run:
	xcopy /E /I .\asset ".\build\asset"
	go build -o build/main ./cmd/server
	./build/main
setup:
	go run ./cmd/setup/

clean:
	