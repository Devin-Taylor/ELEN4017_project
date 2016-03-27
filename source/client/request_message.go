package main

type RequestMessage struct {
	method string
	url string
	version string
	headerLines map[string]string
	entityBody string
}

func NewRequestMessage() *RequestMessage {
    return &RequestMessage{headerLines: make(map[string]string)}
}

func (rm *RequestMessage) toString() string {
	const sp = "\x20"
	const lf = "\x0a"
	const cr = "\x0d"
	requestString := rm.method + sp
	requestString += rm.url + sp
	requestString += rm.version + cr + lf
	//add header lines
	for headerFieldName, value := range rm.headerLines {
		requestString += headerFieldName + ":" + sp
		requestString += value + cr + lf
	}
	requestString += cr + lf
	requestString += rm.entityBody
	return requestString
}

// Function to convert the HTTP request to bytes in the correct format
func (rm *RequestMessage) toBytes() []byte {
	return  []byte(rm.toString())
}

func (rm *RequestMessage) setHeaders(host string, connection string, userAgent string, language string) {

	rm.headerLines["Host"] = host
	rm.headerLines["Connection"] = connection
	rm.headerLines["User-agent"] = userAgent
	rm.headerLines["language"] = language
}

func (rm *RequestMessage) setRequestLine(method string, url string, version string) {
	rm.method = method
	rm.url = url
	rm.version = version
}

func (rm *RequestMessage) setEntityBody(body string) {
	rm.entityBody = body
}

func setRequestMessage(service string, config configSettings, method string, url string, body string) *RequestMessage {
	request := NewRequestMessage()

	request.setHeaders(service, config.connection, "Mozilla/5.0", "en")
	request.setRequestLine(method, url, "HTTP/1.1")
	request.setEntityBody(body)

	return request
}