package main // for you Devin

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
	// check if proxy is required
	if strings.ToUpper(config.proxy) == "ON" {
		service = promptProxy()
	}
	// get the user to input the method to be used as well as the file/url requested
	method, url, entityBody := getUserInputs()
	// set request message
	request := setRequestMessage(service, config, method, url, entityBody)
	// connect and write to server
	conn := dialAndSend(config.protocol, service, request)
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

func dialAndSend(protocol string, service string, request *RequestMessage) net.Conn {
	conn, err := net.Dial(protocol, service)
	checkError(err)
	_, err = conn.Write([]byte(request.toBytes()))
	checkError(err)

	return conn
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

	// get message
	var buf [8192]byte
	// read input 
	_, err := conn.Read(buf[0:])
	// if there was an error exit
	checkError(err)
	// convert message to string and decompose it
	response := string(buf[0:])

	version, code, status, headers, body := decomposeResponse(response)
	// if status = 200 then can be from multiple different requests

	printToConsole(version, code, status, headers, body)
	
	if method != "HEAD" {

		if checkForSources(body) {
			sourceMap := retrieveSources(body)

			for key, value := range sourceMap {
				// compile the request message
				request := setRequestMessage(key, config, "GET", value, "")

				port := ":80"
				if key == "localhost" {
					port = ":1235"
				}

				if ip,_ := net.ResolveIPAddr("ip", key); ip.String() != strings.Split(conn.LocalAddr().String(),":")[0] || strings.ToUpper(config.connection) != "KEEP-ALIVE" {
					
					conn = dialAndSend(config.protocol, key+port, request)
				} else {
					_, err = conn.Write([]byte(request.toBytes()))
					checkError(err)
				}

				fileName := getFileName(value)

				handlerServerSources(conn, "GET", fileName, config)
			}
		}
		// launchPage(body)
	}
}

func getFileName(value string) string {
	StopIndex := strings.LastIndex(value, "/")
	fileName := value[StopIndex:len(value)]

	return fileName
}

func handlerServerSources(conn net.Conn, method string, fileName string, config configSettings) {
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

	version, code, status, headers, body := decomposeResponse(response)
	// if status = 200 then can be from multiple different requests

	printToConsole(version, code, status, headers, body)

	for code != "200" {

		if strings.Split(code, "")[0] != "3" {
			break
		}

		httpUrl := headers["Location:"]

		httpUrl = strings.Split(httpUrl, "//")[1]

		splitURL := strings.SplitAfterN(httpUrl, "/", 2)

		key := strings.Replace(splitURL[0], "/", "", 2)
		value := "/" + splitURL[1]

		// compile the request message
		request := setRequestMessage(key, config, "GET", value, "")

		port := ":80"
		if key == "localhost" {
			port = ":1235"
		}

		conn2 := dialAndSend(config.protocol, key+port, request)

		defer conn2.Close()

		fileName = getFileName(value)

		var buf [2048]byte
		// read input 
		_, err = conn2.Read(buf[0:])
		// if there was an error exit
		checkError(err)
		// convert message to string and decompose it
		response := string(buf[0:])

		version, code, status, headers, body = decomposeResponse(response)
		// if status = 200 then can be from multiple different requests

		printToConsole(version, code, status, headers, body)
	}

	img, _, _ := image.Decode(bytes.NewReader([]byte(body)))
	out,_ := os.Create("../../temp"+fileName)
	err = jpeg.Encode(out, img, nil)

	cmd := exec.Command("xdg-open", "../../temp/"+fileName)
	err = cmd.Start()
	checkError(err)

}

func checkForSources(body string) bool {
	return strings.Contains(body, "src=\"")
}

func retrieveSources(body string) map[string]string {

	reg := regexp.MustCompile("src=\"(.*?)\"")
	allMatches := reg.FindAllStringIndex(body, -1)
	// var splitString string

	var sourceStrings []string

	urlToFileMap := make(map[string]string)

	for _, value := range allMatches {
		sourceStrings = append(sourceStrings, body[value[0]:value[1]])
	}

	// if regexpString == "src=\"(.*?)\"" {
	// 	splitString = "src=\"http://"
	// } else {
	// 	splitString = "href=\"http://"
	// }

	for _, value := range sourceStrings {
		withoutHttp := strings.Split(value, "src=\"http://")
		splitURL := strings.SplitAfterN(withoutHttp[1], "/", 2)
		urlToFileMap[strings.Replace(splitURL[0], "/", "", 2)] = "/" + strings.Replace(splitURL[1], "\"", "", 2)
	}

	return urlToFileMap
}

func printToConsole(version string, code string, status string, headerLines map[string]string, body string) {

	var allHeaders string

	for key, value := range headerLines {
		allHeaders = allHeaders + key + " " + value + "\n"
	}

	content := version + " " + code + " " + status + "\n" + allHeaders + "\n\n" + body
	fmt.Println(content) 
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
			line := strings.SplitN(value, sp, 2)
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