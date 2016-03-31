package lib

import (
	"io/ioutil"
	"strings"
	"os"
)

type ConfigSettings struct {
	Protocol string
	Connection string
	Proxy string
}

 // read contents of the cofiguration file
func ReadConfig() []string {
	config, err := ioutil.ReadFile("../../config/connection_config.txt")
	CheckError(err)
	lines := strings.Split(string(config), "\n")
	return lines
}

// initilize the configuration settings based on what was read from the file
func InitializeConfig() ConfigSettings {
	var config ConfigSettings
	configLines := ReadConfig()
	config.Protocol = configLines[0]
	config.Connection = configLines[1]
	config.Proxy = configLines[2]

	return config
}

// write the new configuration settings to the file
func (config *ConfigSettings) WriteConfig() error {
	writeLines := config.Protocol + "\n"
	writeLines += config.Connection + "\n"
	writeLines += config.Proxy
	err := ioutil.WriteFile("../../config/connection_config.txt", []byte(writeLines), 0644)
	return err
}

func (config *ConfigSettings) CheckInput(service string, ConnectionType string) {
	switch service {
		case "Protocol": 
			config.Protocol = ConnectionType
			break
		case "Connection": 
			config.Connection = ConnectionType
			break
		case "Proxy":
			config.Proxy = ConnectionType
			break
		default:
	}

	err := config.WriteConfig()
	CheckError(err)
	os.Exit(0)
}