package main

import (
	"net"
	"os"
	"fmt"
	"strings"
)

const httpVersion = "HTTP/1.1"

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
				//fmt.Println(string(buf[0:]))
		method, _, version, _, _ := decomposeRequest(message)

		composeResponse := true
		var response ResponseMessage
		response.version = httpVersion

		// make sure that version is compatible with server otherwise send a 505 response
		if version != httpVersion && composeResponse {
			// compose 505
			response.statusCode = "505"
			response.phrase = "HTTP Version Not Supported"
			// set problem flag
			composeResponse = false
		}

		// check if url is valid or if it has been moved

		// check what method was requested
		if composeResponse {
			switch strings.ToUpper(method) {
				case "GET":
					// compose 200
                                        response.statusCode = "200"
					response.phrase = "OK"
					response.entityBody = "<temp>test</temp>"
					// set problem flag 
					composeResponse = false
				default:
					// compose 400
					response.statusCode = "400"
					response.phrase = "Bad Request"
					// set problem flag
					composeResponse = false
			}
		}

		_, err2 := conn.Write(response.ToBytes()) //conn.Write(buf[0:n])
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
