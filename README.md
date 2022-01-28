# Flow Proof of Concept
This codebase includes an implementation for the FNAA and FNUA components of the Event Streaming Open Network.

This is based on <https://github.com/syndeno/draft-spinella-event-streaming-open-network>

# Introduction
We will focus on providing a minimum implementation of the main Event Streaming Open Network component: the Flow Namespace Accessing Agent. This implementation should serve as a Proof of Concept of the overall Event Streaming Open Network proposal.

As described in the previous section, the Flow Namespace Accessing Agent (FNAA) is the main and core required component for the Open Network. All Network Participants must deploy an FNAA server instance in order to be part of the network. The FNAA actually implements a server-like application for the Flow Namespace Accessing Protocol (FNAP). Then, the first objective of this Proof of Concept is to show an initial implementation of the FNAA server component.

On the other hand, the FNAA is accessed by means of a Flow Namespace User Agent (FNUA). This component acts as a client application that connects to a FNAA. Also, this component can take different forms: it could be a web-based application, a desktop application or even a command line tool. For the purposes of this Proof of Concept, we will implement a CLI tool that acts as a client application for the FNAA. Thus, the second objective of this PoC is to provide an initial implementation of the FNUA client component.

In the following sections, we will first describe the minimum functionalities considered for validating the overall proposal for the Event Streaming Open Network. This minimum set of requirements for both the FNAA and the FNUA will compose the Proof of Concept.

Afterwards, we will describe the technology chosen for the initial implementation of both the FNAA and the FNUA. Then, a description of how these tools work in isolation will be provided. Subsequently, we will review different use cases to prove how the network could be used by network participants and its users.

Lastly, we will provide a conclusion for this Proof of Concept, where we mentioning if and how the minimum established requirements have been met or not.

## Minimum functionalities 
Network Participants system administrators must be able to run a Server Application that acts as FNAA.

Users using a Client Application actiong as a FNUA must be able to:
1. Access the flow account and operate its flows.
2. Create a new flow.
3. Describe an existing flow.
4. Subscribe to an external flow.

## FNAA - Server application
The FNAA server application must implement FNAP as described in Section 6. Basically, the FNAA will open a TCP port on all the IP addresses of the host to listen for new FNUA client connections.

The chosen language for the development of the FNAA is GoLang. The reason for choosing GoLang is because Kubernetes is written in this language and there is a robust set of libraries available for integration. Although there is no integration built with Kubernetes for this Proof of Concept, the usage of GoLang will enable a seamless evolution of the FNAA application. In future versions of the FNAA codebase, new functionalities leveraging Kubernetes will be easier to implement than if using a different programming language.

When the FNAA server application is initialized, it provides debug log messages describing all client interactions. In order to start the server application, a Network Participant system administrator can download the binary and execute it in a terminal:

	ignatius ~ 0$./fnaad 
	server.go:146: Listen on [::]:61000
	server.go:148: Accept a connection request.

Now that the 61000 TCP port is open, we can test the behaviour by means of a raw TCP using telnet command in a different terminal:

	ignatius ~ 1$telnet localhost 61000
	Trying 127.0.0.1...
	Connected to localhost.
	Escape character is '^]'.
	220 fnaa.unix.ar FNAA

We can now see that the server has provided the first message in the connection: a welcome message indicating its FQDN fnaa.unix.ar.

On the other hand, the server application starts providing debug information for the new connection established:

	ignatius ~ 0$./fnaad 
	server.go:146: Listen on [::]:61000
	server.go:148: Accept a connection request.
	server.go:154: Handle incoming messages.
	server.go:148: Accept a connection request.

## FNUA - Client application
In order to test the FNAA server application, a CLI-based FNUA application has been developed. The chosen language for this CLI tool is also GoLang. The reason for choosing GoLang for the FNUA is because of its functionalities for building CLI tools, leveraging on the Cobra library.
Thus, the FNUA for the PoC is an executable file that complies with the diagram in Figure 14.

One of the requirements for the flow CLI tool is a configuration file that defines the different FNAA servers together with the credentials to use. An example of this configuration file follows:

	agents:
	  -
	    name: fnaa-unix
	    fqdn: fnaa.unix.ar
	    username: test
	    password: test
	    prefix: unix.ar-
	  -
	    name: fnaa-emiliano
	    fqdn: fnaa.emiliano.ar
	    username: test
	    password: test
	    prefix: emiliano.ar-

	namespaces:
	  -
	    name: flows.unix.ar
	    agent: fnaa-unix
	  -
	    name: flows.emiliano.ar
	    agent: fnaa-emiliano


