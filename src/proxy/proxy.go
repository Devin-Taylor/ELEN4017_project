// Author: Devin Taylor and James Allingham

package main

import (
	"net"
	"strings"
	"fmt"
	"os"
	"io/ioutil"
	"lib"
	"strconv"
)

func main() {
	service := ":1236"

	listener, err := net.Listen("tcp", service)
	lib.CheckError(err)
	// listen for connection from client
	for {
	conn, err := listener.Accept()
	if err != nil {
		continue
	}
	// start new thread to handle client connection
	go  handleClient(conn)
	}
}
// function responsible for handling the proxy connection with the client
// inputs - conn: net connection established with client
func handleClient(conn net.Conn) {

	defer conn.Close()

	// get request message
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
		fmt.Println("*******CLIENT*******")
		fmt.Println(message)
		method, url, version, headers, body := lib.DecomposeRequest(message)
		var host string
		// find the hosts address
		for key, value := range headers {
			if(strings.ToUpper(key) == "HOST"){
				host = value
				break
			}
		}
		if strings.ToUpper(method) == "GET" {
			// search the proxy cache for requested files
			isInCache, lastModified, locationMap := checkInCache(url, strings.Split(host, ":")[0])
			// if is in cache then modify the request message to include the last modified date and recompile message
			if isInCache {
				headers = modifyHeaders(lastModified, headers)
				message = compileNewRequest(method, url, version, headers, body)
			}
			// get the response message from the server
			serverResponse := handleServer(message, host)
			// check the server for the file and if it has been modified
			fmt.Println("*******SERVER*******")
			version, code, status, headers, _ := lib.DecomposeResponse(serverResponse)
			var allHeaders string

			for key, value := range headers {
				allHeaders = allHeaders + key + " " + value + "\n"
			}

			content := version + " " + code + " " + status + "\n" + allHeaders + "\n"
			fmt.Println(content)
			isUpdated, newResponse, newTime := getNewResponse(serverResponse, strings.Split(host, ":")[0], url)
			// if file has been modified then write new file to cache
			if isUpdated {
				destination := strings.Split(host, ":")[0]+url
				locationMap[destination] = newTime
				saveMap(locationMap, "../../cache/cache_map.txt")
			}
			// write the response message back to the client
			_, err = conn.Write(newResponse.ToBytes())
			lib.CheckError(err)
		} else {
			// get the response message from the server
			serverResponse := handleServer(message, host)
			_, err = conn.Write([]byte(serverResponse))
			lib.CheckError(err)
		}
	}
}
// function responsible for getting the new response message based on what was received
func getNewResponse(serverResponse string, host string, url string) (bool, *lib.ResponseMessage, string) {
	version, code, status, headers, body := lib.DecomposeResponse(serverResponse)

	// if 304 the has not been modified - so find file in cache
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

	StopIndex := strings.LastIndex(url, "/")
	newUrl := url[StopIndex:len(url)]

	host = host + url[0:StopIndex]
	// if file is new/updated then write the new file to cache and componse new response message 
	if code == "200" {
		// check if folder already exists otherwise make new on
		exists, _ := lib.FileExists("../../cache/"+host)
		if !exists {
			os.MkdirAll("../../cache/"+host, 0777)
		}

		ioutil.WriteFile("../../cache/"+host+newUrl, []byte(body), 0777)

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
// function responsible for checking the proxy cache for a specific file
// inputs - url: string for file url
//			host: string for file host
// outputs - isInCache: bool determining if the file is in the cache
//			 date: string representing last modified date for a file in cache
//			 locationMap: map of cache in file
func checkInCache(url string, host string) (bool, string, map[string]string) {
	locationMap := loadMap("../../cache/cache_map.txt")

	lastModified := locationMap[host+url]
	if lastModified != "" {
		return true, lastModified, locationMap
	}
	return false, "", locationMap
}
// function responsible modifying the message headers if occurs in cache
// inputs - lastModified: string of date file last modified
//			headers: map of current headers
// outputs - headers: new map of headers including a modified date
func modifyHeaders(lastModified string, headers map[string]string) map[string]string {
	headers["If-Modified-Since"] = lastModified

	return headers
}
// function responsible for compiling a new request message
// inputs - method: string for HTTP method
//			url: string for URL
//			version: HTTP version
// 			headers: map of all headers
//			body: string of message body
// outputs - requestMessage: string of new request message
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
// function responsible for handling the connection with the server
// inputs - relayRequest: string of request message that must be relayed from the client to the server
// 			host: server destination
// outputs - response: string for response message from server
func handleServer(relayRequest string, host string) string {
	// initiate connection
	conn, err := net.Dial("tcp", host)
	lib.CheckError(err)
	// write request information to the server
	_, err = conn.Write([]byte(relayRequest))
	lib.CheckError(err)
	// close the connection after this function executes
	defer conn.Close()

	var buf [4096]byte
	n, err := conn.Read(buf[0:])
	lib.CheckError(err)

	response := string(buf[0:n])
	version, code, status, headers, _ := lib.DecomposeResponse(response)

	// get the header size to determine how much more of the file needs to read
	headerSize := getHeaderSize(version, code, status, headers)
	lengthDiff := 0

	contentLen, err := strconv.Atoi(headers["Content-Length"])
	if err == nil {
		lengthDiff = contentLen + headerSize - n
	} else {
		lengthDiff = -1
	}
	// loop of receiving large files until all of file is received
	if strings.ToUpper(headers["Transfer-Encoding"]) == "CHUNKED" {

		for {
			// get message
			var buf [4096]byte
			// read input 
			n, err = conn.Read(buf[0:])
			lib.CheckError(err)
			response += string(buf[0:n])
			if strings.Contains(response, "\r\n0\r\n\r\n") || n == 0 {
					break
			}
		}
	} else {
		for lengthDiff > 0 {
			var buf [4096]byte
			// read input 
			n, err = conn.Read(buf[0:])
			lib.CheckError(err)
			response += string(buf[0:n])
			lengthDiff -= n
		}
		
	}
	return response
}

// function responsible for obtaining the size of the headers that are returned from the server
// outputs - integer representing the size of the headers in bytes
func getHeaderSize(version string, code string, status string, headers map[string]string) int {
	// create a new response message with all the same properties and then get the byte size of that 
	headerTemp := lib.NewResponseMessage()
	headerTemp.Version = version
	headerTemp.StatusCode = code
	headerTemp.Phrase = status
	headerTemp.HeaderLines = headers
	headerTemp.EntityBody = ""
	headerSize := len(headerTemp.ToBytes())

	return headerSize
}
