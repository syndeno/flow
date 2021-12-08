package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/miekg/dns"
)

var (
	dnskey       *dns.DNSKEY
	short        = flag.Bool("short", false, "abbreviate long DNSSEC records")
	dnssec       = flag.Bool("dnssec", false, "request DNSSEC records")
	query        = flag.Bool("question", false, "show question")
	check        = flag.Bool("check", false, "check internal DNSSEC consistency")
	six          = flag.Bool("6", false, "use IPv6 only")
	four         = flag.Bool("4", false, "use IPv4 only")
	anchor       = flag.String("anchor", "", "use the DNSKEY in this file as trust anchor")
	tsig         = flag.String("tsig", "", "request tsig with key: [hmac:]name:key")
	port         = flag.Int("port", 53, "port number to use")
	laddr        = flag.String("laddr", "", "local address to use")
	aa           = flag.Bool("aa", false, "set AA flag in query")
	ad           = flag.Bool("ad", false, "set AD flag in query")
	cd           = flag.Bool("cd", false, "set CD flag in query")
	rd           = flag.Bool("rd", true, "set RD flag in query")
	fallback     = flag.Bool("fallback", false, "fallback to 4096 bytes bufsize and after that TCP")
	tcp          = flag.Bool("tcp", false, "TCP mode, multiple queries are asked over the same connection")
	timeoutDial  = flag.Duration("timeout-dial", 2*time.Second, "Dial timeout")
	timeoutRead  = flag.Duration("timeout-read", 2*time.Second, "Read timeout")
	timeoutWrite = flag.Duration("timeout-write", 2*time.Second, "Write timeout")
	nsid         = flag.Bool("nsid", false, "set edns nsid option")
	client       = flag.String("client", "", "set edns client-subnet option")
	opcode       = flag.String("opcode", "query", "set opcode to query|update|notify")
	rcode        = flag.String("rcode", "success", "set rcode to noerror|formerr|nxdomain|servfail|...")
)

func main() {
	var (
		qtype  []uint16
		qclass []uint16
		qname  []string
	)

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] [@server] [qtype...] [qclass...] [name ...]\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()
	if *anchor != "" {
		f, err := os.Open(*anchor)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failure to open %s: %s\n", *anchor, err.Error())
		}
		r, err := dns.ReadRR(f, *anchor)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failure to read an RR from %s: %s\n", *anchor, err.Error())
		}
		if k, ok := r.(*dns.DNSKEY); !ok {
			fmt.Fprintf(os.Stderr, "No DNSKEY read from %s\n", *anchor)
		} else {
			dnskey = k
		}
	}

	var nameserver string
	for _, arg := range flag.Args() {
		// If it starts with @ it is a nameserver
		if arg[0] == '@' {
			nameserver = arg
			continue
		}
		// First class, then type, to make ANY queries possible
		// And if it looks like type, it is a type
		if k, ok := dns.StringToType[strings.ToUpper(arg)]; ok {
			qtype = append(qtype, k)
			continue
		}
		// If it looks like a class, it is a class
		if k, ok := dns.StringToClass[strings.ToUpper(arg)]; ok {
			qclass = append(qclass, k)
			continue
		}
		// If it starts with TYPExxx it is unknown rr
		if strings.HasPrefix(arg, "TYPE") {
			i, err := strconv.Atoi(arg[4:])
			if err == nil {
				qtype = append(qtype, uint16(i))
				continue
			}
		}
		// If it starts with CLASSxxx it is unknown class
		if strings.HasPrefix(arg, "CLASS") {
			i, err := strconv.Atoi(arg[5:])
			if err == nil {
				qclass = append(qclass, uint16(i))
				continue
			}
		}
		// Anything else is a qname
		qname = append(qname, arg)
	}
	if len(qname) == 0 {
		qname = []string{"."}
		if len(qtype) == 0 {
			qtype = append(qtype, dns.TypeNS)
		}
	}
	if len(qtype) == 0 {
		qtype = append(qtype, dns.TypeA)
	}
	if len(qclass) == 0 {
		qclass = append(qclass, dns.ClassINET)
	}

	if len(nameserver) == 0 {
		conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		nameserver = "@" + conf.Servers[0]
	}

	nameserver = string([]byte(nameserver)[1:]) // chop off @

	if nameserver[0] == '[' && nameserver[len(nameserver)-1] == ']' {
		nameserver = nameserver[1 : len(nameserver)-1]
	}
	if i := net.ParseIP(nameserver); i != nil {
		nameserver = net.JoinHostPort(nameserver, strconv.Itoa(*port))
	} else {
		nameserver = dns.Fqdn(nameserver) + ":" + strconv.Itoa(*port)
	}

	answer := executeQuery(nameserver, dns.TypePTR, qname[0])
	for _, a := range answer {
		if ptr, ok := a.(*dns.PTR); ok {
			fmt.Printf("%s\n", ptr.String())

			answer = executeQuery(nameserver, dns.TypePTR, ptr.Ptr)
			for _, a := range answer {
				if ptr, ok = a.(*dns.PTR); ok {
					fmt.Printf("%s\n", ptr.String())

					answer = executeQuery(nameserver, dns.TypeSRV, ptr.Ptr)
					for _, b := range answer {
						if srv, ok := b.(*dns.SRV); ok {
							fmt.Printf("%s\n", srv.String())
						}
					}
					answer = executeQuery(nameserver, dns.TypeTXT, ptr.Ptr)
					for _, b := range answer {
						if srv, ok := b.(*dns.TXT); ok {
							fmt.Printf("%s\n", srv.String())
						}
					}
				}
			}
		}
	}
}

func ExampleMX(nameserver string, qname []string) {
	if nameserver[0] == '[' && nameserver[len(nameserver)-1] == ']' {
		nameserver = nameserver[1 : len(nameserver)-1]
	}
	if i := net.ParseIP(nameserver); i != nil {
		nameserver = net.JoinHostPort(nameserver, strconv.Itoa(*port))
	} else {
		nameserver = dns.Fqdn(nameserver) + ":" + strconv.Itoa(*port)
	}

	c := new(dns.Client)
	m := new(dns.Msg)
	//m.SetQuestion("time.flows.emiliano.ar.", dns.TypePTR)
	m.SetQuestion(qname[0], dns.TypePTR)

	m.RecursionDesired = true
	r, _, err := c.Exchange(m, nameserver)
	if err != nil {
		return
	}
	if r.Rcode != dns.RcodeSuccess {
		return
	}

	for _, a := range r.Answer {
		if ptr, ok := a.(*dns.PTR); ok {
			fmt.Printf("%s\n", ptr.String())
		}
	}

	// answer, _ := executeQuery(nameserver, dns.TypePTR, qname[0])
	// for _, a := range answer {
	// 	if ptr, ok := a.(*dns.PTR); ok {
	// 		fmt.Printf("%s\n", ptr.String())
	// 	}
	// }

}

func executeQuery(nameserver string, qtype uint16, qname string) []dns.RR {
	c := new(dns.Client)
	m := new(dns.Msg)
	m.SetQuestion(qname, qtype)

	m.RecursionDesired = true
	r, _, err := c.Exchange(m, nameserver)
	if err != nil {
		return nil
	}
	if r.Rcode != dns.RcodeSuccess {
		return nil
	}

	return r.Answer
}
