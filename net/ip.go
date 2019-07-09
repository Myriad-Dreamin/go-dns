package mdnet

import (
	"net"
)

const (
	badFormat = "bad format"
)

func ParseUDPDNSIP6(host string) string {
	if _, err := net.ResolveUDPAddr("udp6", host); err == nil {
		return host
	} else {
		return badFormat
	}
}

func ParseTCPDNSIP6(host string) string {
	if _, err := net.ResolveUDPAddr("tcp6", host); err == nil {
		return host
	} else {
		return badFormat
	}
}

func ResolveDNSIP(networkType, host string) (string, string) {
	if networkType == "udp" {
		if _, err := net.ResolveUDPAddr("udp4", host); err == nil {
			return "udp4", host
		} else if _, err := net.ResolveIPAddr("ip4", host); err == nil {
			return "udp4", host + ":53"
		} else if ip := net.ParseIP(host); ip != nil {
			return "udp6", "[" + ip.String() + "]:53"
		} else {
			return "udp6", ParseUDPDNSIP6(host)
		}
	} else if networkType == "tcp" {
		if _, err := net.ResolveUDPAddr("tcp4", host); err == nil {
			return "tcp4", host
		} else if _, err := net.ResolveIPAddr("ip4", host); err == nil {
			return "tcp4", host + ":53"
		} else if ip := net.ParseIP(host); ip != nil {
			return "tcp6", "[" + ip.String() + "]:53"
		} else {
			return "tcp6", ParseTCPDNSIP6(host)
		}
	} else {
		return "bad network", badFormat
	}
}
