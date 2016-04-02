// Author: James Allingham

package main

import (
	"net"
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"time"
	"strconv"
	"lib"
)

const httpVersion = "HTTP/1.1"
const path = "../../objects/"

func main() {
	// set the port on which to listen
	service := ":1235"

	go startTCPServer(service)

	go startUDPServer(service)

	// keep server running
	for {

	}
}

// function responsible for starting and running the TCP server
// inputs - a string containing the port on which to run the server
func startTCPServer(service string) {
	defer fmt.Println("closing TCP server")
	//create a listener for a TCP connection on the given port
	listener, err := net.Listen("tcp", service)
	lib.CheckError(err)

	for {
		// make a new socket for any TCP connection that is accepted
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		// handle the TCP connection on a new thread
		fmt.Println("New connection for ", conn.RemoteAddr())
		go  handleTCPClient(conn)
	}
}

// function responsible for starting and running the UDP server
// inputs - a string containing the port on which to run the server
func startUDPServer(service string) {
	defer fmt.Println("closing UDP server")
	packetConn, err := net.ListenPacket("udp", service)
	lib.CheckError(err)

	for {
		// handle any UDP connection
		handleUDPClient(packetConn)
	}
}

// function responsible for handling a client communicating with the UDP server
// inputs - the net.PacketConn which represents the communication channel between the client and server
func handleUDPClient(conn net.PacketConn) {
	// get message of at maximum 512 bytes
	var buf [1024]byte	
	for {
		// read input and get address of sender
		_, addr, err := conn.ReadFrom(buf[0:])
		// if there was an error exit
		if err != nil {
			return
		}

		// convert message to string
		message := string(buf[0:])

		// compose response to message
		response := composeResponse(message, strings.Split(addr.String(), ":")[0])

		// write the response to the socket and send to the correct address
		_, err2 := conn.WriteTo(response.ToBytes(), addr)
		if err2 != nil {
			return
		}
	}
}

// function responsible for handling a client that makes a connection to the TCP server
// inputs - the net.Conn which represents the connection made by the client
func handleTCPClient(conn net.Conn) {
	defer conn.Close()
	defer fmt.Println("closing connection for ", conn.RemoteAddr())

	var buf [2048]byte
	for {
		// read input 
		_, err := conn.Read(buf[0:])		
		// if there was an error exit
		if err != nil {
			return
		}

		// convert message to string
		message := string(buf[0:])		

		// compose reponse to message
		response := composeResponse(message, strings.Split(conn.LocalAddr().String(), ":")[0])
		// write the response to the socket
		_, err2 := conn.Write(response.ToBytes())
		if err2 != nil {
			return
		}
	}
}

