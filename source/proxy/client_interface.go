package main

import (
	"net"
	"strings"
	"fmt"
	"os"
	"io/ioutil"
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

	// fmt.Println(urlMap)
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
	
	method, url, version, headers, body := decomposeRequest(message)
	// get the host ID
	host := mapRequest(url, headers, channel, conn.RemoteAddr().String(), method)

	isInCache, lastModified, locationMap := checkInCache(url)

	if isInCache {
		headers = modifyHeaders(lastModified, headers)
		message = compileNewRequest(method, url, version, headers, body)
	}

	// strings.Split(host, ":")[0]

	// get the response message from the server
	serverResponse := handleServer(message, host)

	isUpdated, newResponse, newTime := getNewResponse(serverResponse, strings.Split(host, ":")[0], url)

	if isUpdated {
		destination := strings.Split(host, ":")[0]+url
		locationMap[destination] = newTime
		saveMap(locationMap, "../../cache/cache_map.txt")
	}
	// write the response message back to the client
	_, err = conn.Write(newResponse.ToBytes())
	checkError(err)
	}
}

func getNewResponse(serverResponse string, host string, url string) (bool, *ResponseMessage, string) {
	version, code, status, headers, body := decomposeResponse(serverResponse)

	if code == "304" {
		file, _ := os.Open("../../cache/"+host+url)
		defer file.Close()
		var response = NewResponseMessage()
		response.version = version
		response.headerLines = headers
		// compose 200
	    response.statusCode = "200"
		response.phrase = "OK"
		// read from file and convert to string
		b, _ := ioutil.ReadAll(file)
		html := string(b)
		response.entityBody = html

		return false, response, ""
	} 

	if code == "200" {

		os.Mkdir("../../cache/"+host, 0644)

		err := ioutil.WriteFile("../../cache/"+host+url, []byte(body), 0644)
		fmt.Println(err)

		var response = NewResponseMessage()
		response.version = version
		response.headerLines = headers
		response.statusCode = "200"
		response.phrase = "OK"
		response.entityBody = body
		newTime := headers["Last-Modified"]

		return true, response, newTime
	}

	var response = NewResponseMessage()
	response.version = version
	response.headerLines = headers
	response.statusCode = code
	response.phrase = status
	response.entityBody = body

	return false, response, ""
}

func checkInCache(url string) (bool, string, map[string]string) {
	locationMap := loadMap("../../cache/cache_map.txt")

	lastModified := locationMap[url]

	if lastModified != "" {
		return true, lastModified, locationMap
	}
	return false, "", locationMap
}

func modifyHeaders(lastModified string, headers map[string]string) map[string]string {
	headers["If-Modified-Since"] = lastModified

	return headers
}

func compileNewRequest(method string, url string, version string, headers map[string]string, body string) string {
	const sp = "\x20"
	const lf = "\x0a"
	const cr = "\x0d"
	requestString := method + sp
	requestString += url + sp
	requestString += version + cr + lf
	//add header lines
	for headerFieldName, value := range headers {
		requestString += headerFieldName + ":" + sp
		requestString += value + cr + lf
	}
	requestString += cr + lf
	requestString += body
	return requestString
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
	var buf [8192]byte
	// read input 
	_, err = conn.Read(buf[0:])
	// if there was an error exit
	checkError(err)
	// convert message to string and decompose it
	serverResponse := string(buf[0:])

	return serverResponse
}

func decomposeRequest(request string) (string, string, string, map[string]string, string){
		const sp = "\x20"
		const cr = "\x0d"
		const lf = "\x0a"
		headers := make(map[string]string)

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
		for _, value := range headerLines {
			//fmt.Println(value)
			line := strings.Split(value, " ")
			//fmt.Println("0: " + line[0] + " 1: " + line[1])
			headers[line[0]] = line[1]
		}
		//check if there is any content in the body
		var bodyLines []string
		if i  < len(temp) {
			// get the body content
			bodyLines = temp[i+1:len(temp)]
		}
		body := strings.Join(bodyLines, cr + lf)

		// split the request line into it's components
		requests := strings.Split(requestLine, sp)
		method := requests[0]
		url := requests[1]
		version := requests[2]

		return method, url, version, headers, body

}

func mapRequest(url string, headers map[string]string, channel chan [3]string, clientAddress string, method string) string {

	var tempMap [3]string
	var host string
	// find the hosts address
	for key, value := range headers {
		if(strings.ToUpper(key) == "HOST:"){
			host = value
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

func decomposeResponse(response string) (string, string, string, map[string]string, string){
		const sp = "\x20"
		const cr = "\x0d"
		const lf = "\x0a"
		headers := make(map[string]string)

		temp := strings.Split(response, cr + lf)
		// get the request line for further processing
		responseLine := temp[0]
		// get the header lines 
		// find out where the header lines end
		var i int
		for i = 1; i < len(temp); i++ {
			if temp[i] == "" {
				break
			}
		}
		headerLines := temp[1:i]
		for _, value := range headerLines {
			line := strings.Split(value, sp)
			headers[line[0]] = line[1]
		}
		//check if there is any content in the body
		var bodyLines []string
		if i  < len(temp) {
			// get the body content
			bodyLines = temp[i+1:len(temp)]
		}
		body := strings.Join(bodyLines, cr + lf)

		// split the response line into it's components
		responses := strings.Split(responseLine, sp)
		status := responses[2]
		code := responses[1]
		version := responses[0]

		return version, code, status, headers, body

}