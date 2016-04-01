// Author: James Allingham

package lib

import (
	"os"
)

// function to check if a file or directory exists
// inputs - string of the path to check
// outputs - a bool containing whether or not the file exists
//         - an error describing what the problem is with the file if the above is false or nil otherwise
func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { return true, nil }
	if os.IsNotExist(err) { return false, nil }
	return true, err
}