package main

import (
	"net"
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	//"time"
)

const httpVersion = "HTTP/1.1"
const path = "../../objects/"

func main() {
	service := ":1235"

	listener, err := net.Listen("tcp", service)
	checkError(err)
	//packetConn, err := net.ListenPacket("udp", service)
	//checkError(err)

	for {
		// make a new socket for any TCP connection that is accepted
		conn, err := listener.Accept()
		if err != nil {
			continue
		}

		// handle the TCP connection
		go  handleTCPClient(conn)

		// handle any UDP connection
		//handleUDPClient(packetConn)
	}
}

func handleUDPClient(conn net.PacketConn) {
	// get message of at maximum 512 bytes
	var buf [512]byte	
	for {
		// read input and get address of sender
		_, addr, err := conn.ReadFrom(buf[0:])
		// if there was an error exit
		if err != nil {
			return
		}

		// convert message to string
		message := string(buf[0:])

		// compose reponse to message
		response := composeResponse(message)

		// write the response to the socket and send to the correct address
		_, err2 := conn.WriteTo(response.ToBytes(), addr)
		if err2 != nil {
			return
		}
	}
}

func persist(message string) bool {
	// get headers
	_, _, _, headers, _ := decomposeRequest(message)
	fmt.Println(headers["Connection:"])
	switch (headers["Connection:"]) {
		case "keep-alive":			
			return true
		case "close":
			return false
		default:
			return false
	}
}


func handleTCPClient(conn net.Conn) {
	defer conn.Close()

	var buf [512]byte
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
		response := composeResponse(message)
		// write the response to the socket
		_, err2 := conn.Write(response.ToBytes())
		if err2 != nil {
			return
		}
	}

	/*defer conn.Close()	
	//defer fmt.Println("closing connection")
	// get message of at maximum 512 bytes
	var buf [512]byte
	for {
		// read input 
		_, err := conn.Read(buf[0:])
		//fmt.Println(n)
		// if there was an error exit
		if err != nil {
			return
		} else {
			// convert message to string
			message := string(buf[0:])		

			// compose reponse to message
			response := composeResponse(message)

			// write the response to the socket
			_, err2 := conn.Write(response.ToBytes())
			if err2 != nil {
				fmt.Println("error")
			}

			// check if the connection must be closed after this message
			//_, _, _, headers, _ := decomposeRequest(message)
			//fmt.Println(headers)
			//fmt.Println(message)
			if !persist(message){
				// close the connection after this function executes
				defer conn.Close()	
				defer fmt.Println("closing connection")
			} else {
				//conn.SetKeepAlive(true)
				//conn.SetReadDeadline(time.Time)
			}
		}		
	}*/
}

