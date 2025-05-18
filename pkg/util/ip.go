package util

import (
	"fmt"
	"net"
	"strconv"
)

func ParseIPMask(maskStr string) (net.IPMask, error) {
	ip := net.ParseIP(maskStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP mask string: %q", maskStr)
	}
	ip4 := ip.To4()
	if ip4 == nil {
		return nil, fmt.Errorf("not a valid IPv4 mask: %q", maskStr)
	}
	return net.IPv4Mask(ip4[0], ip4[1], ip4[2], ip4[3]), nil
}

func ParseCIDRMaskFromString(s string) (net.IPMask, error) {
	prefixLen, err := strconv.Atoi(s)
	if err != nil {
		return nil, fmt.Errorf("invalid prefix length: %q", s)
	}
	if prefixLen < 0 || prefixLen > 32 {
		return nil, fmt.Errorf("invalid prefix length: %d (must be between 0 and 32)", prefixLen)
	}
	return net.CIDRMask(prefixLen, 32), nil
}