In this file, we can see that there are two FNAA instances described with FQDN fnaa.unix.ar and fnaa.emiliano.ar. Then, there are two namespaces: one called flow.unix.ar hosted on fnaa-unix and second namespace flows.emiliano.ar hosted on fnaa-emiliano. This configuration enables the FNUA to interact with two different FNAA, each of which is hosting different Flow Namespaces.

Once the configuration file has been saved, the flow CLI tool can now be used. In the following sections, we will show how to use the minimum functionalities required for the Open Network using this CLI tool.


## Use cases 
### Use case 1: Authenticating a user
After the connection is established, the first command that the client must execute is the authentication command. As previously defined in Chapter 5, every FNAA client must first authenticate in order to execute commands. Thus, the authentication challenge must be supported both by the FNAA as well as the FNUA. 

It is worth mentioning that the chosen authentication mechanism for this PoC is SASL Plain. This command can be extended furtherly with other mechanisms in later versions. However, this simple authentication mechanism is sufficient to demonstrate the authentication step in the FNAP.

The SASL Plain Authentication implies sending the username and the password encoded in Base64. The way to obtain the Base64 if we consider a user test with password test, is as follows:
ignatius ~ 0$echo -en "\0test\0test" | base64
AHRlc3QAdGVzdA==

Now, we can use this Base64 string to authenticate with the FNAA. First, we need to launch the FNAA server instance:

	ignatius~/ $./fnaad --config ./fnaad_flow.unix.ar.yaml
	main.go:41: Using config file: ./fnaad_flow.unix.ar.yaml
	main.go:57:     Using config file: ./fnaad_flow.unix.ar.yaml
	server.go:103: Listen on [::]:61000
	server.go:105: Accept a connection request.

Then, we can connect to the TCP port in which the FNAA is listening:

	ignatius ~ 1$telnet localhost 61000
	Trying 127.0.0.1...
	Connected to localhost.
	Escape character is '^]'.
	220 fnaa.unix.ar FNAA
	AUTHENTICATE PLAIN
	220 OK
	AHRlc3QAdGVzdA==
	220 Authenticated

Once the client is authenticated, it can start executing FNAP commands to manage the Flow Namespace of the authenticated user. For simplicity purposes, in this Proof of Concept, we will be using a single user.

In the case of the CLI tool, there is no need to perform an authentication step, since every command the user executes will be preceded by an authentication in the server.

### Use case 2: Creating a flow
Once the authentication is successful, the client can now create a new Flow.  The way to do this using the CLI tool would be:

	ignatius ~/ 0$./fnua create flow time.flow.unix.ar
	Resolving SRV for fnaa.unix.ar. using server 172.17.0.2:53
	Executing query fnaa.unix.ar. IN 33 using server 172.17.0.2:53
	Executing successful: [fnaa.unix.ar.	604800	IN	SRV	0 0 61000 fnaa.unix.ar.]
	Resolving A for fnaa.unix.ar. using server 172.17.0.2:53
	Executing query fnaa.unix.ar. IN 1 using server 172.17.0.2:53
	Executing successful: [fnaa.unix.ar.	604800	IN	A	127.0.0.1]
	Resolved A to 127.0.0.1 for fnaa.unix.ar. using server 172.17.0.2:53
	C: Connecting to 127.0.0.1:61000
	C: Got a response: 220 fnaa.unix.ar FNAA
	C: Sending command AUTHENTICATE PLAIN
	C: Wrote (20 bytes written)
	C: Got a response: 220 OK
	C: Authentication string sent: AHRlc3QAdGVzdA==
	C: Wrote (18 bytes written)
	C: Got a response: 220 Authenticated
	C: Sending command CREATE FLOW time.flow.unix.ar
	C: Wrote (31 bytes written)
	C: Server sent OK for command CREATE FLOW time.flow.unix.ar
	C: Sending command QUIT
	C: Wrote (6 bytes written)

