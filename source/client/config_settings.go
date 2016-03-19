package main

import (
	"io/ioutil"
	"strings"
)

type configSettings struct {
	protocol string
	connection string
}

func readConfig() []string {
	config, err := ioutil.ReadFile("../../config/connection_config.txt")
	checkError(err)

	lines := strings.Split(string(config), "\n")

	return lines
}

func (config *configSettings) initializeConfig() {
	
	configLines := readConfig()

	config.protocol = configLines[0]
	config.connection = configLines[1]
}

func writeConfig(config configSettings) error {
	writeLines := config.protocol + "\n"
	writeLines += config.connection

	err := ioutil.WriteFile("../../config/connection_config.txt", []byte(writeLines), 0644)

	return err
}

func (config *configSettings) setProtocol(protocol string) {
	config.protocol = protocol
}

func (config *configSettings) setConnection(connection string) {
	config.connection = connection
}