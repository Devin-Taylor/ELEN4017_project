package main

import (
	"net"
	"os"
	"fmt"
	"strings"
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
	// close the connection after this function executes
	defer conn.Close()

	// get message of at maximum 512 bytes
	var buf [512]byte
	for {
		// read input 
		_, err := conn.Read(buf[0:])
		// if there was an error exit
		if err != nil {
			return
		}
		// convert message to string and decompose it
		message := string(buf[0:])
		temp := strings.Split(message,"\x0d\x0a")
		// get the request line for further processing
		requestLine := temp[0]
		// get the header lines 
		// find out where the header lines end
		var i int
		for i = 1; i < len(temp); i++ {
			if temp[i] == "" {
				break
			}
		}
		//headerLines := temp[1:i]
		//check if there is any content in the body
		var bodyLines []string
		if i  < len(temp) {
			// get the body content
			bodyLines = temp[i:len(temp)]
		}
		body := strings.Join(bodyLines, "\x0d\x0a")

		// split the request line into it's components
		requests := strings.Split(requestLine, "\x20")
		method := requests[0]
		url := requests[1]
		version := requests[2]

		//fmt.Println(string(buf[0:]))
		_, err2 := conn.Write([]byte(method + url + version + body)) //conn.Write(buf[0:n])
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
