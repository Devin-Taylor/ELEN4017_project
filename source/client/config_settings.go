package main

import (
	"io/ioutil"
	"strings"
	"os"
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
func initializeConfig() configSettings {
	var config configSettings
	configLines := readConfig()
	config.protocol = configLines[0]
	config.connection = configLines[1]
	config.proxy = configLines[2]

	return config
}

// write the new configuration settings to the file
func (config *configSettings) writeConfig() error {
	writeLines := config.protocol + "\n"
	writeLines += config.connection + "\n"
	writeLines += config.proxy
	err := ioutil.WriteFile("../../config/connection_config.txt", []byte(writeLines), 0644)
	return err
}

func (config *configSettings) checkInput(service string, connectionType string) {
	switch service {
		case "protocol": 
			config.protocol = connectionType
			break
		case "connection": 
			config.connection = connectionType
			break
		case "proxy":
			config.proxy = connectionType
			break
		default:
	}

	err := config.writeConfig()
	checkError(err)
	os.Exit(0)
}