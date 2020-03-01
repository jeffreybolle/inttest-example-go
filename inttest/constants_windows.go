// +build windows

package inttest

import (
	"os"
)

const (
	binaryName = "main.exe"
)

var (
	stopSignal = os.Kill
)
