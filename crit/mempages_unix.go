//go:build !windows

package crit

import "golang.org/x/sys/unix"

const mapShared = unix.MAP_SHARED
