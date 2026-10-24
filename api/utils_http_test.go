package main

import (
	"testing"
)

func TestPublicAddressOnly(t *testing.T) {
	forbidden := []string{
		"127.0.0.1:80",
		"10.1.2.3:80",
		"172.18.0.5:5432",
		"192.168.1.20:443",
		"169.254.169.254:80",
		"100.100.1.1:80",
		"0.0.0.0:80",
		"[::1]:80",
		"[fd00::1]:80",
		"[fe80::1]:80",
		"[::ffff:127.0.0.1]:80",
	}
	for _, address := range forbidden {
		if err := publicAddressOnly("tcp", address, nil); err == nil {
			t.Errorf("expected %s to be refused", address)
		}
	}

	allowed := []string{"93.184.216.34:443", "[2606:2800:220:1:248:1893:25c8:1946]:443"}
	for _, address := range allowed {
		if err := publicAddressOnly("tcp", address, nil); err != nil {
			t.Errorf("unexpected error for %s: %v", address, err)
		}
	}
}
