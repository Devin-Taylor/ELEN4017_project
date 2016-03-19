package main
import {
	"strconv"
}

type ResponseMessage struct{
	version string
	statusCode int
	phrase string
	headerLines map[string]string
	entityBody string
}

// Function to convert the HTTP Response to a string in the correct format
func (rm *ResponseMessage) toString() string {
	const sp := "\x20"
	const lf := "\x0a"
	const cr := "\x0d"
	responseString := rm.version + sp
	responseString += strconv.Itoa(rm.statusCode) + sp
	responseString += rm.phrase + cr + lf
	//add header lines
	for headerFieldName, value := range rm.headerLines {
		responseString += headerFieldName + ":" + sp
		responseString += value + cr + lf
	}
	responseString += cr + lf
	responseString += rm.entityBody
	return responseString
}

// Function to convert the HTTP Response to bytes in the correct format
func (rm * ResponseMessage) toBytes() []byte {
	return  []byte(rm.toString())
}