The client has discovered the FNAA server for Flow Namespace flow.unix.ar by means of SRV DNS records. Thus, it obtained both the FQDN of the FNAA together with the TCP port where it is listening, in this case 61000. Once the resolution process ends, the FNUA connects to the FNAA. First, the FNUA authenticates with the FNAA and then it executes the create flow command.

If we were to simulate the same behavior using a raw TCP connection, the following steps would be executed:

	ignatius ~ 1$telnet localhost 61000
	Trying 127.0.0.1...
	Connected to localhost.
	Escape character is '^]'.
	220 fnaa.unix.ar FNAA
	AUTHENTICATE PLAIN
	220 OK
	AHRlc3QAdGVzdA==
	220 Authenticated
	CREATE FLOW time.flows.unix.ar
	220 OK time.flows.unix.ar

Now, the client has created a new flow called time.flows.unix.ar located in the flows.unix.ar namespace. The FNAA in background has created a Kafka Topic as well as the necessary DNS entries for name resolution.

### Use case 3: Describing a flow
Once a flow has been created, we can obtain information of if by executing the following command using the CLI tool:

	ignatius ~/ 1$./fnua describe flow time.flow.unix.ar
	Resolving SRV for fnaa.unix.ar. using server 172.17.0.2:53
	Executing query fnaa.unix.ar. IN 33 using server 172.17.0.2:53
	Executing successful: [fnaa.unix.ar.	604800	IN	SRV	0 0 61000 fnaa.unix.ar.]
	Nameserver to be used: 172.17.0.2
	Resolving A for fnaa.unix.ar. using server 172.17.0.2:53
	Executing query fnaa.unix.ar. IN 1 using server 172.17.0.2:53
	Executing successful: [fnaa.unix.ar.	604800	IN	A	127.0.0.1]
	Resolved A to 127.0.0.1 for fnaa.unix.ar. using server 172.17.0.2:53
	C: Connecting to 127.0.0.1:61000
	C: Got a response: 220 fnaa.unix.ar FNAA
	C: Sending command AUTHENTICATE PLAIN
	C: Wrote (20 bytes written)
	C: Got a response: 220 OK
	C: Authentication string sent: AHRlc3QAdGVzdA==
	C: Wrote (18 bytes written)
	C: Got a response: 220 Authenticated
	C: Sending command DESCRIBE FLOW time.flow.unix.ar
	C: Wrote (33 bytes written)
	C: Server sent OK for command DESCRIBE FLOW time.flow.unix.ar
	Flow time.flow.unix.ar description:
	flow=time.flow.unix.ar
	type=kafka
	topic=time.flow.unix.ar
	server=kf1.unix.ar:9092
	Flow time.flow.unix.ar described successfully
	Quitting
	C: Sending command QUIT
	C: Wrote (6 bytes written)

In the output of the describe command we can see all the necessary information to connect to the Flow called time.flow.unix.ar: (i) the type of Event Broker is Kafka, (ii) the Kafka topic has the same name of the flow and (iii) the Kafka Bootstrap server with port is provided. If we were to obtain this information using a manual connection, the steps would be:

	ignatius ~ 1$telnet localhost 61000
	Trying 127.0.0.1...
	Connected to localhost.
	Escape character is '^]'.
	220 fnaa.unix.ar FNAA
	AUTHENTICATE PLAIN
	220 OK
	AHRlc3QAdGVzdA==
	220 Authenticated
	DESCRIBE FLOW time.flows.unix.ar
	220 DATA 
	flow=time.flows.unix.ar
	type=kafka
	topic=time.flows.unix.ar
	server=kf1.unix.ar:9092
	220 OK 

Now, we can use this information to connect to the Kafka topic and start producing or consuming events.

### Use case 4: Subscribing to a remote flow
In this section, we will show how a subscription can be set up. When a user commands the FNAA to create a new subscription to a remote Flow, the local FNAA server first needs to discover the remote FNAA server. Once the server is discovered by means of DNS resolution, the local FNAA contacts the remote FNAA, authenticates the user and then executes a subscription command.

Thus, the initial communication between the FNUA and the FNAA, in which the user indicates to subscribe to a remote flow, would be as follows:

	ignatius ~ 1$telnet localhost 61000
	Trying 127.0.0.1...
	Connected to localhost.
	Escape character is '^]'.
	220 fnaa.unix.ar FNAA
	AUTHENTICATE PLAIN
	220 OK
	AHRlc3QAdGVzdA==
	220 Authenticated
	SUBSCRIBE time.flows.unix.ar LOCAL emiliano.ar-time.flows.unix.ar
	220 DATA
	ksdj898.time.flows.unix.ar
	220 OK

