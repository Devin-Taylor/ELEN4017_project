package main

import (
	// "net"
	"os"
	"fmt"
	// "io/ioutil"
	// "strings"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port  message", os.Args[0])
		os.Exit(1)
	}

	var config configSettings
	config.initializeConfig()

	switch os.Args[1] {
		case "protocol": config.setProtocol(os.Args[2])
		case "connection": config.setConnection(os.Args[2])
		default: 
	}

	err := writeConfig(config)
	checkError(err)



/*	service := os.Args[1]
	message := os.Args[2]

	conn, err := net.Dial("udp", service)
	checkError(err)

	_, err = conn.Write([]byte(message))
	checkError(err)

	var buf[512]byte
	_, err = conn.Read(buf[0:])
	checkError(err)

	fmt.Println(string(buf[0:]))*/

	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

