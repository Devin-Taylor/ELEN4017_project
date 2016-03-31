package main

import (
	"time"
	"os"
	"fmt"
	"io/ioutil"
	"strings"
)

type roundTripTimer struct {
	timeMap map[string]string
	startTime time.Time
	duration time.Duration
}

func newRoundTripTimer() *roundTripTimer {
    return &roundTripTimer{timeMap: make(map[string]string)}
}

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

func (timer *roundTripTimer) startTimer() {
	timer.startTime = time.Now()
}

func (timer *roundTripTimer) stopTimer() {
	timer.duration = time.Since(timer.startTime)
}

func (timer *roundTripTimer) addToTimer(callType string) {
	// timeMilli := duration.Nanoseconds()
	timer.timeMap[callType] = timer.duration.String()
}

func (timer *roundTripTimer) writeTimerToFile(fileLocation string) {
	var writeString string

	for key, value := range timer.timeMap {
		writeString += value + " " + key + "\n"
	}

	ioutil.WriteFile(fileLocation, []byte(writeString), 0644)
}