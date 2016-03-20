package main

import (
	"net"
	"os"
	"fmt"
	// "bufio"
	// "io/ioutil"
	// "strings"
)

func main() {
	// get the arguments passed to the code
	service := os.Args[1]
	// if no arguments replied then error
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port  message", os.Args[0])
		os.Exit(1)
	}
	// initialize config settings variables
	var config configSettings
	config.initializeConfig()
	// determine of the input required config settings to be changed
	if len(os.Args) < 2 {
		connectionType := os.Args[2]
		config = checkInput(config, service, connectionType)
	}
	// get the user to input the method to be used as well as the file/url requested
	method, url := getUserInputs()
	// create connection
	conn, err := net.Dial(config.protocol, service)
	checkError(err)
	// initialize request message struct
	request := NewRequestMessage()
	// set request line information
	request.setRequestLine(method, url, "HTTP/1.1")
	// set header information
	request.setHeaders(service, config.connection, "Mozilla/5.0", "en")
	// write request information to the server
	_, err = conn.Write([]byte(request.toBytes()))
	checkError(err)



// for debug
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

func checkInput(config configSettings, service string, connectionType string) configSettings {
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
	return config
}

func getUserInputs() (string, string) {
	var method string
	var url string

	// reader := bufio.NewReader(os.Stdin)
    fmt.Println("Enter method: ")
    fmt.Scanf("%s", &method)
    fmt.Println("Enter URL: ")
    fmt.Scanf("%s", &url)
    // method,_ := reader.ReadString('\n')
    // url,_ := reader.ReadString('\n')

    return method, url
}