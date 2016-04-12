// Author: Devin Taylor

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
	// if no arguments supplied then error
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage: %s host:port  message", os.Args[0])
		os.Exit(1)
	}
	// initialise configuration setting
	config := lib.InitializeConfig()
	// check if user wants to print config settings
	if strings.ToUpper(host) == "PRINT-CONFIG" {
		fmt.Println(config)
		os.Exit(0)
	}
	// determine if the input requires config settings to be changed
	if len(os.Args) > 2 {
		connectionType := os.Args[2]
		config.CheckInput(host, connectionType)
		os.Exit(0)
	}
	// get the user to input the method to be used as well as the file/url requested and body for posting/putting
	method, url, body := getUserInputs()
	// if the protocol is UDP the force the connection to be a closed connection
	if strings.ToUpper(config.Protocol) == "UDP" {
		config.Connection = "close"
	}
	// timer := newRoundTripTimer()
	// timer.loadTimerMap("../../documentation/timer_map1.txt")
	// timer.startTimer()
	handleRequest(method, url, body, host)
	// timer.stopTimer()
	// timer.addToTimer(config.Protocol + " " + config.Connection + " " + config.Proxy)
	// timer.writeTimerToFile("../../documentation/timer_map1.txt")
}
// function handles the majority of responsibility for the client - responsible for subsequent function calls
// inputs - method: string that describes the HTTP method to be used
//			url: string that describes the destination URL from the host website
//			host: string that describes the http:// host destination and port number (eg. www.hostpage.com:80)
func handleRequest(method string, url string, body string, host string) {
	// read configuration
	config := lib.InitializeConfig()
	var dialHost string
	// check if proxy is on, if it is then change the destication
	if config.Proxy != "off" {
		dialHost = config.Proxy
	} else {
		dialHost = host
	}
	// create connection to server (either web server or proxy server depending on configuration)
	conn, err := net.Dial(config.Protocol, dialHost)
	lib.CheckError(err)
	defer conn.Close()
	// KeepAlive is a loop label that is used in a latter goto statement - used for keep-alive connections
	keepAlive:
	// compose request message
	request := lib.SetRequestMessage(host, config, method, url, body)
	// write request to connection
	_, err = conn.Write(request.ToBytes())
	lib.CheckError(err)
	// set message buffer, 4096 bytes is message maximum
	var buf [4096]byte
	// read input from connection
	n, err := conn.Read(buf[0:])
	lib.CheckError(err)
	// parse connection read into string
	response := string(buf[0:n])
	// decompose the message received from the server
	version, code, status, headers, _ := lib.DecomposeResponse(response)
	var port string
	// switch statement determines whether the page needs to be redirected or re-requested
	switch code {
		case "503":	
			handleRequest(method, url, body, host)
			return
		case "301","302":
			// if redirected get new destination
			newHost, newUrl := getRedirectLocation(headers)
			if newHost == "localhost" || host == strings.Split(conn.RemoteAddr().String(), ":")[0] || host == conn.RemoteAddr().String() {
				port = ":80"
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
			// call the same function - recursive until the correct page is obtained
			handleRequest(method, newUrl, body, newHost)
			return
		default:
	}
	// get the header size to determine how much more of the file needs to read
	headerSize := getHeaderSize(version, code, status, headers)
	lengthDiff := 0
	// get content length from headers in reply from server
	contentLen, err := strconv.Atoi(headers["Content-Length"])
	if err == nil {
		// get the remainer of data that needs to be read
		lengthDiff = contentLen + headerSize - 4096
	} else {
		lengthDiff = -1
	}
	// if the header mentions chunked then need to read more data (HTTP/1.1)
	if strings.ToUpper(headers["Transfer-Encoding"]) == "CHUNKED" {
		// itterate until all data is read in
		for {
			var buf [4096]byte
			// read input 
			n, err = conn.Read(buf[0:])
			lib.CheckError(err)
			response += string(buf[0:n])
			// break if EOF character is read in or if there is not more data to read in
			if strings.Contains(response, "\r\n0\r\n\r\n") || n == 0 {
					break
			}
		}
	} else {
		// itterate until no more data to read in 
		
		for lengthDiff > 0 {
			var buf [4096]byte
			// read input 
			n, err = conn.Read(buf[0:])
			lib.CheckError(err)
			response += string(buf[0:n])
			lengthDiff -= 4096
		}
		
	}
	// get the entire message body once all read in
	_, _, _, _, body = lib.DecomposeResponse(response)
	// write the file received to local directory so can be launched
	writeReceivedToFile(body, getFileName(url))
	printToConsole(response)
	// if the HTML text contains a reference to a source then find the sources
	if strings.Contains(body, "src=\"") {
		// get the map of all sources - map of hosts to corresponding urls
		sourceMap := retrieveSources(body)
		// for each source - fetch it
		for host, url = range sourceMap {
			if host == "localhost" || host == strings.Split(conn.RemoteAddr().String(), ":")[0] || host == conn.RemoteAddr().String() {
				port = ":80"
			} else {
				port = ":80"
			}
			ip,_ := net.ResolveIPAddr("ip", host)
			// if the connection is set to close of the IP is different from the original HTML IP the establish a new connection
			if config.Connection != "keep-alive" || ip.String() != strings.Split(conn.RemoteAddr().String(),":")[0] {
				handleRequest("GET", url, "", host+port)
			} else {

				url = url
				host = host+port
				body = ""
				// goto beginning of function and skip establishing new connection (for keep-alive set to true)
				goto keepAlive
			}
		}
	}
}
// function is responsible for writing a set of text to a file
// inputs - body: the text to  be written to a file
//			fileName: the path to the file to which the body must be written
func writeReceivedToFile(body string, fileName string) {
	// if filename is blank then just give it a blank names
	if fileName == "/" {
		fileName = "index.html"
	}
	err := ioutil.WriteFile("../../temp/"+fileName, []byte(body), 0644)
	lib.CheckError(err)
}
// function is responsible for getting the location to which the new request must point to
// inputs - headers: the headers of the message received from the server (will contain 3XX message)
// outputs -host: a string representing the new host
//			url: a string representing the directory to the new file in the host	
func getRedirectLocation(headers map[string]string) (string, string) {
	location := headers["Location"]
	// split to remove any http
	httpUrl := strings.Split(location, "//")

	if len(httpUrl) < 2 {
		return "", ""
	}
	// split host from url
	splitURL := strings.SplitAfterN(httpUrl[1], "/", 2)
	if len(splitURL) < 2 {
		return splitURL[0], "/"
	}
	return strings.Replace(splitURL[0], "/", "", 2), "/" + splitURL[1]	
}
// function responsible for retrieving the user inputs
// outputs - method: string representing the HTTP request statement
//			 url: string representing the required destination
//			 entityBody: string representing the information that must be PUT/POST'd
func getUserInputs() (string, string, string) {
	var method string
	var url string
    fmt.Println("Enter method ")
    fmt.Scanf("%s", &method)
    fmt.Println("Enter URL ")
    fmt.Scanf("%s", &url)

    method = strings.ToUpper(method)

    var entityBody string
    // only request an entityBody if the method is POST of PUT
    if method == "POST" || method == "PUT" {
    	fmt.Println("Enter Text ")
    	fmt.Scanf("%s", &entityBody)
    } else {
    	entityBody = ""
    }
    return method, url, entityBody
}
// function responsible for getting the name of the file from a URL
// inputs - value: string representing the URL that the fileName must be extracted from
// outputs - fileName: string representing the name of the file
func getFileName(value string) string {
	StopIndex := strings.LastIndex(value, "/")
	fileName := value[StopIndex:len(value)]

	return fileName
}
// function responsible for obtaining a map of host to url from the body of HTML text returned from the server
// inputs - body: string representing the entityBody of the recieved message
// outputs - urlToFileMap: map of sources host to URL
func retrieveSources(body string) map[string]string {
	// find all occurances of src in the file
	reg := regexp.MustCompile("src=\"(.*?)\"")
	allMatches := reg.FindAllStringIndex(body, -1)

	var sourceStrings []string

	urlToFileMap := make(map[string]string)
	// for all the occurances of src extract everything the is referenced
	for _, value := range allMatches {
		sourceStrings = append(sourceStrings, body[value[0]:value[1]])
	}
	// for all sources extract the new host and url and parse it into the map
	for _, value := range sourceStrings {
		withoutHttp := strings.Split(value, "//")
		splitURL := strings.SplitAfterN(withoutHttp[1], "/", 2)
		urlToFileMap[strings.Replace(splitURL[0], "/", "", 2)] = "/" + strings.Replace(splitURL[1], "\"", "", 2)
	}
	return urlToFileMap
}
// function responsible for printing the servers reply to the console
// inputs - response: string containing the servers response including headers and body
func printToConsole(response string) {
	version, code, status, headers, _ := lib.DecomposeResponse(response)
	var allHeaders string

	for key, value := range headers {
		allHeaders = allHeaders + key + ": " + value + "\n"
	}

	content := version + " " + code + " " + status + "\n" + allHeaders + "\n"
	fmt.Println(content) 
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