Once the user is authenticated, a SUBSCRIBE command is executed. This command indicates first the remote flow to subscribe to. Then, it also specifies with LOCAL the flow where the remote events will be written. In this example, the remote flow to subscribe to is time.flows.unix.ar, and the local flow is emiliano.ar-time.flows.unix.ar. Basically, a new flow has been created, emiliano.ar-time.flows.unix.ar, where all the events of flow time.flows.unix.ar will be written. 

The server answers back with a new Flow URI, in this case ksdj898.time.flows.unix.ar. This Flow URI indicates a copy of the original flow time.flows.unix.ar created for this subscription. Thus, the remote FNAA has full control over this subscription, being able to revoke it by simply deleting this flow or applying Quality of Service rules.

The remote FNAA has set up a Bridge Processor to transcribe messages in topic time.flows.unix.ar to the new topic ksdj898.time.flows.unix.ar. Another alternative to a Bridge Processor would be a Distributor Processor, which could be optimized for a Flow with high demand. Moreover, instead of creating a single Bridge Processor per subscription, a Distributor Processor could be used, in order to have a single consumer of the source flow and write the events to several subscription flows.

The user could use the FNUA CLI tool to execute this command in the following manner:

	ignatius ~ 0$./fnua --config=./flow.yml subscribe time.flows.unix.ar --nameserver 172.17.0.2 -d --agent fnaa-emiliano
	Initializing initConfig
	    Using config file: ./flow.yml
	Subscribe to flow
	Agent selected: fnaa-emiliano
	Resolving FNAA FQDN fnaa.emiliano.ar
	Starting FQDN resolution with 172.17.0.2
	Resolving SRV for fnaa.emiliano.ar. using server 172.17.0.2:53
	Executing query fnaa.emiliano.ar. IN 33 using server 172.17.0.2:53
	FNAA FQDN Resolved to fnaa.emiliano.ar. port 51000
	Resolving A for fnaa.emiliano.ar. using server 172.17.0.2:53
	Resolved A to 127.0.0.1 for fnaa.emiliano.ar. using server 172.17.0.2:53
	C: Connecting to 127.0.0.1:51000
	C: Got a response: 220 fnaa.unix.ar FNAA
	Connected to FNAA
	Authenticating with PLAIN mechanism
	C: Sending command AUTHENTICATE PLAIN
	C: Wrote (20 bytes written)
	C: Got a response: 220 OK
	C: Authentication string sent: AHRlc3QAdGVzdA==
	C: Wrote (18 bytes written)
	C: Got a response: 220 Authenticated
	Authenticated
	Executing command SUBSCRIBE time.flows.unix.ar LOCAL emiliano.ar-time.flows.unix.ar
	C: Sending command SUBSCRIBE time.flows.unix.ar LOCAL emiliano.ar-time.flows.unix.ar
	C: Wrote (67 bytes written)
	C: Server sent OK for command SUBSCRIBE time.flows.unix.ar LOCAL emiliano.ar-time.flows.unix.ar
	Flow emiliano.ar-time.flows.unix.ar subscription created successfully
	Server responded: emiliano.ar-time.flows.unix.ar SUBSCRIBED TO ksdj898.time.flows.unix.ar
	Quitting
	C: Sending command QUIT
	C: Wrote (6 bytes written)
	Connection closed

This interaction of the FNUA with the FNAA of the Flow Namespace emiliano.ar (fnaa-emiliano) has trigger an interaction with the FNAA of unix.ar Flow Namespace (fnaa-unix). This means that before fnaa-emiliano was able to respond to the FNUA, a new connection was opened to the remote FNAA and the SUBSCRIBE command was executed.

