package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"syscall"
	"time"
)

// Useful references
//

// how iperf gets retransmits: https://github.com/esnet/iperf/blob/98d87bd7e82b98775d9e4c62235132caa54233ab/src/tcp_info.c#L118

// Fastly exposes some similar information and documents the fields: https://developer.fastly.com/reference/vcl/variables/backend-connection/
// E.g., https://developer.fastly.com/reference/vcl/variables/backend-connection/backend-socket-tcpi-total-retrans/

var (
	flagPort   = flag.Int("port", 9999, "Port to listen on")
	flagReport = flag.Duration("report-interval", time.Second, "Interval to report on")
)

func main() {
	flag.Parse()

	proxyWrap(fmt.Sprintf("127.0.0.1:%v", *flagPort), flag.Args()[0])

}

func proxyWrap(src, dst string) error {
	ln, err := net.Listen("tcp", src)
	if err != nil {
		return err
	}

	for {
		inConn, err := ln.Accept()
		if err != nil {
			log.Printf("Connection accept error: %v", err)
			if err.(net.Error).Temporary() {
				log.Printf("  retrying")
				continue
			}
			return err
		}

		outConn, err := net.Dial("tcp", dst)
		if err != nil {
			log.Printf("Failed to connect to dst %v: %v", dst, err)
		}

		ctx, cancel := context.WithCancel(context.Background())
		go proxy(cancel, inConn, outConn)
		go reportConn(ctx, outConn)
	}
}

func proxy(cancel context.CancelFunc, inConn, outConn net.Conn) {
	defer cancel()

	go io.Copy(inConn, outConn)
	io.Copy(outConn, inConn)
}

func reportConn(ctx context.Context, conn net.Conn) {
	printInfo := func() {
		tcpInfo, err := getTCPInfo(conn)
		if err != nil {
			log.Printf("stop getTCPInfo loop, err %v", err)
			return
		}

		marshalled, _ := json.MarshalIndent(tcpInfo, "", "  ")
		fmt.Println(string(marshalled))
	}

	printInfo()
	for ctx.Err() == nil {
		time.Sleep(*flagReport)
		printInfo()
	}
}

func getTCPInfo(c net.Conn) (*SocketData, error) {
	syscallConn, ok := c.(syscall.Conn)
	if !ok {
		return nil, fmt.Errorf("conn is not syscall.Conn: %T", c)
	}

	sysConn, err := syscallConn.SyscallConn()
	if err != nil {
		return nil, fmt.Errorf("cannot get SyscallConn: %v", err)
	}

	var d SocketData
	if err := sysConn.Control(d.Control); err != nil {
		return nil, fmt.Errorf("Control failed: %v", err)
	}

	return &d, nil
}
