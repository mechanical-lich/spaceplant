name = "spaceplant"
GIT_TAG := $(shell git describe --tags --always --dirty)
LDFLAGS := -ldflags "-X github.com/mechanical-lich/spaceplant/internal/buildinfo.Version=$(GIT_TAG)"

bin:
	mkdir bin
	cp -r ./assets ./bin/assets
	cp -r ./data ./bin/data

.PHONY: clean
clean:
	rm -rf ./bin

run:
	go run $(LDFLAGS) ./cmd/game/*.go

run-debug:
	GODEBUG=gctrace=1 go run $(LDFLAGS) ./cmd/game/*.go

build: clean bin
	go build $(LDFLAGS) -o ./bin/$(name) ./cmd/game

build-all: clean build-mac build-windows build-linux
	@echo "Builds for all platforms are complete."

build-mac: bin
	GOOS=darwin GOARCH=arm64 go build $(LDFLAGS) -o ./bin/$(name)-mac ./cmd/game

build-linux: bin
	GOOS=linux GOARCH=amd64 go build $(LDFLAGS) -o ./bin/$(name)-linux ./cmd/game

build-windows: bin
	GOOS=windows GOARCH=amd64 go build $(LDFLAGS) -o ./bin/$(name)-windows.exe ./cmd/game

.PHONY: update-go-deps
update-go-deps:
	@echo ">> updating Go dependencies"
	@for m in $$(go list -mod=readonly -m -f '{{ if and (not .Indirect) (not .Main)}}{{.Path}}{{end}}' all); do \
		go get $$m; \
	done
	go mod tidy
ifneq (,$(wildcard vendor))
	go mod vendor
endif
