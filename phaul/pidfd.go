package phaul

import (
	"fmt"
	"golang.org/x/sys/unix"
)

func pidfdOpen(pid int) (int32, error) {
	r0, _, e1 := unix.Syscall(unix.SYS_PIDFD_OPEN, uintptr(pid), uintptr(0), uintptr(0))
	pidfd := int32(r0)

	if e1 != 0 {
		return -1, e1
	}

	return pidfd, nil
}

// isPidClosed function
// Checks that the process with the given pidfd is not closed.
// When the process that the pidfd refers to terminates, poll
// indicates the file descriptor as readable.
func isPidClosed(pidfd int32) bool {
	pollfds := []unix.PollFd{
		{ Fd: pidfd, Events: unix.POLLIN },
	}

	for {
		ready, err := unix.Poll(pollfds, 0)
		if err == unix.EINTR {
			continue // restart poll syscall
		} else if err != nil {
			fmt.Printf("poll failed: %v (skipping pidfd based pid reuse detection)\n", err)
			return false
		} else if ready == 0 {
			return false // pidfd not readable (process still running)
		} else {
			return true // pidfd readable (process closed)
		}
	}
}
