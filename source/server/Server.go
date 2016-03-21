package main

import (
	"net"
	"os"
	"fmt"
	"strings"
	"io/ioutil"
	"log"
)

const httpVersion = "HTTP/1.1"
const path = "../../objects/"

func main() {
	service := ":1235"

	listener, err := net.Listen("tcp", service)
	//packetConn, err := net.ListenPacket("udp", service)
	checkError(err)

	for {
		conn, err := listener.Accept()
		if err != nil {
			continue
		}
		go  handleClient(conn)
		//go handlePacketConn(packetConn)
	}
}


func handleClient(conn net.Conn) {
	// close the connection after this function executes
	defer conn.Close()

	// load the map describing location changes
	locationMap := loadMovesMap()
	//fmt.Println(locationMap["test/index.html"])

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

		method, url, version, _, _ := decomposeRequest(message)

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
			response.headerLines["location:"] = locationMap[url]
			response.entityBody = "<!DOCTYPE HTML PUBLIC \"-//IETF//DTD HTML 2.0//EN\">\n<html>\n<head>\n<title>301 Moved Permanently</title>\n</head>\n<body>\n<h1>Moved Permanently</h1>\n<p>The document has moved <a href=\"" + url + "\">here</a>.</p>\n</body>\n</html>"
			// set flag
			composeResponse = false
		}

		// check if url is valid 
		exists, _ := fileExists(path + url)
		if !exists && composeResponse {
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

					file, err := os.Open(path + url)
					if err != nil {
						log.Fatal(err)
					}
					defer file.Close()

					b, err := ioutil.ReadAll(file)
					html := string(b)

					response.entityBody = html

					// set flag
					composeResponse = false
				case "HEAD":

				case "PUT":

				case "DELETE":

				case "POST":
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
		//fmt.Println(method)
		//fmt.Println(version)
		_, err2 := conn.Write(response.ToBytes())
		if err2 != nil {
			return
		}
	}
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

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
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

	file, err := os.Open(mapLocation)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	mapping := string(b)

	lines :=  strings.Split(mapping, "\n")
	lines = lines[0:len(lines)-1]
	locationMap := make(map[string]string)

	for _, value := range lines {
		locations := strings.Split(value, "\x20")
		fmt.Println(locations)
		locationMap[locations[0]] = locations[1]
	}

	return locationMap
}

/*func handlePacketConn(conn net.PacketConn) {
	var buf [512]byte
	for {
		n, addr, err := conn.ReadFrom(buf[0:])
		if err != nil {
			return
		}
		fmt.Println(string(buf[0:]))
		_, err2 := conn.WriteTo(buf[0:n], addr)
		if err2 != nil {
			return
		}
	}
}*/
