package main

import (
	"net"
	"os"
	"os/exec"
	"fmt"
	// "bufio"
	"io/ioutil"
	"strings"
	"regexp"
	"bytes"
	"image"
	"image/jpeg"
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
	if len(os.Args) > 2 {
		connectionType := os.Args[2]
		checkInput(config, service, connectionType)
	}
	// initialize request message struct
	request := NewRequestMessage()
	// set header information
	request.setHeaders(service, config.connection, "Mozilla/5.0", "en")
	// check if proxy is required
	if strings.ToUpper(config.proxy) == "ON" {
		service = promptProxy()
	}
	// get the user to input the method to be used as well as the file/url requested
	method, url, entityBody := getUserInputs()
	// create connection
	conn, err := net.Dial(config.protocol, service)
	checkError(err)
	// set request version as need to when launch 505 error later on
	requestVersion := "HTTP/1.1"
	// set request line information
	request.setRequestLine(method, url, requestVersion)
	// set entity body
	request.setEntityBody(entityBody)
	// write request information to the server
	_, err = conn.Write([]byte(request.toBytes()))
	checkError(err)
	// call to handle server response
	handleServer(conn, method, config)

	os.Exit(0)
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}

func checkInput(config configSettings, service string, connectionType string) {
	switch service {
		case "protocol": 
			config.protocol = connectionType
			break
		case "connection": 
			config.connection = connectionType
			break
		case "proxy":
			config.proxy = connectionType
			break
		default:
	}

	err := writeConfig(config)
	checkError(err)
	os.Exit(0)
}

func getUserInputs() (string, string, string) {
	var method string
	var url string
    fmt.Println("Enter method ")
    fmt.Scanf("%s", &method)
    fmt.Println("Enter URL ")
    fmt.Scanf("%s", &url)

    var entityBody string

    if strings.ToUpper(method) == "POST" || strings.ToUpper(method) == "PUT" {
    	fmt.Println("Enter Text ")
    	fmt.Scanf("%s", &entityBody)
    } else {
    	entityBody = ""
    }


    return method, url, entityBody
}

func handleServer(conn net.Conn, method string, config configSettings) {
	// close the connection after this function executes
	defer conn.Close()

	// get message of at maximum 512 bytes
	var buf [8192]byte
	// read input 
	_, err := conn.Read(buf[0:])
	// if there was an error exit
	checkError(err)
	// convert message to string and decompose it
	response := string(buf[0:])

	version, code, status, _, body, headerLines := decomposeResponse(response)
	// if status = 200 then can be from multiple different requests

	printToConsole(version, code, status, headerLines, body)

	regexpString := "src=\"(.*?)\""
	
	if method != "HEAD" {

		isSource := checkForSources(body)

		if isSource {
			// Loop:
			sourceMap := retrieveSources(body, regexpString)

			for key, value := range sourceMap {

				request := NewRequestMessage()
				// set header information
				request.setHeaders(key, config.connection, "Mozilla/5.0", "en")
				conn, err := net.Dial(config.protocol, key+":80")
				checkError(err)
				// set request version as need to when launch 505 error later on
				requestVersion := "HTTP/1.1"
				// set request line information
				request.setRequestLine("GET", value, requestVersion)
				// set entity body
				request.setEntityBody("")
				// write request information to the server
				_, err = conn.Write([]byte(request.toBytes()))
				checkError(err)
				// call to handle server response

				StopIndex := strings.LastIndex(value, "/")
				fileName := value[StopIndex:len(value)]

				// var twoHundred bool

				/*twoHundred, body =*/ handlerServerSources(conn, "GET", fileName)

				// if !twoHundred {
				// 	regexpString = "href=\"(.*?)\""
				// 	goto Loop
				// } else {
				// 	cmd := exec.Command("xdg-open", "../../temp/"+fileName)
				// 	err = cmd.Start()
				// 	checkError(err)
				// }
			}
		}

		// launchPage(body)
	}
}

