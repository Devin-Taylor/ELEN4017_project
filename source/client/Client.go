package main

import (
	"net"
	"os"
	"fmt"
	// "io/ioutil"
	// "strings"
)

func main() {

	service := os.Args[1]
	connectionType := os.Args[2]

	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port  message", os.Args[0])
		os.Exit(1)
	}

	var config configSettings
	config.initializeConfig()

	switch service {
		case "protocol": 
			config.protocol = connectionType
			err := writeConfig(config)
			checkError(err)
			os.Exit(0)
		case "connection": 
			config.connection = connectionType
			err := writeConfig(config)
			checkError(err)
			os.Exit(0)
		default:
	}

	conn, err := net.Dial(config.protocol, service)
	checkError(err)

	var request RequestMessage
	request.setRequestLine("GET", "index.html", "HTTP/1.1")
	request.setHeaders(os.Args[1], config.connection, "Mozilla/5.0", "en")

	_, err = conn.Write([]byte(request.toBytes()))
	checkError(err)

	var buf[512]byte
	_, err = conn.Read(buf[0:])
	checkError(err)

	fmt.Println(string(buf[0:]))

	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