func composeResponse(message string) *ResponseMessage{
		// load the map describing location changes
		locationMap := loadMovesMap()

		// decompose message
		method, url, version, _, body := decomposeRequest(message) // maybe move this out of function

		composeResponse := true
		var response = NewResponseMessage()
		response.version = httpVersion

		// make sure that version is compatible with server otherwise send a 505 response
		if version != httpVersion && composeResponse {
			fmt.Println("505")
			// compose 505
			response.statusCode = "505"
			response.phrase = "HTTP Version Not Supported"
			response.entityBody = "<!DOCTYPE HTML PUBLIC \"-//IETF//DTD HTML 2.0//EN\">\n<html>\n<head>\n<title>505 Version Not Supported</title>\n</head>\n<body>\n<h1>Version Not Supported</h1>\n<p>Your HTTP version is not supported by this server, please use HTTP/1.1.</p>\n</body>\n</html>"
			// set flag
			composeResponse = false
		}

		// check if url has been moved
		if locationMap[url] != "" && composeResponse {
			fmt.Println("301")
			// compose 301
			response.statusCode = "301"
			response.phrase = "Moved Permanently"
			response.headerLines["Location:"] = locationMap[url]
			response.entityBody = "<!DOCTYPE HTML PUBLIC \"-//IETF//DTD HTML 2.0//EN\">\n<html>\n<head>\n<title>301 Moved Permanently</title>\n</head>\n<body>\n<h1>Moved Permanently</h1>\n<p>The document has moved <a href=\"" + url + "\">here</a>.</p>\n</body>\n</html>"
			// set flag
			composeResponse = false
		}

		// check if url is valid 
		exists, _ := fileExists(path + url)
		if !exists && composeResponse && !(strings.ToUpper(method) == "PUT" || strings.ToUpper(method) == "POST") {
			fmt.Println("404")
			// compose 404
			response.statusCode = "404"
			response.phrase = "Not Found"
			response.entityBody = "<!DOCTYPE HTML PUBLIC \"-//IETF//DTD HTML 2.0//EN\">\n<html>\n<head>\n<title>404 Not Found</title>\n</head>\n<body>\n<h1>Not Found</h1>\n<p>The requested URL " + url + " was not found on this server.</p>\n</body>\n</html>"
			// set flag
			composeResponse = false
		}

		// check what method was requested
		if composeResponse {
			switch strings.ToUpper(method) {
				case "GET":
					fmt.Println("200")
					// compose 200
                    response.statusCode = "200"
					response.phrase = "OK"

					// load html file
					file, err := os.Open(path + url)
					if err != nil {
						//need to figure out how to handle this
					}
					defer file.Close()
					// read from file and convert to string
					b, err := ioutil.ReadAll(file)
					html := string(b)

					response.entityBody = html

					// set flag
					composeResponse = false

				case "HEAD":
					fmt.Println("200")
					// compose 200
                    response.statusCode = "200"
					response.phrase = "OK"

					// set flag
					composeResponse = false

				case "PUT":
					fmt.Println("200")
					// compose 200
                    response.statusCode = "200"
					response.phrase = "OK"

					// convert the html to bytes and write to file
					data := []byte(body)
					err := ioutil.WriteFile(path + url, data, 0644)
					checkError(err)

					response.entityBody = "<html>\n<body>\n<h1>The file was created.</h1>\n</body>\n</html>"

					// set flag
					composeResponse = false

				case "DELETE":
					fmt.Println("200")
					// compose 200
					response.statusCode = "200"
					response.phrase = "OK"

					// delete the file
					err := os.RemoveAll(path + url)
					checkError(err)

					response.entityBody = "<html>\n<body>\n<h1>URL deleted.</h1>\n</body>\n</html>"

					//set flag
					composeResponse = false

				case "POST":
					fmt.Println("200")
					// compose 200
                    response.statusCode = "200"
					response.phrase = "OK"

					// write to file
					data := []byte(body)
					err := ioutil.WriteFile(path + url, data, 0644)
					checkError(err)

					response.entityBody = "<html>\n<body>\n<h1>Request Processed Successfully.</h1>\n</body>\n</html>"

					// set flag
					composeResponse = false

				default:
					fmt.Println("400")
					// compose 400
					response.statusCode = "400"
					response.phrase = "Bad Request"
					response.entityBody = "<!DOCTYPE HTML PUBLIC \"-//IETF//DTD HTML 2.0//EN\">\n<html>\n<head>\n<title>400 Bad Request</title>\n</head>\n<body>\n<h1>Bad Request</h1>\n<p>Your browser sent a request that this server could not understand.</p>\n<p>The request line contained invalid characters following the protocol string.</p>\n</body>\n</html>"
					// set flag
					composeResponse = false

			}
		}

		return response
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
			bodyLines = temp[i:len(temp)]
		}
		body := strings.Join(bodyLines, cr + lf)

		// split the request line into it's components
		requests := strings.Split(requestLine, sp)
		method := requests[0]
		url := requests[1]
		version := requests[2]

		return method, url, version, headers, body

}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
	}
}

func fileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}

func loadMovesMap() map[string]string {
	const mapLocation = "../../config/moved_objects.txt"
	locationMap := make(map[string]string)

	file, err := os.Open(mapLocation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return locationMap
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	mapping := string(b)

	lines :=  strings.Split(mapping, "\n")
	lines = lines[0:len(lines)-1]

	for _, value := range lines {
		locations := strings.Split(value, "\x20")

		locationMap[locations[0]] = locations[1]
	}

	return locationMap
}