func handlerServerSources(conn net.Conn, method string, fileName string) {
	// close the connection after this function executes
	defer conn.Close()

	// get message of at maximum 512 bytes
	var buf [8192]byte
	// read input 
	_, err := conn.Read(buf[0:])
	// if there was an error exit
	checkError(err)
	// convert message to string and decompose it
	response := string(buf[0:])

	version, code, status, headers, body, headerLines := decomposeResponse(response)
	// if status = 200 then can be from multiple different requests

	printToConsole(version, code, status, headerLines, body)

	if code == "301" || code == "302" {
		httpUrl := headers["Location:"]

		httpUrl = strings.Split(httpUrl, "//")[1]

		splitURL := strings.SplitAfterN(httpUrl, "/", 2)

		key := strings.Replace(splitURL[0], "/", "", 2)
		value := "/" + splitURL[1]

		request := NewRequestMessage()
		// set header information
		request.setHeaders(key, "tcp", "Mozilla/5.0", "en")
		conn2, err := net.Dial("tcp", key+":80")
		checkError(err)
		defer conn2.Close()
		// set request version as need to when launch 505 error later on
		requestVersion := "HTTP/1.1"
		// set request line information
		request.setRequestLine("GET", value, requestVersion)
		// set entity body
		request.setEntityBody("")
		// write request information to the server
		_, err = conn2.Write([]byte(request.toBytes()))
		checkError(err)
		// call to handle server response

		StopIndex := strings.LastIndex(value, "/")
		fileName = value[StopIndex:len(value)]

		var buf [2*65136]byte
		// read input 
		_, err = conn2.Read(buf[0:])
		// if there was an error exit
		checkError(err)
		// convert message to string and decompose it
		response := string(buf[0:])

		version, code, status, _, body, headerLines = decomposeResponse(response)
		// if status = 200 then can be from multiple different requests

		printToConsole(version, code, status, headerLines, body)

		// 	cmd := exec.Command("xdg-open", "../../temp/"+fileName)
				// 	err = cmd.Start()
				// 	checkError(err)
	}

	if code != "301" {
		data := []byte("JPEG89a000\x0000000\x00\x00\x00\x000\x000\x02\b\r0000000\x00")
		var img, _, _ = image.Decode(bytes.NewReader(data))

		// img, _ := jpeg.Decode(bytes.NewReader([]byte(body)))
		fmt.Println("******************")
		fmt.Println(img)
		fmt.Println("******************")
		out,_ := os.Create("../../temp/"+fileName)
		err = jpeg.Encode(out, img, nil)

		// err = ioutil.WriteFile("../../temp" + fileName, []byte(body), 0644)
		// checkError(err)

		cmd := exec.Command("xdg-open", "../../temp/"+fileName)
		err = cmd.Start()
		checkError(err)
	}
}

func checkForSources(body string) bool {
	return strings.Contains(body, "src=\"")
}

func retrieveSources(body string, regexpString string) map[string]string {

	reg := regexp.MustCompile(regexpString)
	allMatches := reg.FindAllStringIndex(body, -1)
	var splitString string

	var sourceStrings []string

	urlToFileMap := make(map[string]string)

	for _, value := range allMatches {
		sourceStrings = append(sourceStrings, body[value[0]:value[1]])
	}

	if regexpString == "src=\"(.*?)\"" {
		splitString = "src=\"http://"
	} else {
		splitString = "href=\"http://"
	}

	for _, value := range sourceStrings {
		withoutHttp := strings.Split(value, splitString)
		splitURL := strings.SplitAfterN(withoutHttp[1], "/", 2)
		urlToFileMap[strings.Replace(splitURL[0], "/", "", 2)] = "/" + strings.Replace(splitURL[1], "\"", "", 2)
	}

	return urlToFileMap
}

func printToConsole(version string, code string, status string, headerLines []string, body string) {

	var allHeaders string

	for _, value := range headerLines {
		allHeaders = allHeaders + value + "\n"
	}

	content := version + " " + code + " " + status + "\n" + allHeaders + "\n\n" + body
	fmt.Println(content) 
}

func decomposeResponse(response string) (string, string, string, map[string]string, string, []string){
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
			bodyLines = temp[i:len(temp)]
		}
		body := strings.Join(bodyLines, cr + lf)

		// split the response line into it's components
		responses := strings.Split(responseLine, sp)
		status := responses[2]
		code := responses[1]
		version := responses[0]

		return version, code, status, headers, body, headerLines

}

func launchPage(body string) {

	err := ioutil.WriteFile("../../temp/launch_file.html", []byte(body), 0644)
	checkError(err)
	cmd := exec.Command("xdg-open", "../../temp/launch_file.html")
	err = cmd.Start()
	checkError(err)
}

func promptProxy() string {
	var proxyUrl string
    fmt.Println("Enter proxy URL:port ")
    fmt.Scanf("%s", &proxyUrl)	

    return(proxyUrl)
}