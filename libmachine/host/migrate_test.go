package host

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMigrateHost(t *testing.T) {
	testCases := []struct {
		description                string
		hostBefore                 *Host
		rawData                    []byte
		expectedHostAfter          *Host
		expectedMigrationPerformed bool
		expectedMigrationError     error
	}{
		{
			description: "Config version 4 (from the FUTURE) on disk",
			hostBefore: &Host{
				Name: "default",
			},
			rawData: []byte(`{
    "ConfigVersion": 4,
    "Driver": {"MachineName": "default"},
    "DriverName": "virtualbox",
    "HostOptions": {
        "Driver": "",
        "Memory": 0,
        "Disk": 0,
        "AuthOptions": {
            "StorePath": "/Users/nathanleclaire/.local/machine/machines/default"
        }
    },
    "Name": "default"
}`),
			expectedHostAfter:          nil,
			expectedMigrationPerformed: false,
			expectedMigrationError:     errConfigFromFuture,
		},
	}

	for _, tc := range testCases {
		actualHostAfter, actualMigrationPerformed, actualMigrationError := MigrateHost(tc.hostBefore, tc.rawData)

		assert.Equal(t, tc.expectedHostAfter, actualHostAfter)
		assert.Equal(t, tc.expectedMigrationPerformed, actualMigrationPerformed)
		assert.Equal(t, tc.expectedMigrationError, actualMigrationError)
	}
}
