// Author: James Allingham

package lib

type ResponseMessage struct{
	Version string
	StatusCode string
	Phrase string
	HeaderLines map[string]string
	EntityBody string
}

func NewResponseMessage() *ResponseMessage {
    return &ResponseMessage{HeaderLines: make(map[string]string)}
}

// Function to convert the HTTP Response to a string in the correct format
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

// Function to convert the HTTP Response to bytes in the correct format
func (rm * ResponseMessage) ToBytes() []byte {
	return  []byte(rm.ToString())
}
