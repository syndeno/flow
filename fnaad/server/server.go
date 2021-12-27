/*
Copyright © 2021 Emiliano Spinella emilianofs@gmail.com

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package server

import (
	"bufio"
	"encoding/base64"
	"fnaad/client"
	"fnaad/commons"
	"fnaad/config"

	"log"
	"net"
	"os"
	"strings"
	"sync"

	"github.com/emersion/go-sasl"
	"github.com/pkg/errors"
)

//Structures used by the FNAA
type SASLServerFactory func(*Endpoint) sasl.Server
type HandleFunc func(*Endpoint, net.Conn, *bufio.ReadWriter, *bufio.Scanner, config.Config)
type Endpoint struct {
	listener      net.Listener
	handler       map[string]HandleFunc
	auths         map[string]SASLServerFactory
	connection    net.Conn
	rw            *bufio.ReadWriter
	authenticated bool
	m             sync.RWMutex
}

//Server creation function based on configuration
//In this function the commands can be added
func Server(config config.Config) error {
	endpoint := NewEndpoint()

	endpoint.AddHandleFunc("quit", handleQuit)
	endpoint.AddHandleFunc("authenticate", handleAuth)
	endpoint.AddHandleFunc("create", handleCreate)
	endpoint.AddHandleFunc("subscribe", handleSubscribe)
	endpoint.AddHandleFunc("describe", handleDescribe)
	endpoint.AddHandleFunc("desc", handleDescribe)
	endpoint.AddHandleFunc("get", handleGet)

	return endpoint.Listen(config)
}

//Function that manages each individual client connection
func NewEndpoint() *Endpoint {
	return &Endpoint{
		handler:       map[string]HandleFunc{},
		authenticated: false,
		auths: map[string]SASLServerFactory{
			sasl.Plain: func(e *Endpoint) sasl.Server {
				return sasl.NewPlainServer(func(identity, username, password string) error {
					if identity != "" && identity != username {
						log.Println("Identities not supported")
					}

					if username != "test" || password != "test" {
						return errors.New("Invalid credentials")
					}

					e.authenticated = true

					return nil
				})
			},
		},
	}
}

//Function that manages command handlers
func (e *Endpoint) AddHandleFunc(name string, f HandleFunc) {
	e.m.Lock()
	e.handler[name] = f
	e.m.Unlock()
}

//Function that establishes conections with clients
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

//Function that handles client messages
func (e *Endpoint) handleMessages(conn net.Conn, config config.Config) {
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
		} else {
			handleCommand(e, conn, rw, scanner, config)
		}

	}

}

/*********************** COMMAND HANDLERS ***********************/

//Function that handles AUTHENTICATION command
func handleAuth(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {
	/*
		ignatius ~ 0$echo -en "\0test\0test" | base64
		AHRlc3QAdGVzdA==
	*/
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

//Function that handles GET command
func handleGet(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {

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

//Function that handles DESCRIBE command
func handleDescribe(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {

	log.Println("FULL COMMAND: " + scanner.Text())
	log.Println(strings.Split(scanner.Text(), " "))
	flowName := strings.Split(scanner.Text(), " ")[2]
	log.Println("Resolving flow endpoint " + flowName)

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

//Function that handles SUBSCRIBE command
func handleSubscribe(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {
	log.Println("FULL COMMAND: " + scanner.Text())
	log.Println(strings.Split(scanner.Text(), " "))
	flowNameSrc := strings.Split(scanner.Text(), " ")[1]

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

		log.Printf("FNAA FQDN Resolved to %v port %v", server.Host, server.Port)

		address, result := client.AddressResolve(server.Host, nameserver)
		if !result {
			log.Printf("Error: Could not resolve A RR for FQDN %v", server.Host)
			os.Exit(1)
		}

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

	return
}

//Function that handles CREATE command
func handleCreate(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {

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

//Function that handles QUIT command
func handleQuit(e *Endpoint, conn net.Conn, rw *bufio.ReadWriter, scanner *bufio.Scanner, config config.Config) {

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
