package flow

import (
	"flow/client"
	"flow/cmd/config"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/miekg/dns"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var cfgFile string

var FlowDescribeCmd = &cobra.Command{
	Use:   "flow",
	Short: "A brief description of your command",
	Long:  `A longer description`,
	Run: func(cmd *cobra.Command, args []string) {
		debug, _ := cmd.Flags().GetBool("debug")
		log.Printf("*Debug %v", debug)

		if debug {
			log.Println("*Create flow")
		}

		cfg := config.Config{}
		err := viper.Unmarshal(&cfg)
		if err != nil {
			log.Printf("unable to decode into struct, %v\n", err)
		} else {
			// fmt.Printf("Config loaded: %v\n", config)
		}
		// fmt.Println("viper.GetString(\"agents\"))")
		// for _, agent := range config.Agents {
		// 	fmt.Printf("Agent name: %v\n", agent.Name)
		// 	fmt.Printf("\tfqdn: %v\n", agent.Fqdn)
		// 	fmt.Printf("\tusername: %v\n", agent.Username)
		// 	fmt.Printf("\tpassword: %v\n", agent.Password)
		// }

		// for _, namespace := range config.Namespaces {
		// 	fmt.Printf("Namespace name: %v\n", namespace.Name)
		// 	fmt.Printf("\tagent: %v\n", namespace.AgentName)
		// }

		/*Start Check if flowName is included*/
		argsNum := len(args)
		if argsNum == 0 {
			log.Println("Error: Too few arguments, specify flowName")
			os.Exit(1)

		} else if argsNum > 1 {
			log.Println("Error: Too many arguments, only flowName allowed")
			os.Exit(1)
		}

		flowNew := args[0]
		if debug {
			log.Printf("flow to create: %v", flowNew)
		}
		/*End Check if flowName is included*/

		/* Start check if user specified agent */
		// selectedAgent := viper.Get("agent")
		selectedAgent, _ := cmd.Flags().GetString("agent")

		var agentConfig config.Agent
		if selectedAgent != "" {
			if debug {
				log.Printf("agent selected: %v", selectedAgent)
			}
			//Find agent in config
			agentExists := false
			for _, agent := range cfg.Agents {
				if agent.Name == selectedAgent {
					agentExists = true
					agentConfig = agent
				}
			}

			if !agentExists {
				log.Println("Error: Agent does not exists, add to config file")
				os.Exit(1)
			}
		} else {
			if debug {
				log.Println("Debug: No agent selected, discovering by domain")
			}
		}

		/*Start discover agent for flow*/
		// If agent is set
		var namespaceConfig config.Namespace

		if (config.Agent{} != agentConfig) {
			agentManagesNamespace := false
			for _, namespace := range cfg.Namespaces {
				if namespace.AgentName == selectedAgent {
					if strings.Contains(flowNew, namespace.Name) {
						agentManagesNamespace = true
						namespaceConfig = namespace
					}
				}
				if !agentManagesNamespace {
					log.Println("Error: agent does not manage namespace")
					os.Exit(1)

				} else {
					if debug {
						log.Printf("Debug: agent %v manages namespace %v for new flow %v", agentConfig.Name, namespaceConfig.Name, flowNew)
					}
				}
			}
			// If agent is not set
		} else {
			var agentName string
			for _, namespace := range cfg.Namespaces {
				if strings.Contains(flowNew, namespace.Name) {
					agentName = namespace.AgentName
					namespaceConfig = namespace
				}
			}

			if agentName != "" {
				for _, agent := range cfg.Agents {
					if agent.Name == agentName {
						agentConfig = agent
					}
				}
				if (agentConfig != config.Agent{}) {
					if debug {
						log.Println("Debug: Discovered agent " + agentConfig.Name)
						log.Printf("Debug: agent %v manages namespace %v for new flow %v", agentConfig.Name, namespaceConfig.Name, flowNew)
					}
				} else {
					log.Println("Error: Namespace configured with inexistant agent")
					os.Exit(1)
				}
			} else {
				log.Println("Error: No agent not discovered")
				os.Exit(1)
			}
		}
		/*End discover agent for flow*/
		/* End check if user specified agent */

		/* Start flow creation*/
		if debug {
			log.Printf("-----------------")
			log.Printf("Creating new flow %v", flowNew)
			log.Printf("Resolving FNAA FQDN %v", agentConfig.Fqdn)
		}
		server, result := serviceResolve(agentConfig.Fqdn+".", "@172.17.0.4")
		if !result {
			log.Printf("Error: Could not resolve SRV RR for FQDN %v", agentConfig.Fqdn)
			os.Exit(1)
		}

		// log.Printf("DNS Resolution result: %v:%v", server.Host, server.Port)

		if debug {
			log.Printf("FNAA FQDN Resolved to %v port %v", server.Host, server.Port)
		}
		address, result := addressResolve(server.Host, "@172.17.0.4")
		if !result {
			log.Printf("Error: Could not resolve A RR for FQDN %v", server.Host)
			os.Exit(1)
		}
		// log.Printf("Connecting to %v port %v", address, server.Port)

		// conn, rw, err := client.Client(address, server.Port)
		conn, rw, err := client.Client(address, server.Port)

		if err != nil {
			log.Printf("Error: Connection to FNAA %v failed, %v", server.Host, err)
			os.Exit(1)
		}

		c := *conn
		defer c.Close()

		if debug {
			log.Printf("Connected to FNAA")
			log.Printf("Authenticating with PLAIN mechanism")
		}
		_, err = client.AuthenticatePlain(conn, rw, agentConfig.Username, agentConfig.Password)

		if err != nil {
			log.Printf("Error: Authentication to FNAA %v failed, %v", server.Host, err)
			os.Exit(1)
		}

		if debug {
			log.Printf("Authenticated")
		}

		// /* EXECUTING COMMAND CREATE FLOW */
		// if debug {
		// 	log.Printf("Executing command CREATE FLOW " + flowNew)
		// }
		// command := "CREATE FLOW " + flowNew
		// response, err := client.SendCommand(conn, rw, command)

		// if err != nil {
		// 	log.Printf("Error: Create Flow %v in FNAA %v failed, %v", flowNew, server.Host, err)
		// 	os.Exit(1)
		// }

		// if debug {
		// 	log.Printf("Flow %v created successfully", flowNew)
		// 	log.Printf("Server responded: %v", response)

		// }
		/* EXECUTING COMMAND DESCRIBE FLOW */
		if debug {
			log.Printf("Executing command DESCRIBE FLOW " + flowNew)
		}
		command := "DESCRIBE FLOW " + flowNew
		response, err := client.SendCommand(conn, rw, command)

		if err != nil {
			log.Printf("Error: Describe Flow %v in FNAA %v failed, %v", flowNew, server.Host, err)
			os.Exit(1)
		}
		log.Printf("Flow %v description:\n%v", flowNew, response)

		log.Printf("Flow %v described successfully", flowNew)

		// log.Printf("Got: 220 DATA")
		// log.Printf("Got: 220 OK")
		/* End flow creation*/

		log.Printf("Quitting")
		command = "QUIT"
		response, err = client.SendCommand(conn, rw, command)
		if err != nil {
			log.Printf("Error: Send command %v to FNAA %v failed, %v", command, server.Host, err)
			os.Exit(1)
		}

		c.Close()

		if debug {
			log.Printf("Connection closed")
		}
	},
}

func init() {
	log.Println("Initializing flow create")

	/****** FLOW COMMANDS ******/

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	// FlowCreateCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.flow.yml)")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// FlowCreateCmd.Flags().BoolP("debug", "d", false, "Enable debug")

	FlowDescribeCmd.Flags().String("agent", "", "Port to run Application server on")
	// viper.BindPFlag("agent", FlowCreateCmd.Flags().Lookup("agent"))

}

func serviceResolve(fqdn string, nameserver string) (fnaaServer, bool) {
	log.Println("**Starting FQDN resolution")

	var (
		// qtype  []uint16
		// qclass []uint16
		qname []string
	)

	qname = append(qname, fqdn)

	if len(nameserver) == 0 {
		conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		nameserver = "@" + conf.Servers[0]
	}

	nameserver = string([]byte(nameserver)[1:]) // chop off @
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

	server := fnaaServer{}
	// server.Host = "time.flows.unix.ar"
	// server.Port = 1
	log.Printf("**Resolving SRV for %v using server %v", fqdn, nameserver)

	answer := executeQuery(nameserver, dns.TypeSRV, qname[0])
	for _, a := range answer {
		if srv, ok := a.(*dns.SRV); ok {
			server.Port = int(srv.Port)
			server.Host = string(srv.Target)
			// log.Printf("**%s\n", srv.String())

			// answer = executeQuery(nameserver, dns.TypePTR, ptr.Ptr)
			// for _, a := range answer {
			// 	if ptr, ok = a.(*dns.PTR); ok {
			// 		fmt.Printf("**%s\n", ptr.String())

			// 		answer = executeQuery(nameserver, dns.TypeSRV, ptr.Ptr)
			// 		for _, b := range answer {
			// 			if srv, ok := b.(*dns.SRV); ok {
			// 				fmt.Printf("**%s\n", srv.String())
			// 			}
			// 		}
			// 		answer = executeQuery(nameserver, dns.TypeTXT, ptr.Ptr)
			// 		for _, b := range answer {
			// 			if srv, ok := b.(*dns.TXT); ok {
			// 				fmt.Printf("**%s\n", srv.String())
			// 			}
			// 		}
			// 	}
			// }
		} else {
			return server, false
		}
	}

	return server, true
}

type fnaaServer struct {
	Host string
	Port int
}

func addressResolve(fqdn string, nameserver string) (string, bool) {
	log.Println("**Starting Address resolution")

	var (
		// qtype  []uint16
		// qclass []uint16
		qname []string
	)

	qname = append(qname, fqdn)

	if len(nameserver) == 0 {
		conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
		if err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(2)
		}
		nameserver = "@" + conf.Servers[0]
	}

	nameserver = string([]byte(nameserver)[1:]) // chop off @
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
	answer := executeQuery(nameserver, dns.TypeA, qname[0])
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

func executeQuery(nameserver string, qtype uint16, qname string) []dns.RR {
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
