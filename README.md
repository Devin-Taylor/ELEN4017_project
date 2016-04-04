# ELEN4017_project
## The project entails the implementation of a client, server and proxy intereaction through the use of the Go programming language

**Prerequisits**

Follow the instructions on the [Go Webpage](https://golang.org/doc/install) to install Go.

**Directory Structure**

- Project Root
	- cache: proxy cached files and maps
	- config: system configuration settings
	- documentation: report and timer related data files
	- objects: all server objects
	- src: all source code
	- temp: temporary directory for launching client files

**Using The Application** 

In order to run the server, proxy and client the user is required to navigate to the respective folder, for example `/src/server/` from this directory the user can execute the command `go run *.go`. By default the server and proxy do not accept any arguements. The client is configured to accept two different type of aguements, these are:

- `go run *.go <config> <new setting>`
	- `go run *.go protocol upd/tcp`: sets the protocol to either TCP or UDP
	- `go run *.go connection close/keep-alive`: sets the connection type to either non-persistent or persistent
	- `go run *.go proxy off/IP:port`: sets the proxy to on or to connection to the proxy on the specified address
- `go run *.go <options>`
	- `go run *.go destinationIP:port`: dials into the server on the specified address
	- `go run *.go print-config`: prints the current configuration settings to screen