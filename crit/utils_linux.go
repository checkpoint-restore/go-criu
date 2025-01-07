//go:build linux

package crit

import (
	"golang.org/x/sys/unix"
)

var (
	addressFamilyMap = map[uint32]string{
		unix.AF_UNIX:    "UNIX",
		unix.AF_NETLINK: "NETLINK",
		unix.AF_BRIDGE:  "BRIDGE",
		unix.AF_KEY:     "KEY",
		unix.AF_PACKET:  "PACKET",
		unix.AF_INET:    "IPV4",
		unix.AF_INET6:   "IPV6",
	}
	socketTypeMap = map[uint32]string{
		unix.SOCK_STREAM:    "STREAM",
		unix.SOCK_DGRAM:     "DGRAM",
		unix.SOCK_SEQPACKET: "SEQPACKET",
		unix.SOCK_RAW:       "RAW",
		unix.SOCK_RDM:       "RDM",
		unix.SOCK_PACKET:    "PACKET",
	}
	socketProtocolMap = map[uint32]string{
		unix.IPPROTO_ICMP:    "ICMP",
		unix.IPPROTO_ICMPV6:  "ICMPV6",
		unix.IPPROTO_IGMP:    "IGMP",
		unix.IPPROTO_RAW:     "RAW",
		unix.IPPROTO_TCP:     "TCP",
		unix.IPPROTO_UDP:     "UDP",
		unix.IPPROTO_UDPLITE: "UDPLITE",
	}
)
