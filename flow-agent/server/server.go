package server

import (
	"bufio"
	"encoding/base64"
	"flow-agent/client"
	"flow-agent/commons"
	"flow-agent/config"
	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/emersion/go-sasl"
	"github.com/pkg/errors"
)

// A function that creates SASL servers.
type SASLServerFactory func(*Endpoint) sasl.Server

// A struct with a mix of fields, used for the GOB example.
// type complexData struct {
// 	N int
// 	S string
// 	M map[string]int
// 	P []byte
// 	C *complexData
// }

/*
## Incoming connections

Preparing for incoming data is a bit more involved. According to our ad-hoc
protocol, we receive the name of a command terminated by `\n`, followed by data.
The nature of the data depends on the respective command. To handle this, we
create an `Endpoint` object with the following properties:

* It allows to register one or more handler functions, where each can handle a
  particular command.
* It dispatches incoming commands to the associated handler based on the commands
  name.

*/

// HandleFunc is a function that handles an incoming command.
// It receives the open connection wrapped in a `ReadWriter` interface.
// type HandleFunc func(*bufio.ReadWriter)

type HandleFunc func(*Endpoint, net.Conn, *bufio.ReadWriter, *bufio.Scanner, config.Config)

// Endpoint provides an endpoint to other processess
// that they can send data to.
type Endpoint struct {
	listener      net.Listener
	handler       map[string]HandleFunc
	auths         map[string]SASLServerFactory
	connection    net.Conn
	rw            *bufio.ReadWriter
	authenticated bool
	// Maps are not threadsafe, so we need a mutex to control access.
	m sync.RWMutex
}

// server listens for incoming requests and dispatches them to
// registered handler functions.
func Server(config config.Config) error {
	endpoint := NewEndpoint()

	// Add the handle funcs.
	// endpoint.AddHandleFunc("STRING", handleStrings)
	// endpoint.AddHandleFunc("LIST", handleStrings)
	endpoint.AddHandleFunc("quit", handleQuit)
	endpoint.AddHandleFunc("authenticate", handleAuth)
	endpoint.AddHandleFunc("create", handleCreate)
	endpoint.AddHandleFunc("subscribe", handleSubscribe)
	endpoint.AddHandleFunc("describe", handleDescribe)
	endpoint.AddHandleFunc("desc", handleDescribe)
	endpoint.AddHandleFunc("get", handleGet)

	// Start listening.
	return endpoint.Listen(config)
}

// NewEndpoint creates a new endpoint. To keep things simple,
// the endpoint listens on a fixed port number.
func NewEndpoint() *Endpoint {
	// Create a new Endpoint with an empty list of handler funcs.
	return &Endpoint{
		handler:       map[string]HandleFunc{},
		authenticated: false,
		// auths:   nil,
		auths: map[string]SASLServerFactory{
			sasl.Plain: func(e *Endpoint) sasl.Server {
				return sasl.NewPlainServer(func(identity, username, password string) error {
					if identity != "" && identity != username {
						log.Println("Identities not supported")
					}

					if username != "test" || password != "test" {
						return errors.New("Invalid credentials")
					}
					// if username != "test" {
					// 	return errors.New("Invalid username: " + username)
					// } else {
					// 	log.Println("username: " + username)

					// }

					// if password != "test" {
					// 	return errors.New("Invalid password: " + password)
					// } else {
					// 	log.Println("password: " + password)

					// }

					// if identity != "test" {
					// 	return errors.New("Invalid identity: " + identity)
					// } else {
					// 	log.Println("identity: " + identity)

					// }
					e.authenticated = true

					return nil
				})
			},
		},
	}
}

// AddHandleFunc adds a new function for handling incoming data.
func (e *Endpoint) AddHandleFunc(name string, f HandleFunc) {
	e.m.Lock()
	e.handler[name] = f
	e.m.Unlock()
}

