// Author: Devin Taylor

package lib

import (
	"io/ioutil"
	"strings"
)
// struct containing all relevant configuration settings
type ConfigSettings struct {
	Protocol string
	Connection string
	Proxy string
}
// function responsible for reading contents of the cofiguration file
// outputs - string array containing all current configuration settings
func ReadConfig() []string {
	config, err := ioutil.ReadFile("../../config/connection_config.txt")
	CheckError(err)
	lines := strings.Split(string(config), "\n")
	return lines
}
// function responsible for initilizing the configuration settings based on what was read from the file
// outputs - config: a configSettings object with all configuration settings in a struct
func InitializeConfig() ConfigSettings {
	var config ConfigSettings
	configLines := ReadConfig()
	config.Protocol = configLines[0]
	config.Connection = configLines[1]
	config.Proxy = configLines[2]

	return config
}
// function responsible for writing the new configuration settings to the file
// outputs - err: an error corresponding to writing to file
func (config *ConfigSettings) WriteConfig() error {
	writeLines := config.Protocol + "\n"
	writeLines += config.Connection + "\n"
	writeLines += config.Proxy
	err := ioutil.WriteFile("../../config/connection_config.txt", []byte(writeLines), 0644)
	return err
}
// function responsible for checking the user inputs and edditing the configuration folder
// inputs - config: the configuration settings read in from a folder
//			configStatement: a string representing the configuration setting that must be changed
//			connectionType: the change that must be changed to configStatement
func (config *ConfigSettings) CheckInput(configStatement string, connectionType string) {
	switch configStatement {
		case "protocol": 
			config.Protocol = connectionType
			break
		case "connection": 
			config.Connection = connectionType
			break
		case "proxy":
			config.Proxy = connectionType
			break
		default:
	}
	// write the new configuration settings to the configuration file
	config.WriteConfig()
}