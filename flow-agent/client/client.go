package client

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"log"
	"net"
	"strconv"
	"strings"

	"github.com/miekg/dns"
	"github.com/pkg/errors"
)

type FNAAClient struct {
	conn          *net.Conn
	rw            *bufio.ReadWriter
	debug         bool
	authenticated bool
}

func Open(addr string) (*net.Conn, *bufio.ReadWriter, error) {
	// Dial the remote process.
	// Note that the local port is chosen on the fly. If the local port
	// must be a specific one, use DialTCP() instead.
	log.Println("C: Connecting to " + addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, nil, errors.Wrap(err, "Dialing "+addr+" failed")
	}
	return &conn, bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

// client is called if the app is called with -connect=`ip addr`.
// func Client(ip string, portNumber int) (*FNAAClient, error) {
func Client(ip string, portNumber int) (*net.Conn, *bufio.ReadWriter, error) {

	port := strconv.Itoa(portNumber)
	var clientFNAA FNAAClient

	// Open a connection to the server.
	conn, rw, err := Open(ip + ":" + port)
	if err != nil {
		// return nil, nil, errors.Wrap(err, "C: Failed to open connection to "+ip+port)
		return nil, nil, errors.Wrap(err, "C: Failed to open connection to "+ip+port)
	}

	clientFNAA.conn = conn
	clientFNAA.rw = rw
	clientFNAA.debug = false

	scanner := bufio.NewScanner(rw)
	scanner.Split(ScanCRLF)
	scanner.Scan()
	response := scanner.Text()
	log.Println("C: Got a response:", response)

	err = rw.Flush()
	if err != nil {
		return nil, nil, errors.Wrap(err, "Flush failed.")
	}

	return conn, rw, nil
}

func SendCommand(conn *net.Conn, rw *bufio.ReadWriter, command string) (string, error) {
	var response string
	log.Printf("C: Sending command %v", command)

	n, err := rw.WriteString(command + "\r\n")
	if err != nil {
		return "", errors.Wrap(err, "C: Could not send the command"+command)
	}
	log.Println("C: Wrote (" + strconv.Itoa(n) + " bytes written)")

	err = rw.Flush()
	if err != nil {
		return "", errors.Wrap(err, "Flush failed.")
	}

	scanner := bufio.NewScanner(rw)
	scanner.Split(ScanCRLF)
	scanner.Scan()
	line := scanner.Text()
	// log.Println("C: Got a response:", line)

	if strings.Contains(line, "220 DATA") {
		// cont := false
		scanner.Scan()
		line = scanner.Text()
		// log.Printf("C: Got a response:\n%v", line)
		for !strings.Contains(line, "220 OK") {
			// log.Println("New line")

			// log.Printf("C: Server sent data line:\n%v", line)
			response = response + line
			scanner.Scan()
			line = scanner.Text()
			// log.Println("Finish line")
		}
	}

	if strings.Contains(line, "220 OK") {
		log.Printf("C: Server sent OK for command %v", command)
	}

	// if strings.Contains(line, "220 DATA") {
	// 	cont := false
	// 	for !cont {
	// 		// log.Println("New line")
	// 		line = scanner.Text()
	// 		if !strings.Contains(line, "220 OK") {
	// 			log.Printf("C: Server sent data line: %v", line)
	// 			scanner.Scan()

	// 		} else if strings.Contains(line, "220 OK") {
	// 			log.Printf("C: Server finished sending data with 220 OK for command %v", command)
	// 			cont = true
	// 		} else {
	// 			log.Printf("C: Server sent data line: %v", line)
	// 			response = response + line
	// 			scanner.Scan()
	// 		}
	// 		// log.Println("Finish line")
	// 	}
	// } else if strings.Contains(line, "220 OK") {
	// 	log.Printf("C: Server sent OK for command %v", command)

	// }

	err = rw.Flush()
	if err != nil {
		return "", errors.Wrap(err, "Flush failed.")
	}
	// log.Printf("C: Server response\n%v", response)

	return response, nil
}

func AuthenticatePlain(conn *net.Conn, rw *bufio.ReadWriter, username string, password string) (string, error) {
	command := "AUTHENTICATE PLAIN"
	log.Printf("C: Sending command " + command)

	/* REQUEST AUTH */
	n, err := rw.WriteString(command + "\r\n")
	if err != nil {
		return "", errors.Wrap(err, "C: Could not send the command"+command)
	}
	log.Println("C: Wrote (" + strconv.Itoa(n) + " bytes written)")

	err = rw.Flush()
	if err != nil {
		return "", errors.Wrap(err, "Flush failed.")
	}

	scanner := bufio.NewScanner(rw)
	scanner.Split(ScanCRLF)
	scanner.Scan()
	response := scanner.Text()
	log.Println("C: Got a response:", response)

	err = rw.Flush()
	if err != nil {
		return "", errors.Wrap(err, "Flush failed.")
	}

	/* SEND CHALLENGE */
	data := "\x00" + username + "\x00" + password
	sEnc := base64.StdEncoding.EncodeToString([]byte(data))
	log.Println("C: Authentication string sent: " + sEnc)
	// fmt.Println(sEnc)
	n, err = rw.WriteString(sEnc + "\r\n")
	if err != nil {
		return "", errors.Wrap(err, "C: Could not send the command"+command)
	}
	log.Println("C: Wrote (" + strconv.Itoa(n) + " bytes written)")

	err = rw.Flush()
	if err != nil {
		return "", errors.Wrap(err, "Flush failed.")
	}

	scanner = bufio.NewScanner(rw)
	scanner.Split(ScanCRLF)
	scanner.Scan()
	response = scanner.Text()
	log.Println("C: Got a response:", response)
	if !strings.Contains(response, "220") {
		// log.Println("Did not authenticate")
		return "", errors.New("Authentication failed")

	}
	err = rw.Flush()
	if err != nil {
		return "", errors.Wrap(err, "Flush failed.")
	}

	return response, nil
}

func dropCR(data []byte) []byte {
	if len(data) > 0 && data[len(data)-1] == '\r' {
		return data[0 : len(data)-1]
	}
	return data
}

func ScanCRLF(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, []byte{'\r', '\n'}); i >= 0 {
		// We have a full newline-terminated line.
		return i + 2, dropCR(data[0:i]), nil
	}
	// If we're at EOF, we have a final, non-terminated line. Return it.
	if atEOF {
		return len(data), dropCR(data), nil
	}
	// Request more data.
	return 0, nil, nil
}

