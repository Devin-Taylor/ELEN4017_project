package main

import (
	"net"
	"os"
	"os/exec"
	"fmt"
	// "bufio"
	"io/ioutil"
	"strings"
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
	// set request version as need to when launch 505 error later on
	requestVersion := "HTTP/1.1"
	// set request line information
	request.setRequestLine(method, url, requestVersion)
	// set header information
	request.setHeaders(service, config.connection, "Mozilla/5.0", "en")
	// write request information to the server
	_, err = conn.Write([]byte(request.toBytes()))
	checkError(err)
	// call to handle server response
	handleServer(conn, requestVersion)

// for debug
	// var buf[512]byte
	// _, err = conn.Read(buf[0:])
	// checkError(err)

	// response := string(buf[0:])

	
	// fmt.Println()

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
    fmt.Println("Enter method: ")
    fmt.Scanf("%s", &method)
    fmt.Println("Enter URL: ")
    fmt.Scanf("%s", &url)

    return method, url
}

func handleServer(conn net.Conn, requestVersion string) {
	// close the connection after this function executes
	defer conn.Close()

	// get message of at maximum 512 bytes
	var buf [512]byte
	// read input 
	_, err := conn.Read(buf[0:])
	// if there was an error exit
	checkError(err)
	// convert message to string and decompose it
	response := string(buf[0:])

	_, code, _, _, body := decomposeResponse(response)
	// call the function that decides which page gets launched
	launchPage(code, body, requestVersion)

}

func decomposeResponse(response string) (string, string, string, []string, string){
		const sp = "\x20"
		const cr = "\x0d"
		const lf = "\x0a"

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
		//check if there is any content in the body
		var bodyLines []string
		if i  < len(temp) {
			// get the body content
			bodyLines = temp[i:len(temp)]
		}
		body := strings.Join(bodyLines, cr + lf)

		// split the response line into it's components
		responses := strings.Split(responseLine, sp)
		status := responses[2]
		code := responses[1]
		version := responses[0]

		return version, code, status, headerLines, body

}

func launchPage(code string, body string, version string) {

	/*tempfile, err := ioutil.TempFile(os.TempDir(), "temp")
	checkError(err)

	defer os.Remove(tempfile.Name())*/

	var content string

	switch code {
		case "505": 
			content = fmt.Sprint("<HTML><HEAD>\n<TITLE>505 %s Not Supported</TITLE>\n</HEAD><BODY>\n<H1>505 ",version ," Not Supported</H1>\n</BODY></HTML>")
			break
		case "200":
			content = body
			break
		case "400":
			content = "<HTML><HEAD>\n<TITLE>400 Bad Request</TITLE>\n</HEAD><BODY>\n<H1>400 Bad Request</H1>\n</BODY></HTML>"
			break
		default:
			content = "<HTML><HEAD>\n<TITLE>Request expired</TITLE>\n</HEAD><BODY>\n<H1>Request expired</H1>\n</BODY></HTML>"
			break
	}
	/*_, err = tempfile.Write([]byte(content))
	checkError(err)*/

	// write contents to file
	err := ioutil.WriteFile("../../temp/launch_file.html", []byte(content), 0644)
	checkError(err)
	// launch default browser with .html file
	cmd := exec.Command("xdg-open", "../../temp/launch_file.html")
	err = cmd.Start()
	checkError(err)
	// err = cmd.Wait()
	// checkError(err)

/*	err = tempfile.Close()
	checkError(err)*/
}