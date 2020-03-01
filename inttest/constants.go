// +build !windows

package inttest

import "os"

const (
	binaryName = "main"
)

var (
	stopSignal = os.Interrupt
)
