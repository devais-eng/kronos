package util

import (
	"github.com/rotisserie/eris"
	"net"
	"os"
)

func MaxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func MinInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func ClampInt(val, min, max int) int {
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}

// CeilDiv performs integer division ceiling the result
func CeilDiv(x, y int) int {
	if x%y == 0 {
		return x / y
	}
	return x/y + 1
}

func Retry(times int, fn func() error) error {
	var err error
	for i := 0; i < times; i++ {
		err = fn()
		if err == nil {
			return nil
		}
	}
	return err
}

// IsInDocker checks if the current process is running inside a Docker container
func IsInDocker() bool {
	if _, err := os.Stat("/.dockerenv"); err == nil {
		return true
	}

	return false
}

// ReadMACAddress reads the MAC address of the first active network interface
func ReadMACAddress() (net.HardwareAddr, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, eris.Wrap(err, "failed to get network interfaces")
	}

	for _, netIf := range interfaces {
		if netIf.Flags&net.FlagUp != 0 && len(netIf.HardwareAddr) > 0 {
			return netIf.HardwareAddr, nil
		}
	}

	return nil, eris.New("no active network interface found")
}
