package main

import (
	"net"
)

func main() {
	service := ":1236"

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
	
	method, url, version, _, body := decomposeRequest(message)

	
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

func mapRequest()