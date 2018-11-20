# Initialize version and gc flags
GO_LDFLAGS := -X `go list ./version`.GitCommit=`git rev-parse --short HEAD 2>/dev/null`
GO_GCFLAGS :=

GO_SRC = $(shell find . -type f -name '*.go')

podman-machine: $(GO_SRC)
	go build -ldflags "$(GO_LDFLAGS)" $(GO_GCFLAGS) ./cmd/podman-machine

cross: podman-machine.linux-amd64 podman-machine.darwin-amd64 podman-machine.windows-amd64

podman-machine.linux-amd64: $(GO_SRC)
	GOOS=linux GOARCH=amd64 go build -o $@ -ldflags "$(GO_LDFLAGS)" $(GO_GCFLAGS) ./cmd/podman-machine

podman-machine.darwin-amd64: $(GO_SRC)
	GOOS=darwin GOARCH=amd64 go build -o $@ -ldflags "$(GO_LDFLAGS)" $(GO_GCFLAGS) ./cmd/podman-machine

podman-machine.windows-amd64: podman-machine.windows-amd64.exe
podman-machine.windows-amd64.exe: $(GO_SRC)
	GOOS=windows GOARCH=amd64 go build -o $@ -ldflags "$(GO_LDFLAGS)" $(GO_GCFLAGS) ./cmd/podman-machine

.PHONY: test
test:
	go test ./...
