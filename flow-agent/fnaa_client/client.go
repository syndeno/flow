package fnaa_client

import (
	"bufio"
	"flow-agent/fnaa"
	"log"
	"net"
	"strconv"

	"github.com/pkg/errors"
)

/*
## Outgoing connections

Using an outgoing connection is a snap. A `net.Conn` satisfies the io.Reader
and `io.Writer` interfaces, so we can treat a TCP connection just like any other
`Reader` or `Writer`.
*/

// Open connects to a TCP Address.
// It returns a TCP connection armed with a timeout and wrapped into a
// buffered ReadWriter.
func Open(addr string) (*bufio.ReadWriter, error) {
	// Dial the remote process.
	// Note that the local port is chosen on the fly. If the local port
	// must be a specific one, use DialTCP() instead.
	log.Println("Dial " + addr)
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, errors.Wrap(err, "Dialing "+addr+" failed")
	}
	return bufio.NewReadWriter(bufio.NewReader(conn), bufio.NewWriter(conn)), nil
}

// client is called if the app is called with -connect=`ip addr`.
func Client(ip string, port string) error {
	// Some test data. Note how GOB even handles maps, slices, and
	// recursive data structures without problems.
	// testStruct := complexData{
	// 	N: 23,
	// 	S: "string data",
	// 	M: map[string]int{"one": 1, "two": 2, "three": 3},
	// 	P: []byte("abc"),
	// 	C: &complexData{
	// 		N: 256,
	// 		S: "Recursive structs? Piece of cake!",
	// 		M: map[string]int{"01": 1, "10": 2, "11": 3},
	// 	},
	// }

	// Open a connection to the server.
	rw, err := Open(ip + port)
	if err != nil {
		return errors.Wrap(err, "Client: Failed to open connection to "+ip+port)
	}

	// Send a STRING request.
	// Send the request name.
	// Send the data.
	log.Println("Send the string request.")
	n, err := rw.WriteString("STRING\r\n")
	if err != nil {
		return errors.Wrap(err, "Could not send the STRING request ("+strconv.Itoa(n)+" bytes written)")
	}
	// log.Println("Send the Additional Data request.")
	// n, err = rw.WriteString("Additional Data.\r\n")
	// if err != nil {
	// 	return errors.Wrap(err, "Could not send additional STRING data ("+strconv.Itoa(n)+" bytes written)")
	// }
	log.Println("Flush the buffer.")
	err = rw.Flush()
	if err != nil {
		return errors.Wrap(err, "Flush failed.")
	}

	// Read the reply.

	log.Println("Read the reply.")

	scanner := bufio.NewScanner(rw)
	scanner.Split(fnaa.ScanCRLF)
	// Set the split function for the scanning operation.
	//scanner.Split(split)
	// Validate the input

	//response, err := rw.ReadString('\n')
	response := scanner.Text()
	// if err != nil {
	// 	return errors.Wrap(err, "Client: Failed to read the reply: '"+response+"'")
	// }

	log.Println("STRING request: got a response:", response)

	// Send a GOB request.
	// Create an encoder that directly transmits to `rw`.
	// Send the request name.
	// Send the GOB.
	// log.Println("Send a struct as GOB:")
	// log.Printf("Outer complexData struct: \n%#v\n", testStruct)
	// log.Printf("Inner complexData struct: \n%#v\n", testStruct.C)
	// enc := gob.NewEncoder(rw)
	// n, err = rw.WriteString("GOB\r\n")
	// if err != nil {
	// 	return errors.Wrap(err, "Could not write GOB data ("+strconv.Itoa(n)+" bytes written)")
	// }
	// err = enc.Encode(testStruct)
	// if err != nil {
	// 	return errors.Wrapf(err, "Encode failed for struct: %#v", testStruct)
	// }

	n, err = rw.WriteString("QUIT\r\n")
	log.Println("Read the reply.")

	scanner = bufio.NewScanner(rw)
	scanner.Split(fnaa.ScanCRLF)
	response = scanner.Text()
	log.Println("STRING request: got a response:", response)

	err = rw.Flush()
	if err != nil {
		return errors.Wrap(err, "Flush failed.")
	}

	return nil
}