type fnaaServer struct {
	Host string
	Port int
}

func AddressResolve(fqdn string, nameserver string) (string, bool) {
	log.Println("**Starting Address resolution")

	var (
		// qtype  []uint16
		// qclass []uint16
		qname []string
	)

	qname = append(qname, fqdn)

	// if len(nameserver) == 0 {
	// 	conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
	// 	if err != nil {
	// 		fmt.Fprintln(os.Stderr, err)
	// 		os.Exit(2)
	// 	}
	// 	nameserver = "@" + conf.Servers[0]
	// }
	// nameserver = string([]byte(nameserver)[1:]) // chop off @

	port := 53
	log.Printf("**Nameserver to be used: %v", nameserver)

	if nameserver[0] == '[' && nameserver[len(nameserver)-1] == ']' {
		nameserver = nameserver[1 : len(nameserver)-1]
	}
	if i := net.ParseIP(nameserver); i != nil {
		nameserver = net.JoinHostPort(nameserver, strconv.Itoa(port))
	} else {
		nameserver = dns.Fqdn(nameserver) + ":" + strconv.Itoa(port)
	}

	log.Printf("**Resolving A for %v using server %v", fqdn, nameserver)

	var address string
	answer := ExecuteQuery(nameserver, dns.TypeA, qname[0])
	for _, a := range answer {
		if ar, ok := a.(*dns.A); ok {
			address = string(ar.A.String())

		} else {
			return address, false
		}
	}

	log.Printf("**Resolved A to %v for %v using server %v", address, fqdn, nameserver)

	return address, true
}

func ExecuteQuery(nameserver string, qtype uint16, qname string) []dns.RR {
	log.Printf("***Executing query %v IN %v using server %v", qname, qtype, nameserver)

	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(qname, qtype)

	m.RecursionDesired = true
	r, _, err := c.Exchange(m, nameserver)
	if err != nil {
		log.Printf("***Contacting nameserver resulted in error: %v", err)
		return nil
	}
	if r.Rcode != dns.RcodeSuccess {
		log.Println("***Executing query did not return Rcode Success")
		return nil
	}

	log.Printf("***Executing successful: %v", r.Answer)

	return r.Answer
}

func ServiceResolve(fqdn string, nameserver string) (fnaaServer, bool) {
	log.Println("**Starting FQDN resolution with " + nameserver)

	var (
		// qtype  []uint16
		// qclass []uint16
		qname []string
	)

	qname = append(qname, fqdn)

	port := 53
	log.Printf("**Nameserver to be used: %v", nameserver)

	if nameserver[0] == '[' && nameserver[len(nameserver)-1] == ']' {
		nameserver = nameserver[1 : len(nameserver)-1]
		log.Printf("[] %v", nameserver)
	}
	if i := net.ParseIP(nameserver); i != nil {
		nameserver = net.JoinHostPort(nameserver, strconv.Itoa(port))
		log.Printf("IP %v", nameserver)
	} else {
		nameserver = dns.Fqdn(nameserver) + ":" + strconv.Itoa(port)
		log.Printf("Else %v", nameserver)
	}

	server := fnaaServer{}
	// server.Host = "time.flows.unix.ar"
	// server.Port = 1
	log.Printf("**#Resolving SRV for %v using server %v", fqdn, nameserver)

	answer := ExecuteQuery(nameserver, dns.TypeSRV, qname[0])
	for _, a := range answer {
		if srv, ok := a.(*dns.SRV); ok {
			server.Port = int(srv.Port)
			server.Host = string(srv.Target)

		}
	}

	if server.Host == "" || server.Port == 0 {
		log.Printf("**Error Resolving SRV for %v using server %v", "fnaa._flow._tcp."+fqdn, nameserver)
		answer = ExecuteQuery(nameserver, dns.TypeSRV, "fnaa._flow._tcp."+qname[0])
		for _, a := range answer {
			if srv, ok := a.(*dns.SRV); ok {
				server.Port = int(srv.Port)
				server.Host = string(srv.Target)

			} else {
				return server, false
			}
		}
	}

	return server, true
}
