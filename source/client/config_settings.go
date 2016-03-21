package main

import (
	"io/ioutil"
	"strings"
)

type configSettings struct {
	protocol string
	connection string
	proxy string
}
 // read contents of the cofiguration file
func readConfig() []string {
	config, err := ioutil.ReadFile("../../config/connection_config.txt")
	checkError(err)
	lines := strings.Split(string(config), "\n")
	return lines
}

// initilize the configuration settings based on what was read from the file
func (config *configSettings) initializeConfig() {
	configLines := readConfig()
	config.protocol = configLines[0]
	config.connection = configLines[1]
	config.proxy = configLines[2]
}

// write the new configuration settings to the file
func writeConfig(config configSettings) error {
	writeLines := config.protocol + "\n"
	writeLines += config.connection + "\n"
	writeLines += config.proxy
	err := ioutil.WriteFile("../../config/connection_config.txt", []byte(writeLines), 0644)
	return err
}