package main

import (
	"net"
	"os"
	"fmt"
	"io/ioutil"
	"strings"
	"regexp"
	"strconv"
)

func main() {
	// get the arguments passed to the code
	host := os.Args[1]
	// if no arguments replied then error
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port  message", os.Args[0])
		os.Exit(1)
	}
	// determine of the input required config settings to be changed
	if len(os.Args) > 2 {
		// initialize config settings variables
		config := initializeConfig()
		connectionType := os.Args[2]
		checkInput(config, host, connectionType)
		os.Exit(0)
	}
	// get the user to input the method to be used as well as the file/url requested
	method, url, body := getUserInputs()

	handleRequest(method, url, body, host)

}

func handleRequest(method string, url string, body string, host string) {
	// read configuration
	config := initializeConfig()
	// set request message
	request := setRequestMessage(host, config, method, url, body)
	// check for proxy
	if config.proxy != "off" {
		host = config.proxy
	}
	// create connection
	conn, err := net.Dial(config.protocol, host)
	checkError(err)
	// write request to connection
	_, err = conn.Write(request.toBytes())
	checkError(err)
	// get message
	var buf [4000]byte
	// read input 
	n, err := conn.Read(buf[0:])
	checkError(err)

	response := string(buf[0:])

	version, code, status, headers, _ := decomposeResponse(response)

	switch code {
		case "503":	
			handleRequest(method, url, body, host)
			return
		case "301","302":
			newHost, newUrl := getRedirectLocation(headers)
			newHost += ":80"
			fmt.Println(newHost, host, newUrl, url)
			if newHost == "" && newUrl == "" {
				break
			}
			if newHost == host && newUrl == url {
				break
			}
			handleRequest(method, newUrl, body, newHost)
			return
		default:
	}

	headerTemp := NewResponseMessage()
	headerTemp.version = version
	headerTemp.statusCode = code
	headerTemp.phrase = status
	headerTemp.headerLines = headers
	headerTemp.entityBody = ""
	headerSize := len(headerTemp.ToBytes())
	lengthDiff := 0

	contentLen, err := strconv.Atoi(headers["Content-Length"])
	if err == nil {
		lengthDiff = contentLen + headerSize - 4000
	} else {
		lengthDiff = -1
	}

	if strings.ToUpper(headers["Transfer-Encoding"]) == "CHUNKED" {

		for {
			// get message
			var buf [4000]byte
			// read input 
			n, err = conn.Read(buf[0:])
			checkError(err)
			response += string(buf[0:])
			if strings.Contains(response, "\r\n0\r\n\r\n") || n == 0 {
					break
			}
		}
	} else {
		for lengthDiff > 0 {
			var buf [4000]byte
			// read input 
			n, err = conn.Read(buf[0:])
			checkError(err)
			response += string(buf[0:])
			lengthDiff -= 4000
		}
		
	}

	_, _, _, _, body = decomposeResponse(response)

	writeReceivedToFile(body, getFileName(url))
	printToConsole(response)

	if strings.Contains(body, "src=\"") {

		sourceMap := retrieveSources(body)
		fmt.Println(sourceMap)

		for host, url := range sourceMap {
			handleRequest("GET", url, "", host+":80")
		}
	}
}

func writeReceivedToFile(body string, fileName string) {

	if fileName == "/" {
		fileName = "index.html"
	}

	err := ioutil.WriteFile("../../temp/"+fileName, []byte(body), 0644)
	checkError(err)
}

func getRedirectLocation(headers map[string]string) (string, string) {
	location := headers["Location"]

	httpUrl := strings.Split(location, "//")

	if len(httpUrl) < 2 {
		return "", ""
	}
	splitURL := strings.SplitAfterN(httpUrl[1], "/", 2)
	if len(splitURL) < 2 {
		return splitURL[0], "/"
	}
	return strings.Replace(splitURL[0], "/", "", 2), "/" + splitURL[1]	
}

func checkInput(config configSettings, host string, connectionType string) {
	switch host {
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
	err := config.writeConfig()
	checkError(err)
}

func getUserInputs() (string, string, string) {
	var method string
	var url string
    fmt.Println("Enter method ")
    fmt.Scanf("%s", &method)
    fmt.Println("Enter URL ")
    fmt.Scanf("%s", &url)

    //convert method to upper case
    method = strings.ToUpper(method)

    var entityBody string

    if method == "POST" || method == "PUT" {
    	fmt.Println("Enter Text ")
    	fmt.Scanf("%s", &entityBody)
    } else {
    	entityBody = ""
    }
    return method, url, entityBody
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
		responses := strings.SplitN(responseLine, sp, 3)
		status := responses[2]
		code := responses[1]
		version := responses[0]

		return version, code, status, headers, body

}

func getFileName(value string) string {
	StopIndex := strings.LastIndex(value, "/")
	fileName := value[StopIndex:len(value)]

	return fileName
}

func retrieveSources(body string) map[string]string {

	reg := regexp.MustCompile("src=\"(.*?)\"")
	allMatches := reg.FindAllStringIndex(body, -1)

	var sourceStrings []string

	urlToFileMap := make(map[string]string)

	for _, value := range allMatches {
		sourceStrings = append(sourceStrings, body[value[0]:value[1]])
	}

	for _, value := range sourceStrings {
		withoutHttp := strings.Split(value, "//")
		splitURL := strings.SplitAfterN(withoutHttp[1], "/", 2)
		urlToFileMap[strings.Replace(splitURL[0], "/", "", 2)] = "/" + strings.Replace(splitURL[1], "\"", "", 2)
	}
	return urlToFileMap
}


func printToConsole(response string) {

	version, code, status, headers, body := decomposeResponse(response)

	var allHeaders string

	for key, value := range headers {
		allHeaders = allHeaders + key + ": " + value + "\n"
	}

	content := version + " " + code + " " + status + "\n" + allHeaders + "\n\n" + body
	fmt.Println(content) 
}

func checkError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Fatal error: %s", err.Error())
		os.Exit(1)
	}
}