// function responsible for composing an appropriate response for the client
// inputs - string containing the request send by the client
// outputs - the response message as a ResponseMessage struct
func composeResponse(message string, address string) *lib.ResponseMessage{
		// load the map describing location changes
		locationMap := loadMovesMap()

		// decompose message
		method, url, version, headers, body := lib.DecomposeRequest(message)

		// set the flag indicating that a response must still be composed to true
		composeResponse := true

		// create a new response
		var response = lib.NewResponseMessage()
		// set the HTTP version
		response.Version = httpVersion

		// set response headers		
		response.HeaderLines["Server"] = "FooBar"
		response.HeaderLines["Date"] = time.Now().Format(time.RFC1123Z) // this is the time format recommended by golang for a server application
		response.HeaderLines["Content-Language"] = "en"

		// make sure that version is compatible with server otherwise send a 505 response
		if version != httpVersion && composeResponse {
			fmt.Println("505")
			// compose 505
			response.StatusCode = "505"
			response.Phrase = "HTTP Version Not Supported"
			response.EntityBody = "<!DOCTYPE HTML PUBLIC \"-//IETF//DTD HTML 2.0//EN\">\n<html>\n<head>\n<title>505 Version Not Supported</title>\n</head>\n<body>\n<h1>Version Not Supported</h1>\n<p>Your HTTP version is not supported by this server, please use HTTP/1.1.</p>\n</body>\n</html>"
			response.HeaderLines["Content-Length"] = strconv.Itoa(len([]byte(response.EntityBody)))
			// set flag to false indicating that a response has been composed
			composeResponse = false
		}

		// check if url has been moved
		if locationMap[url] != "" && composeResponse {
			fmt.Println("301")
			// compose 301
			response.StatusCode = "301"
			response.Phrase = "Moved Permanently"
			response.HeaderLines["Location"] = "http://" + address + locationMap[url]
			response.EntityBody = "<!DOCTYPE HTML PUBLIC \"-//IETF//DTD HTML 2.0//EN\">\n<html>\n<head>\n<title>301 Moved Permanently</title>\n</head>\n<body>\n<h1>Moved Permanently</h1>\n<p>The document has moved <a href=\"" + url + "\">here</a>.</p>\n</body>\n</html>"
			response.HeaderLines["Content-Length"] = strconv.Itoa(len([]byte(response.EntityBody)))
			// set flag
			composeResponse = false
		}

		// check if url is valid 
		exists, _ := lib.FileExists(path + url)
		if !exists && composeResponse && !(strings.ToUpper(method) == "PUT" || strings.ToUpper(method) == "POST") {
			fmt.Println("404")
			// compose 404
			response.StatusCode = "404"
			response.Phrase = "Not Found"
			response.EntityBody = "<!DOCTYPE HTML PUBLIC \"-//IETF//DTD HTML 2.0//EN\">\n<html>\n<head>\n<title>404 Not Found</title>\n</head>\n<body>\n<h1>Not Found</h1>\n<p>The requested URL " + url + " was not found on this server.</p>\n</body>\n</html>"
			response.HeaderLines["Content-Length"] = strconv.Itoa(len([]byte(response.EntityBody)))
			// set flag
			composeResponse = false
		}

		// check what method was requested
		if composeResponse {
			switch strings.ToUpper(method) {
				case "GET":
					// get last modified time
     				stat, err := os.Stat(path + url)
     				if err != nil {
        				fmt.Println(err)
     				}
     				serverTime, _ := time.Parse(time.RFC1123Z, stat.ModTime().Format(time.RFC1123Z))
     				proxyTime, _ := time.Parse(time.RFC1123Z, headers["If-Modified-Since"])

     				// check if modified time is after a last modified time
     				if headers["If-Modified-Since"] == "" || serverTime.After(proxyTime){
						fmt.Println("200")
						// compose 200
                    	response.StatusCode = "200"
						response.Phrase = "OK"

						// load html file
						file, err := os.Open(path + url)
						if err != nil {
							//need to figure out how to handle this
						}
						defer file.Close()
						// read from file and convert to string
						b, err := ioutil.ReadAll(file)
						html := string(b)

						response.EntityBody = html	
						response.HeaderLines["Content-Length"] = strconv.Itoa(len([]byte(response.EntityBody)))			
     				
     					// add last modified header
						response.HeaderLines["Last-Modified"] = serverTime.Format(time.RFC1123Z)
     				} else {
     					fmt.Println("304")
						// compose 304
                    	response.StatusCode = "304"
						response.Phrase = "Not Modified"

						response.EntityBody = ""
						response.HeaderLines["Content-Length"] = strconv.Itoa(len([]byte(response.EntityBody)))
     				}					

					// set flag
					composeResponse = false

				case "HEAD":
					fmt.Println("200")
					// compose 200
                    response.StatusCode = "200"
					response.Phrase = "OK"

					// get last modified time
     				stat, err := os.Stat(path + url)
     				if err != nil {
        				fmt.Println(err)
     				}
     				response.HeaderLines["Content-Length"] = strconv.FormatInt(stat.Size(),10)
     				response.EntityBody = ""

					// set flag
					composeResponse = false

				case "PUT":
					fmt.Println("200")
					// compose 200
                    response.StatusCode = "200"
					response.Phrase = "OK"

					// convert the html to bytes and write to file
					data := []byte(body)
					err := ioutil.WriteFile(path + url, data, 0644)
					lib.CheckError(err)

					response.EntityBody = "<html>\n<body>\n<h1>The file was created.</h1>\n</body>\n</html>"
					response.HeaderLines["Content-Length"] = strconv.Itoa(len([]byte(response.EntityBody)))

					// set flag
					composeResponse = false

				case "DELETE":
					fmt.Println("200")
					// compose 200
					response.StatusCode = "200"
					response.Phrase = "OK"

					// delete the file
					err := os.RemoveAll(path + url)
					lib.CheckError(err)

					response.EntityBody = "<html>\n<body>\n<h1>URL deleted.</h1>\n</body>\n</html>"
					response.HeaderLines["Content-Length"] = strconv.Itoa(len([]byte(response.EntityBody)))

					//set flag
					composeResponse = false

				case "POST":
					fmt.Println("200")
					// compose 200
                    response.StatusCode = "200"
					response.Phrase = "OK"

					// write to file
					data := []byte(body)
					err := ioutil.WriteFile(path + url, data, 0644)
					lib.CheckError(err)

					response.EntityBody = "<html>\n<body>\n<h1>Request Processed Successfully.</h1>\n</body>\n</html>"
					response.HeaderLines["Content-Length"] = strconv.Itoa(len([]byte(response.EntityBody)))

					// set flag
					composeResponse = false

				default:
					fmt.Println("400")
					// compose 400
					response.StatusCode = "400"
					response.Phrase = "Bad Request"
					response.EntityBody = "<!DOCTYPE HTML PUBLIC \"-//IETF//DTD HTML 2.0//EN\">\n<html>\n<head>\n<title>400 Bad Request</title>\n</head>\n<body>\n<h1>Bad Request</h1>\n<p>Your browser sent a request that this server could not understand.</p>\n<p>The request line contained invalid characters following the protocol string.</p>\n</body>\n</html>"
					response.HeaderLines["Content-Length"] = strconv.Itoa(len([]byte(response.EntityBody)))
					// set flag
					composeResponse = false

			}
		}

		return response
}

// function responsible for load the map which indicates if a resource has been moved from one URL to another
func loadMovesMap() map[string]string {
	const mapLocation = "../../config/moved_objects.txt"
	locationMap := make(map[string]string)

	// open the file containing the data
	file, err := os.Open(mapLocation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return locationMap
	}
	defer file.Close()

	// read the contents of the file and convert to string
	b, err := ioutil.ReadAll(file)
	mapping := string(b)

	// split up the contents by line
	lines :=  strings.Split(mapping, "\n")
	lines = lines[0:len(lines)-1]

	// split up each line by a space and store values in map
	for _, value := range lines {
		locations := strings.Split(value, "\x20")

		locationMap[locations[0]] = locations[1]
	}

	return locationMap
}