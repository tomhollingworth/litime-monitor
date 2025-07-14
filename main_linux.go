//go:build linux

package main

import (
	"tinygo.org/x/bluetooth"
)

func Write(char *bluetooth.DeviceCharacteristic, p []byte) (n int, err error) {
	// Write the command to the characteristic to request data
	// This is a blocking operation, so it will wait for the device to respond
	n, err = char.WriteWithoutResponse(p)
	if err != nil {
		return 0, err
	}
	return
}
