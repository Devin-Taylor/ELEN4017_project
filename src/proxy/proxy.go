package main

import (
	"net"
	"strings"
	"fmt"
	"os"
	"io/ioutil"
	"lib"
)

func main() {
	service := ":1236"

	listener, err := net.Listen("tcp", service)
	//packetConn, err := net.ListenPacket("udp", service)
	lib.CheckError(err)
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
	}
}

func handleClient(conn net.Conn, channel chan [3]string) {

	// defer conn.Close()

	// get message of at maximum 512 bytes
	var buf [1024]byte
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

	isInCache, lastModified, locationMap := checkInCache(url, strings.Split(host, ":")[0])

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
	lib.CheckError(err)
	}
}

func getNewResponse(serverResponse string, host string, url string) (bool, *lib.ResponseMessage, string) {
	version, code, status, headers, body := decomposeResponse(serverResponse)

	fmt.Println(code)

	if code == "304" {
		file, _ := os.Open("../../cache/"+host+url)
		defer file.Close()
		var response = lib.NewResponseMessage()
		response.Version = version
		response.HeaderLines = headers
		// compose 200
	    response.StatusCode = "200"
		response.Phrase = "OK"
		// read from file and convert to string
		b, _ := ioutil.ReadAll(file)
		html := string(b)
		response.EntityBody = html

		return false, response, ""
	} 

	if code == "200" {

		exists, _ := fileExists("../../cache/"+host)
		if !exists {
			os.Mkdir("../../cache/"+host, 0777)
		}

		ioutil.WriteFile("../../cache/"+host+url, []byte(body), 0777)

		var response = lib.NewResponseMessage()
		response.Version = version
		response.HeaderLines = headers
		response.StatusCode = "200"
		response.Phrase = "OK"
		response.EntityBody = body
		newTime := headers["Last-Modified"]
		return true, response, newTime
	}

	var response = lib.NewResponseMessage()
	response.Version = version
	response.HeaderLines = headers
	response.StatusCode = code
	response.Phrase = status
	response.EntityBody = body

	return false, response, ""
}

func checkInCache(url string, host string) (bool, string, map[string]string) {
	locationMap := loadMap("../../cache/cache_map.txt")

	lastModified := locationMap[host+url]
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
	fmt.Println(host)
	conn, err := net.Dial("tcp", host)
	lib.CheckError(err)
	// write request information to the server
	_, err = conn.Write([]byte(relayRequest))
	lib.CheckError(err)
	// close the connection after this function executes
	defer conn.Close()
	// get message of at maximum 512 bytes
	var buf [8192]byte
	// read input 
	_, err = conn.Read(buf[0:])
	// if there was an error exit
	lib.CheckError(err)
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
			
			line := strings.SplitN(value, ":"+" ", 2)
			
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
		if(strings.ToUpper(key) == "HOST"){
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
			line := strings.SplitN(value, ":"+sp, 2)
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

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}