package main

import (
	// "net"
	"os"
	"fmt"
	"io/ioutil"
	"strings"
)

func main() {
/*	if len(os.Args) != 3 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port  message", os.Args[0])
		os.Exit(1)
	}*/
	config_lines := readConfig()

	switch os.Args[1] {
		case "protocol": setProtocol(os.Args[2])
	}

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

func readConfig() []string {
	config, err := ioutil.ReadFile("../config/connection_config.txt")
	checkError(err)

	lines := strings.Split(string(config), "\n")

	return lines
}

func setProtocol(protocol string) {
	
}