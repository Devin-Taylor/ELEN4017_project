package main

import (
	"net"
	"os"
	"fmt"
)

func main() {
	service := ":1235"

	listener, err := net.Listen("tcp", service)
	//packetConn, err := net.ListenPacket("udp", service)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go  handleClient(conn)
		//go handlePacketConn(packetConn)
	}
}

func handlePacketConn(conn net.PacketConn) {
	var buf [512]byte
	for {
		n, addr, err := conn.ReadFrom(buf[0:])
		if err != nil {
			return
		}
		fmt.Println(string(buf[0:]))
		_, err2 := conn.WriteTo(buf[0:n], addr)
		if err2 != nil {
			return
		}
	}
}

func handleClient(conn net.Conn) {
	defer conn.Close()

	var buf [512]byte
	for {
		n, err := conn.Read(buf[0:])
		if err != nil {
			return
		}
		fmt.Println(string(buf[0:]))
		_, err2 := conn.Write(buf[0:n])
		if err2 != nil {
			return
		}
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}
