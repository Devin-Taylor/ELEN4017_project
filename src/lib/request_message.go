// Author: Devin Taylor

package lib

type RequestMessage struct {
	Method string
	Url string
	Version string
	HeaderLines map[string]string
	EntityBody string
}

func NewRequestMessage() *RequestMessage {
    return &RequestMessage{HeaderLines: make(map[string]string)}
}

func (rm *RequestMessage) ToString() string {
	const sp = "\x20"
	const lf = "\x0a"
	const cr = "\x0d"
	requestString := rm.Method + sp
	requestString += rm.Url + sp
	requestString += rm.Version + cr + lf
	//add header lines
	for headerFieldName, value := range rm.HeaderLines {
		requestString += headerFieldName + ":" + sp
		requestString += value + cr + lf
	}
	requestString += cr + lf
	requestString += rm.EntityBody
	return requestString
}

// Function to convert the HTTP request to bytes in the correct format
func (rm *RequestMessage) ToBytes() []byte {
	return  []byte(rm.ToString())
}

func (rm *RequestMessage) SetHeaders(host string, connection string, userAgent string, language string) {

	rm.HeaderLines["Host"] = host
	rm.HeaderLines["Connection"] = connection
	rm.HeaderLines["User-agent"] = userAgent
	rm.HeaderLines["language"] = language
}

func (rm *RequestMessage) SetRequestLine(Method string, Url string, Version string) {
	rm.Method = Method
	rm.Url = Url
	rm.Version = Version
}

func (rm *RequestMessage) SetEntityBody(body string) {
	rm.EntityBody = body
}

func SetRequestMessage(service string, config ConfigSettings, Method string, Url string, body string) *RequestMessage {
	request := NewRequestMessage()

	request.SetHeaders(service, config.Connection, "Mozilla/5.0", "en")
	request.SetRequestLine(Method, Url, "HTTP/1.0")
	request.SetEntityBody(body)

	return request
}