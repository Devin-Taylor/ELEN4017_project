// Author: James Allingham

package lib

import (
	"strings"
)

func DecomposeRequest(request string) (string, string, string, map[string]string, string){
		const sp = "\x20"
		const cr = "\x0d"
		const lf = "\x0a"
		headers := make(map[string]string)

		temp := strings.Split(request, cr + lf)
		// get the request line for further processing
		requestLine := temp[0]
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
			
			line := strings.SplitN(value, ":"+" ", 2)
			
			headers[line[0]] = line[1]
		}
		//check if there is any content in the body
		var bodyLines []string
		if i  < len(temp) {
			// get the body content
			bodyLines = temp[i+1:len(temp)]
		}
		body := strings.Join(bodyLines, cr + lf)

		// split the request line into it's components
		requests := strings.Split(requestLine, sp)
		method := requests[0]
		url := requests[1]
		version := requests[2]

		return method, url, version, headers, body
	}