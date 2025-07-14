//go:build windows

package main

import "tinygo.org/x/bluetooth"

func Write(char *bluetooth.DeviceCharacteristic, p []byte) (n int, err error) {
	return char.Write(p)
}
