package util

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

const defaultDialTimeout = 500 * time.Millisecond

// IsTCPPortOpen reports whether something is accepting connections on host:port.
func IsTCPPortOpen(host string, port int) bool {
	address := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", address, defaultDialTimeout)
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

// DetectLocalHostIPv4 returns the first non-loopback IPv4 address on a local interface.
func DetectLocalHostIPv4() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			ipNet, ok := addr.(*net.IPNet)
			if !ok || ipNet.IP.IsLoopback() {
				continue
			}
			ip4 := ipNet.IP.To4()
			if ip4 == nil {
				continue
			}
			return ip4.String(), nil
		}
	}
	return "", fmt.Errorf("no non-loopback IPv4 address found on local interfaces")
}
