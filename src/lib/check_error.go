// Author: James Allingham

package lib

import (
	"fmt"
	"os"
)

// function to report an error if it isn't nill
// inputs - an error to check
func CheckError(err error) {
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
	}
}