package main

import (
	"net"
	"os"
	"fmt"
	"io/ioutil"
	"strings"
	"regexp"
	"strconv"
	"lib"
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
		config := lib.InitializeConfig()
		connectionType := os.Args[2]
		checkInput(config, host, connectionType)
		os.Exit(0)
	}
	// get the user to input the method to be used as well as the file/url requested
	method, url, body := getUserInputs()

	config := lib.InitializeConfig()

	if strings.ToUpper(config.Protocol) == "UDP" {
		config.Connection = "close"
	}

	handleRequest(method, url, body, host)
}

func handleRequest(method string, url string, body string, host string) {
	// read configuration
	config := lib.InitializeConfig()
	var dialHost string
	// check for proxy
	if config.Proxy != "off" {
		dialHost = config.Proxy
	} else {
		dialHost = host
	}
	// create connection
	conn, err := net.Dial(config.Protocol, dialHost)
	lib.CheckError(err)
	defer conn.Close()

	keepAlive:
	// set request message
	request := lib.SetRequestMessage(host, config, method, url, body)
	// write request to connection
	_, err = conn.Write(request.ToBytes())
	lib.CheckError(err)
	// get message
	var buf [65000]byte
	// read input 
	n, err := conn.Read(buf[0:])
	lib.CheckError(err)

	response := string(buf[0:n])
	version, code, status, headers, _ := decomposeResponse(response)

	var port string

	switch code {
		case "503":	
			handleRequest(method, url, body, host)
			return
		case "301","302":
			newHost, newUrl := getRedirectLocation(headers)
			if newHost == "localhost" {
				port = ":1235"
			} else {
				port = ":80"
			}
			newHost += port
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

	headerTemp := lib.NewResponseMessage()
	headerTemp.Version = version
	headerTemp.StatusCode = code
	headerTemp.Phrase = status
	headerTemp.HeaderLines = headers
	headerTemp.EntityBody = ""
	headerSize := len(headerTemp.ToBytes())
	lengthDiff := 0

	contentLen, err := strconv.Atoi(headers["Content-Length"])
	if err == nil {
		lengthDiff = contentLen + headerSize - 65000
	} else {
		lengthDiff = -1
	}
	if strings.ToUpper(headers["Transfer-Encoding"]) == "CHUNKED" {

		for {
			// get message
			var buf [65000]byte
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
			var buf [65000]byte
			// read input 
			n, err = conn.Read(buf[0:])
			lib.CheckError(err)
			response += string(buf[0:n])
			lengthDiff -= 65000
		}
		
	}

	_, _, _, _, body = decomposeResponse(response)

	writeReceivedToFile(body, getFileName(url))
	printToConsole(response)

	if strings.Contains(body, "src=\"") {

		sourceMap := retrieveSources(body)

		for host, url = range sourceMap {
			if host == "localhost" {
				port = ":1235"
			} else {
				port = ":80"
			}
			ip,_ := net.ResolveIPAddr("ip", host)
			if config.Connection != "keep-alive" || ip.String() != strings.Split(conn.LocalAddr().String(),":")[0] {
				fmt.Println("HERE")
				handleRequest("GET", url, "", host+port)
			} else {
				fmt.Println("OH NO")
				url = url
				host = host+port
				body = ""
				goto keepAlive
			}
		}
	}
}

func writeReceivedToFile(body string, fileName string) {

	if fileName == "/" {
		fileName = "index.html"
	}

	err := ioutil.WriteFile("../../temp/"+fileName, []byte(body), 0644)
	lib.CheckError(err)
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

func checkInput(config lib.ConfigSettings, host string, connectionType string) {
	switch host {
		case "protocol": 
			config.Protocol = connectionType
			break
		case "connection": 
			config.Connection = connectionType
			break
		case "proxy":
			config.Proxy = connectionType
			break
		default:
	}
	err := config.WriteConfig()
	lib.CheckError(err)
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