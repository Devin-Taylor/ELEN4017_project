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
	- `go run *.go proxy off/proxyIP:port`: sets the proxy to on or to connection to the proxy on the specified address
- `go run *.go <options>`
	- `go run *.go destinationIP:port`: dials into the server on the specified address
	- `go run *.go print-config`: prints the current configuration settings to screen

__NOTE:__ The proxy and server must be running before the client can be run. If the proxy is set to off then only the server needs to be running

**User Inputs**

When dialing into the server the client must specify the host IP or DNS address:
> `go run *.go localhost:1235` or `go run *.go www.amazon.com:80` (port :1235 was specified as the localhost port for the server and :1236 was specified as the localhost port for the proxy)

Once the user runs the client connection to the server the user will be prompted to enter the method, this can be of the form of one of the following: GET, HEAD, PUT, POST, DELETE. The user will then be prompted to enter the URL, this is the location of the desired file of the server. 
> localhost: `/index.html` or for amazon home page `/`

__NOTE:__ The URL **MUST** begin with a single forwardslash

If the method that was specified was either PUT or POST the user will then be prompted to enter the body of the message, this can be in the form of anything BUT if the user desires it to be an HTML page then the user must enter the text in full HTML format. 