The log of fnaa-emiliano when the SUBCRIBE command was issued looks as follows:

	server.go:111: Handle incoming messages.
	server.go:105: Accept a connection request.
	server.go:253: User authenticated
	server.go:347: FULL COMMAND: SUBSCRIBE time.flows.unix.ar LOCAL emiliano.ar-time.flows.unix.ar
	server.go:401: Flow is REMOTE
	client.go:280: **#Resolving SRV for time.flows.unix.ar. using server 172.17.0.2:53
	server.go:417: FNAA FQDN Resolved to fnaa.unix.ar. port 61000
	client.go:42: C: Connecting to 127.0.0.1:61000
	client.go:69: C: Got a response: 220 fnaa.unix.ar FNAA
	server.go:435: Connected to FNAA
	server.go:436: Authenticating with PLAIN mechanism
	client.go:126: C: Sending command AUTHENTICATE PLAIN
	client.go:133: C: Wrote (20 bytes written)
	client.go:144: C: Got a response: 220 OK
	client.go:154: C: Authentication string sent: AHRlc3QAdGVzdA==
	client.go:159: C: Wrote (18 bytes written)
	client.go:170: C: Got a response: 220 Authenticated
	server.go:444: Authenticated
	client.go:82: C: Sending command SUBSCRIBE time.flows.unix.ar
	client.go:88: C: Wrote (30 bytes written)
	client.go:112: C: Server sent OK for command SUBSCRIBE time.flows.unix.ar
	server.go:456: Flow time.flows.unix.ar subscribed successfully
	server.go:457: Server responded: ksdj898.time.flows.unix.ar
	server.go:459: Quitting

We can see how fnaa-emiliano had to trigger a client subroutine to contact the remote fnaa-unix. Once the server FQDN, IP and Port is discovered by means of DNS, a new connection is established and the SUBSCRIBE command is issued. Here we can see the log of fnaa-unix:

	server.go:111: Handle incoming messages.
	server.go:105: Accept a connection request.
	server.go:253: User authenticated
	server.go:139: Received command: subscribe
	server.go:348: [SUBSCRIBE time.flows.unix.ar]
	server.go:367: Creating flow endpoint time.flows.unix.ar
	server.go:368: Creating new topic ksdj898.time.flows.unix.ar in Apache Kafka instance kafka_local
	server.go:369: Creating Flow Processor src=time.flows.unix.ar dst=ksdj898.time.flows.unix.ar
	server.go:370: Adding DNS Records for ksdj898.time.flows.unix.ar
	server.go:372: Flow enabled ksdj898.time.flows.unix.ar
	server.go:139: Received command: quit

Thus, we were able to set up a new subscription in fnaa-emiliano that trigger a background interaction with fnaa-unix.

## Results of the PoC
We can confirm the feasibility of the overall Event Streaming Open Network architecture. The test of the proposed protocol FNAP and its implementation, both in the FNAA and FNUA (CLI application), show that the architecture can be employed for the purpose of distributed subscription management among Network Participants.

The minimum functionalities defined both for the Network Participants and the Users were met. Network Participants can run this type of service by means of a server application, the FNAA server. Also, the CLI-tool resulted in a convenient low-level method to interact with a FNAA server.

In further implementations, the server application should be optimized as well as secured, for instance with a TLS handshake. Also, the CLI-tool could be enhanced by a web-based application with a friendly user interface.

Nevertheless, the challenge for a stable implementation of both components is the possibility of supporting different Event Brokers and their evolution. Not only Apache Kafka should be supported but also the main Public Cloud providers events solutions, such as AWS SQS or Google Cloud Pub/Sub. Since the Event Brokers are continuously evolving, the implementation of the FNAA component should keep up both with the API and new functionalities of these vendors. 

Regarding the protocol design, it would be needed to enhance the serialization of the exchanged data. In this sense, it could be convenient to define a packet header for the overall interaction between the FNAA both with remote FNAA as well as with FNUA.

Regarding the subscription use case, it would be necessary to establish a convenient format for the server response. Currently, the server is returning a key/value structure with the details of the Flow. This structure may not be the most adequate, since it may differ depending on the Event Broker used.

Also, the security aspect needs further analysis and design since its fragility could lead to great economical damage for organizations. Thus, it would be recommended to review the different security controls needed for this solution as part of an Information Security Management System.

Finally, the implementation should leverage the Cloud Native functionalities provided by the Kubernetes API. For example, the FNAA should trigger the deployment of Flow Processors on demand, in order to provide isolated computing resources for each subscription. Also, a Kubernetes resource could be developed to use the kubectl CLI tool for management, instead of a custom CLI tool.
	
