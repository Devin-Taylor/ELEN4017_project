package main

import (
	"net"
	"strings"
	"fmt"
	"os"
)

func main() {
	service := ":1236"

	listener, err := net.Listen("tcp", service)
	//packetConn, err := net.ListenPacket("udp", service)
	checkError(err)
	// initialize map
	innerMap := make(map[string]string)
	urlMap := make(map[string]map[string]string)
	// initialize channel to allow multiple threads to communicate between each other
	channel := make(chan [3]string)

	for {
	conn, err := listener.Accept()
	if err != nil {
		continue
	}
	go  handleClient(conn, channel)
	// read from channel and assign array to temporary array
	returnArray  := <- channel
	// add mapping values to the map
	innerMap[returnArray[0]] = returnArray[2]
	urlMap[returnArray[1]] = innerMap

	fmt.Println(urlMap)
	}
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func handleClient(conn net.Conn, channel chan [3]string) {

	// defer conn.Close()

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
	
	method, url, _, headerLines, _ := decomposeRequest(message)
	// get the host ID
	host := mapRequest(url, headerLines, channel, conn.RemoteAddr().String(), method)
	// get the response message from the server
	serverResponse := handleServer(message, host)
	// write the response message back to the client
	_, err = conn.Write([]byte(serverResponse))
	checkError(err)
	}
}

func handleServer(relayRequest string, host string) string {
	// initiate connection
	conn, err := net.Dial("tcp", host)
	checkError(err)
	// write request information to the server
	_, err = conn.Write([]byte(relayRequest))
	checkError(err)
	// close the connection after this function executes
	defer conn.Close()
	// get message of at maximum 512 bytes
	var buf [512]byte
	// read input 
	_, err = conn.Read(buf[0:])
	// if there was an error exit
	checkError(err)
	// convert message to string and decompose it
	serverResponse := string(buf[0:])

	return serverResponse
}

func decomposeRequest(request string) (string, string, string, []string, string){
		const sp = "\x20"
		const cr = "\x0d"
		const lf = "\x0a"

		temp := strings.Split(request, cr + lf)
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
		headerLines := temp[1:i]
		//check if there is any content in the body
		var bodyLines []string
		if i  < len(temp) {
			// get the body content
			bodyLines = temp[i:len(temp)]
		}
		body := strings.Join(bodyLines, cr + lf)

		// split the request line into it's components
		requests := strings.Split(requestLine, sp)
		method := requests[0]
		url := requests[1]
		version := requests[2]

		return method, url, version, headerLines, body

}

func mapRequest(url string, headerLines []string, channel chan [3]string, clientAddress string, method string) string {

	var splitString []string
	var tempMap [3]string
	var host string
	// find the hosts address
	for _, value := range headerLines {
		splitString = strings.Split(value, ": ")
		if(strings.ToUpper(splitString[0]) == "HOST"){
			host = splitString[1]
			break
		}
	}
	// create host name and url as a single item
	tempMap[0] = host + url
	// add clients address to the array
	tempMap[1] = clientAddress
	tempMap[2] = method
	// push the map values into the channel
	channel <- tempMap

	return host
}