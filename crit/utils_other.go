//go:build !linux

package crit

var (
	addressFamilyMap = map[uint32]string{
		0x1:  "UNIX",    // syscall.AF_UNIX
		0x10: "NETLINK", // syscall.AF_NETLINK
		0x7:  "BRIDGE",  // syscall.AF_BRIDGE
		0xf:  "KEY",     // syscall.AF_KEY
		0x11: "PACKET",  // syscall.AF_PACKET
		0x2:  "IPV4",    // syscall.AF_INET
		0xa:  "IPV6",    // syscall.AF_INET6
	}
	socketTypeMap = map[uint32]string{
		0x1: "STREAM",    // syscall.SOCK_STREAM
		0x2: "DGRAM",     // syscall.SOCK_DGRAM
		0x5: "SEQPACKET", // syscall.SOCK_SEQPACKET
		0x3: "RAW",       // syscall.SOCK_RAW
		0x4: "RDM",       // syscall.SOCK_RDM
		0xa: "PACKET",    // syscall.SOCK_PACKET
	}
	socketProtocolMap = map[uint32]string{
		0x1:  "ICMP",    // syscall.IPPROTO_ICMP
		0x3a: "ICMPV6",  // syscall.IPPROTO_ICMPV6
		0x2:  "IGMP",    // syscall.IPPROTO_IGMP
		0xff: "RAW",     // syscall.IPPROTO_RAW
		0x6:  "TCP",     // syscall.IPPROTO_TCP
		0x11: "UDP",     // syscall.IPPROTO_UDP
		0x88: "UDPLITE", // syscall.IPPROTO_UDPLITE
	}
)
