package main

type RequestMessage struct {
	method string
	url string
	version string
	headerLines map[string]string
	entityBody string
}