// Listen starts listening on the endpoint port on all interfaces.
// At least one handler function must have been added
// through AddHandleFunc() before.
func (e *Endpoint) Listen(config config.Config) error {
	var err error
	e.listener, err = net.Listen("tcp", ":"+config.Port)
	if err != nil {
		return errors.Wrapf(err, "Unable to listen on port %s\n", config.Port)
	}
	log.Println("Listen on", e.listener.Addr().String())
	for {
		log.Println("Accept a connection request.")
		conn, err := e.listener.Accept()
		if err != nil {
			log.Println("Failed accepting a connection request:", err)
			continue
		}
		log.Println("Handle incoming messages.")
		go e.handleMessages(conn, config)
	}
}

// handleMessages reads the connection up to the first newline.
// Based on this string, it calls the appropriate HandleFunc.
func (e *Endpoint) handleMessages(conn net.Conn, config config.Config) {
	// Wrap the connection into a buffered reader for easier reading.
	rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	defer conn.Close()

	_, err := rw.WriteString("220 fnaa.unix.ar FNAA\r\n")
	if err != nil {
		log.Println("Welcome failed.", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}

	scanner := bufio.NewScanner(rw)
	scanner.Split(commons.ScanCRLF)
	if err := scanner.Err(); err != nil {
		log.Printf("Invalid input: %s", err)
	}

	for scanner.Scan() {
		cmd := strings.Split(scanner.Text(), " ")[0]
		cmd = strings.ToLower(cmd)
		log.Println("Received command: " + cmd)

		// Fetch the appropriate handler function from the 'handler' map and call it.
		e.m.RLock()
		handleCommand, ok := e.handler[cmd]
		e.m.RUnlock()
		if !ok {
			log.Println("Command '" + cmd + "' is not registered.")
			_, err := rw.WriteString("404 Unknown command\r\n")
			if err != nil {
				log.Println("Writing failed.", err)
			}
			err = rw.Flush()
			if err != nil {
				log.Println("Flush failed.", err)
			}
			// return
		} else {
			//handleCommand(rw)
			handleCommand(e, conn, rw, scanner, config)
		}

	}

	// Read from the connection until EOF. Expect a command name as the
	// next input. Call the handler that is registered for this command.
	// for {
	// 	log.Print("Receive command 1*******\n")
	// 	cmd, err := rw.ReadString('\n')
	// 	switch {
	// 	case err == io.EOF:
	// 		log.Println("Reached EOF - close this connection.\n   ---")
	// 		return
	// 	case err != nil:
	// 		log.Println("\nError reading command. Got: '"+cmd+"'\n", err)
	// 		return
	// 	}
	// 	// Trim the request string - ReadString does not strip any newlines.
	// 	cmd = strings.Trim(cmd, "\n")
	// 	log.Println(cmd + "|2******\n")

	// 	// Fetch the appropriate handler function from the 'handler' map and call it.
	// 	e.m.RLock()
	// 	handleCommand, ok := e.handler[cmd]
	// 	e.m.RUnlock()
	// 	if !ok {
	// 		log.Println("Command '" + cmd + "' is not registered.")
	// 		return
	// 	}
	// 	handleCommand(rw)
	// }
}

/* Now let's create two handler functions. The easiest case is where our
ad-hoc protocol only sends string data.

The second handler receives and processes a struct that was sent as GOB data.
*/
/*

ignatius ~ 0$echo -en "\0test\0test" | base64
AHRlc3QAdGVzdA==

*/

/*********************** HANDLERS ***********************/
func handleAuth(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {
	// Receive a string.
	// rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))

	log.Println("FULL COMMAND: " + scanner.Text())
	mech := strings.Split(scanner.Text(), " ")[1]
	mech = strings.ToLower(mech)
	if mech != "plain" {
		_, err := rw.WriteString("404 Authentication method not available\r\n")
		if err != nil {
			log.Println("Authentication method not available", err)
		}
		err = rw.Flush()
		if err != nil {
			log.Println("Authentication method not available flush failed.", err)
		}
		return
	} else {
		_, err := rw.WriteString("220 OK\r\n")
		if err != nil {
			log.Println("Authentication ready ", err)
		}
		err = rw.Flush()
		if err != nil {
			log.Println("Authentication ready flush failed.", err)
		}
	}
	// saslServer := sasl.NewPlainServer(func(identity, username, password string) error {
	// 	if identity != "" && identity != username {
	// 		log.Println("Identities not supported")
	// 	}

	// 	if username != "test" {
	// 		return errors.New("Invalid username: " + username)
	// 	} else {
	// 		log.Println("username: " + username)

	// 	}

	// 	if password != "test" {
	// 		return errors.New("Invalid password: " + password)
	// 	} else {
	// 		log.Println("password: " + password)

	// 	}

	// 	// if identity != "test" {
	// 	// 	return errors.New("Invalid identity: " + identity)
	// 	// } else {
	// 	// 	log.Println("identity: " + identity)

	// 	// }

	// 	e.authenticated = true
	// 	return nil
	// })

	scanner = bufio.NewScanner(rw)
	scanner.Split(commons.ScanCRLF)
	if err := scanner.Err(); err != nil {
		log.Printf("Invalid input: %s", err)
	}

	saslServer := e.auths[sasl.Plain](e)
	challenge, done, err := saslServer.Next(nil)
	if err != nil {
		log.Println("Error while starting server:", err)
	}
	if done {
		log.Println("Done after starting server")
	}
	if len(challenge) > 0 {
		log.Println("Invalid non-empty initial challenge:", challenge)
	}

	if scanner.Scan() {
		token := scanner.Text()
		log.Println("Received token " + token)

		data, err := base64.StdEncoding.DecodeString(token)
		if err != nil {
			_, err := rw.WriteString("404 Authentication failed\r\n")
			if err != nil {
				log.Println("Authentication failed.", err)
			}
			err = rw.Flush()
			if err != nil {
				log.Println("Authentication failed flush failed.", err)
			}
			return
		}
		// expected := []byte{105, 100, 101, 110, 116, 105, 116, 121, 0, 117, 115, 101, 114, 110, 97, 109, 101, 0, 112, 97, 115, 115, 119, 111, 114, 100}

		challenge, done, err := saslServer.Next([]byte(data))

		if err != nil {
			log.Println("Error while finishing authentication:", err)
		}
		if !done {
			log.Println("Authentication not finished after sending PLAIN credentials")
		}
		if len(challenge) > 0 {
			log.Println("Invalid non-empty final challenge:", challenge)
		}

		if !e.authenticated {
			log.Println("Not authenticated")
			_, err := rw.WriteString("404 Authentication failed\r\n")
			if err != nil {
				log.Println("Authentication failed.", err)
			}
			err = rw.Flush()
			if err != nil {
				log.Println("Authentication failed flush failed.", err)
			}

		} else {
			log.Println("User authenticated")

			_, err := rw.WriteString("220 Authenticated\r\n")
			if err != nil {
				log.Println("Authenticated failed.", err)
			}
			err = rw.Flush()
			if err != nil {
				log.Println("Authenticated flush failed.", err)
			}
		}
	}

	return
}

// handleStrings handles the "STRING" request.
func handleGet(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {
	// Receive a string.
	// rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	log.Println("FULL COMMAND: " + scanner.Text())
	log.Println(strings.Split(scanner.Text(), " "))

	response := ""
	resourceName := strings.Split(scanner.Text(), " ")[1]

	switch resourceName {
	case "namespaces", "ns":
		response += "220 DATA \n"
		response += "namespace=flow.unix.ar\n"
		response += "220 OK"

	case "flows", "fl":
		response += "220 DATA \n"
		response += "flow=time.flow.unix.ar\n"
		response += "220 OK"

	default:
		response += "404 Resource unavailable"
	}

	_, err := rw.WriteString(response + "\r\n")
	if err != nil {
		log.Println("Write DATA failed.", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}
	return
}

// handleStrings handles the "STRING" request.
func handleDescribe(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {
	// Receive a string.
	// rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	log.Println("FULL COMMAND: " + scanner.Text())
	log.Println(strings.Split(scanner.Text(), " "))
	flowName := strings.Split(scanner.Text(), " ")[2]
	log.Println("Resolving flow endpoint " + flowName)
	// Considering ksdj898

	/* DNS RESOLUTION */
	// log.Println(flowName + " IN PTR time.flow.unix.ar")
	// log.Println("time.flow.unix.ar IN PTR _fnaa._tcp.time.flow.unix.ar")
	// log.Println("queue._fnaa._tcp.time.flow.unix.ar IN SRV kf1.unix.ar")
	// log.Println("queue._fnaa._tcp.time.flow.unix.ar IN TXT type=kafka topic=ksdj898.time.flow.unix.ar")

	// response := flowName + " IN PTR time.flow.unix.ar\r\n"
	// response = response + "time.flow.unix.ar IN PTR _fnaa._tcp.time.flow.unix.ar\r\n"
	// response = response + "queue._fnaa._tcp.time.flow.unix.ar IN SRV kf1.unix.ar\r\n"
	// response = response + "queue._fnaa._tcp.time.flow.unix.ar IN TXT type=kafka topic=ksdj898.time.flow.unix.ar\r\n"
	response := "flow=" + flowName + "\n"
	response += "type=kafka\n"
	response += "topic=" + flowName + "\n"
	response += "server=kf1.unix.ar:9092\n"

	_, err := rw.WriteString("220 DATA \r\n")
	if err != nil {
		log.Println("Write DATA failed.", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}

	_, err = rw.WriteString(response + "\r\n")
	if err != nil {
		log.Println("Write DATA failed.", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}
	_, err = rw.WriteString("220 OK \r\n")
	if err != nil {
		log.Println("Write DATA failed.", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}
	return
}

// handleStrings handles the "STRING" request.
func handleSubscribe(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {
	// Receive a string.
	// rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	log.Println("FULL COMMAND: " + scanner.Text())
	log.Println(strings.Split(scanner.Text(), " "))
	flowNameSrc := strings.Split(scanner.Text(), " ")[1]

	// nameserver := flag.String("nameserver", "", "Nameserver to use")
	// user := flag.String("user", "", "Nameserver to use")
	// password := flag.String("password", "", "Nameserver to use")
	nameserver := config.Nameserver
	if nameserver == "" {
		log.Println("No Nameserver specified, use --nameserver")
		os.Exit(1)
	}

	// First, check if namespace if local
	local := false
	for _, namespace := range config.Namespaces {
		if strings.Contains(flowNameSrc, namespace.Name) {
			local = true
		}
	}

	if local {
		//Namespace is local, creating subscription
		log.Println("Creating flow endpoint " + flowNameSrc)
		log.Println("Creating new topic ksdj898." + flowNameSrc + " in Apache Kafka instance kafka_local")
		log.Println("Creating Flow Processor src=" + flowNameSrc + " dst=ksdj898." + flowNameSrc)
		log.Println("Adding DNS Records for ksdj898." + flowNameSrc)

		log.Println("Flow enabled ksdj898." + flowNameSrc)

		_, err := rw.WriteString("220 DATA\r\n")
		if err != nil {
			log.Println("Write BYE failed.", err)
		}
		err = rw.Flush()
		if err != nil {
			log.Println("Flush failed.", err)
		}

		_, err = rw.WriteString("ksdj898." + flowNameSrc + "\r\n")
		if err != nil {
			log.Println("Write BYE failed.", err)
		}
		err = rw.Flush()
		if err != nil {
			log.Println("Flush failed.", err)
		}

		_, err = rw.WriteString("220 OK\r\n")
		if err != nil {
			log.Println("Write BYE failed.", err)
		}
		err = rw.Flush()
		if err != nil {
			log.Println("Flush failed.", err)
		}
	} else {
		log.Println("Flow is REMOTE")
		flowNameDst := strings.Split(scanner.Text(), " ")[3]

		//Discover FNAA
		//Connect to FNAA
		//Authenticate with FNAA
		//Subscribe to flow
		//Create local flow
		//Launch FP
		server, result := client.ServiceResolve(flowNameSrc+".", nameserver)

		if !result {
			log.Printf("Error: Could not resolve SRV RR for FQDN %v", flowNameSrc)
			os.Exit(1)
		}

		// log.Printf("DNS Resolution result: %v:%v", server.Host, server.Port)

		log.Printf("FNAA FQDN Resolved to %v port %v", server.Host, server.Port)

		address, result := client.AddressResolve(server.Host, nameserver)
		if !result {
			log.Printf("Error: Could not resolve A RR for FQDN %v", server.Host)
			os.Exit(1)
		}
		// log.Printf("Connecting to %v port %v", address, server.Port)

		// conn, rw, err := client.Client(address, server.Port)
		Rconn, Rrw, Rerr := client.Client(address, server.Port)

		if Rerr != nil {
			log.Printf("Error: Connection to FNAA %v failed, %v", server.Host, Rerr)
			os.Exit(1)
		}

		c := *Rconn
		defer c.Close()

		log.Printf("Connected to FNAA")
		log.Printf("Authenticating with PLAIN mechanism")
		_, Rerr = client.AuthenticatePlain(Rconn, Rrw, "test", "test")

		if Rerr != nil {
			log.Printf("Error: Authentication to FNAA %v failed, %v", server.Host, Rerr)
			os.Exit(1)
		}

		log.Printf("Authenticated")

		/* EXECUTING COMMAND CREATE FLOW */
		log.Printf("Executing command SUBSCRIBE " + flowNameSrc)

		command := "SUBSCRIBE " + flowNameSrc
		response, err := client.SendCommand(Rconn, Rrw, command)

		if err != nil {
			log.Printf("Error: Create Flow %v in FNAA %v failed, %v", flowNameSrc, server.Host, err)
			os.Exit(1)
		}

		log.Printf("Flow %v subscribed successfully", flowNameSrc)
		log.Printf("Server responded: %v", response)

		log.Printf("Quitting")
		command = "QUIT"
		_, err = client.SendCommand(Rconn, Rrw, command)
		if err != nil {
			log.Printf("Error: Send command %v to FNAA %v failed, %v", command, server.Host, err)
			os.Exit(1)
		}

		c.Close()

		log.Printf("Connection closed")

		_, err = rw.WriteString("220 DATA\r\n")
		if err != nil {
			log.Println("Write BYE failed.", err)
		}
		err = rw.Flush()
		if err != nil {
			log.Println("Flush failed.", err)
		}

		_, err = rw.WriteString(flowNameDst + " SUBSCRIBED TO " + response + "\r\n")
		if err != nil {
			log.Println("Write BYE failed.", err)
		}
		err = rw.Flush()
		if err != nil {
			log.Println("Flush failed.", err)
		}

		_, err = rw.WriteString("220 OK\r\n")
		if err != nil {
			log.Println("Write BYE failed.", err)
		}
		err = rw.Flush()
		if err != nil {
			log.Println("Flush failed.", err)
		}

	}

	//Namespace is not local, creating a new remote

	return
}

// handleStrings handles the "STRING" request.
func handleCreate(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {
	// Receive a string.
	// rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	log.Println("FULL COMMAND: " + scanner.Text())
	log.Println(strings.Split(scanner.Text(), " "))
	flowName := strings.Split(scanner.Text(), " ")[2]
	log.Println("Creating flow " + flowName)
	log.Println("Creating new topic " + flowName + ".local in Apache Kafka instance kafka_local")
	log.Println("Adding DNS Records for " + flowName)
	log.Println("Flow enabled " + flowName)

	_, err := rw.WriteString("220 OK " + flowName + "\r\n")
	if err != nil {
		log.Println("Write BYE failed.", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}
	return
}

// handleStrings handles the "STRING" request.
func handleQuit(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {
	// Receive a string.
	// rw := bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn))
	log.Println("FULL COMMAND: " + scanner.Text())
	_, err := rw.WriteString("220 Bye\r\n")
	if err != nil {
		log.Println("Write BYE failed.", err)
	}
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}
	conn.Close()
	return
}

// handleStrings handles the "STRING" request.
func handleStrings(e *Endpoint, rw *bufio.ReadWriter) {
	// Receive a string.
	log.Print("Receive STRING message:")
	// log.Print("Buf message:" + string(rw.ReadLine()))
	// s, err := rw.ReadString('\n')
	// if err != nil {
	// 	log.Println("Cannot read from connection.\n", err)
	// }
	// s = strings.Trim(s, "\n ")
	// log.Println("trim:" + s)
	_, err := rw.WriteString("Thank you.\r\n")
	// if err != nil {
	// 	log.Println("Cannot write to connection.\n", err)
	// }
	err = rw.Flush()
	if err != nil {
		log.Println("Flush failed.", err)
	}
}
