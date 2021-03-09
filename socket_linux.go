package main

import (
	"golang.org/x/sys/unix"
)

// Useful references
// getsockopt: https://man7.org/linux/man-pages/man2/getsockopt.2.html
// ioctls: https://man7.org/linux/man-pages/man4/tty_ioctl.4.html
// tcp_info struct: https://github.com/torvalds/linux/blob/master/include/uapi/linux/tcp.h#L214

// how iperf gets retransmits: https://github.com/esnet/iperf/blob/98d87bd7e82b98775d9e4c62235132caa54233ab/src/tcp_info.c#L118

// Fastly exposes some similar information and documents the fields: https://developer.fastly.com/reference/vcl/variables/backend-connection/
// E.g., https://developer.fastly.com/reference/vcl/variables/backend-connection/backend-socket-tcpi-total-retrans/

// SocketData wraps socket data extracted from syscalls.
type SocketData struct {
	RecvQ       int
	SendQ       int
	RecvBuf     int
	SendBuf     int
	RecvTimeout *unix.Timeval
	SendTimeout *unix.Timeval
	Linger      *unix.Linger
	TCPInfo     *unix.TCPInfo
}

// Control extracts socket information tracked by the kernel into this SocketData.
// This should be passed to Control directly.
func (s *SocketData) Control(fdUintPtr uintptr) {
	fd := int(fdUintPtr)
	s.RecvQ, _ = unix.IoctlGetInt(fd, unix.TIOCINQ)
	s.SendQ, _ = unix.IoctlGetInt(fd, unix.TIOCOUTQ)
	s.RecvBuf, _ = unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF)
	s.SendBuf, _ = unix.GetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_SNDBUF)
	s.RecvTimeout, _ = unix.GetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO)
	s.SendTimeout, _ = unix.GetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_SNDTIMEO)
	s.Linger, _ = unix.GetsockoptLinger(fd, unix.SOL_SOCKET, unix.SO_LINGER)
	s.TCPInfo, _ = unix.GetsockoptTCPInfo(fd, unix.SOL_TCP, unix.TCP_INFO)
}
