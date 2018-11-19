# Initialize version and gc flags
GO_LDFLAGS := -X `go list ./version`.GitCommit=`git rev-parse --short HEAD 2>/dev/null`
GO_GCFLAGS :=

podman-machine: $(shell find . -type f -name '*.go')
	go build -ldflags "$(GO_LDFLAGS)" $(GO_GCFLAGS) ./cmd/podman-machine

.PHONY: test
test:
	go test ./...
