package main // for you Devin

import (
	"os"
	"fmt"
	"io/ioutil"
	"strings"
)

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

	for _, value := range lines {
		locations := strings.SplitN(value, "\x20", 2)

		locationMap[locations[0]] = locations[1]
	}

	return locationMap
}

func saveMap(locationMap map[string]string, mapLocation string) {

	var writeString string

	for key, value := range locationMap {
		writeString = writeString + key + "\x20" + value + "\n"
	}

	ioutil.WriteFile(mapLocation, []byte(writeString), 0644)
} 