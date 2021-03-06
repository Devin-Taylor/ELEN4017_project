// Author: Devin Taylor

package lib

import (
	"strings"
)
// function responsible for decomposing the received string from the server into the required components
// inputs - response: string containing entire server response
// outputs - version: string representing HTTP version the server supports
//			 code: string representing the response code
//			 status: string representing the response status
// 			 headers: map of headers to their contents
// 			 body: string containing reponse message entityBody
func DecomposeResponse(response string) (string, string, string, map[string]string, string){
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
