// Author: James Allingham

package lib

// struct that represents a HTTP response 
type ResponseMessage struct{
	Version string
	StatusCode string
	Phrase string
	HeaderLines map[string]string
	EntityBody string
}

// constructor for the ResponseMessage struct
func NewResponseMessage() *ResponseMessage {
    return &ResponseMessage{HeaderLines: make(map[string]string)}
}

// function to convert the HTTP Response to a string in the correct format
// outputs - a string that contains all the information for the response message in the format dictated by RFC7230
func (rm *ResponseMessage) ToString() string {
	const sp = "\x20"
	const lf = "\x0a"
	const cr = "\x0d"
	responseString := rm.Version + sp
	responseString += rm.StatusCode + sp
	responseString += rm.Phrase + cr + lf
	//add header lines
	for headerFieldName, value := range rm.HeaderLines {
		responseString += headerFieldName + ":" + sp
		responseString += value + cr + lf
	}
	responseString += cr + lf
	responseString += rm.EntityBody
	return responseString
}

// function to convert the HTTP Response to bytes in the correct format
// outputs - the bytes value of the string described above
func (rm * ResponseMessage) ToBytes() []byte {
	return  []byte(rm.ToString())
}
