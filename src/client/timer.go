// Author: Devin Taylor

package main

import (
	"time"
	"os"
	"fmt"
	"io/ioutil"
	"strings"
)
// struct containing all roundTripTimer attributed data
type roundTripTimer struct {
	timeMap map[string]string
	startTime time.Time
	duration time.Duration
}
// function responsible for initiating a new round trip timers
// outputs - roundTripTimer: a pointer to a roundTripTimer object
func newRoundTripTimer() *roundTripTimer {
    return &roundTripTimer{timeMap: make(map[string]string)}
}
// function responsible for loading the prexisting timer map from a file so data is not overwritten
// inputs - mapLocation: the path to the map
func (timer *roundTripTimer) loadTimerMap(mapLocation string) {

	tempMap := make(map[string]string)

	file, err := os.Open(mapLocation)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s", err.Error())
		timer.timeMap = tempMap
	}
	defer file.Close()

	b, err := ioutil.ReadAll(file)
	mapping := string(b)

	lines :=  strings.Split(mapping, "\n")
	lines = lines[0:len(lines)-1]

	for _, value := range lines {
		locations := strings.SplitN(value, "\x20", 2)

		tempMap[locations[1]] = locations[0]
	}

	timer.timeMap = tempMap
}
// function starts the timer
func (timer *roundTripTimer) startTimer() {
	timer.startTime = time.Now()
}
// function stops the timer and calculates duration
func (timer *roundTripTimer) stopTimer() {
	timer.duration = time.Since(timer.startTime)
}
// function adds the new time obtained to the time map
// inputs - callType: the characteristics corresponding to the time obtained
func (timer *roundTripTimer) addToTimer(callType string) {
	timer.timeMap[callType] = timer.duration.String()
}
// function responsible for writing the new timer map to a file so the data can be analysed
// inputs - fileLocation: the directory to the file for saving
func (timer *roundTripTimer) writeTimerToFile(fileLocation string) {
	var writeString string

	for key, value := range timer.timeMap {
		writeString += value + " " + key + "\n"
	}
	ioutil.WriteFile(fileLocation, []byte(writeString), 0644)
}