package app

import (
	"fmt"
	"log"
	"net"
)

// Run start run server
func Run(address string) {
	// start listen in specify address
	listen, err := net.Listen("tcp", address)
	if err != nil {
		log.Fatalln(err)
	}
	log.Println(fmt.Sprintf("Server running on: http://%s", address))

	// to handler client connection within for loop
	for {
		// waits for and returns the next connection to the listener
		conn, _ := listen.Accept()
		go connectionHandler(conn)
	}
}

// connectionHandler to handler connection data
func connectionHandler(conn net.Conn) {
	// reads data from the connection
	bytes := make([]byte, MaxRequestSize)
	read, _ := conn.Read(bytes)

	// data handler
	dataHandler(conn, string(bytes[0:read]))
}

// dataHandler to handler receive data
func dataHandler(conn net.Conn, data string) {
	// get context from request data
	context := getRequestContext(data)

	// for loop in routers
	for _, v := range routers {
		// filter request
		if filterRequest(v, &context) {
			break
		}
	}

	// writes data to the connection client
	writeData(conn, context)
}
