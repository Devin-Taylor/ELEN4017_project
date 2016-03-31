// Author: Devin Taylor

package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"strings"
)
// function responsible for loading the map
// inputs - mapLocation: path to file location
// outputs - locationMap: a map of string to string
func loadMap(mapLocation string) map[string]string {
	locationMap := make(map[string]string)

	file, err := os.Open(mapLocation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		return locationMap
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	mapping := string(b)

	lines :=  strings.Split(mapping, "\n")
	lines = lines[0:len(lines)-1]
	// for all values loop and add to the map
	for _, value := range lines {
		locations := strings.SplitN(value, "\x20", 2)
		locationMap[locations[0]] = locations[1]
	}

	return locationMap
}
// function responsible for saving a map
// inputs - locationMap: a new map of string to string
// 			mapLocation: path the save file
func saveMap(locationMap map[string]string, mapLocation string) {
	var writeString string

	for key, value := range locationMap {
		writeString = writeString + key + "\x20" + value + "\n"
	}

	ioutil.WriteFile(mapLocation, []byte(writeString), 0644)
} 