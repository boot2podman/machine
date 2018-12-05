package main

import (
	"os"
	"testing"

	"github.com/boot2podman/machine/commands/mcndirs"
)

func TestStorePathSetCorrectly(t *testing.T) {
	mcndirs.BaseDir = ""
	os.Args = []string{"podman-machine", "--storage-path", "/tmp/foo"}
	main()
	/*
		if mcndirs.BaseDir != "/tmp/foo" {
			t.Fatal("Expected MACHINE_STORAGE_PATH environment variable to be /tmp/foo but was ", os.Getenv("MACHINE_STORAGE_PATH"))
		}
	*/
}
