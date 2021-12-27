package subscribe

import (
	"flow/client"
	"flow/cmd/config"
	"fmt"
	"log"
	"os"

	"github.com/miekg/dns"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// createCmd represents the create command
var SubscribeCmd = &cobra.Command{
	Use:   "subscribe",
	Short: "A brief description of your command",
	Long:  `Subscribe`,
	Run: func(cmd *cobra.Command, args []string) {
		debug, _ := cmd.Flags().GetBool("debug")
		nameserver, _ := cmd.Flags().GetString("nameserver")

		if len(nameserver) == 0 {
			conf, err := dns.ClientConfigFromFile("/etc/resolv.conf")
			if err != nil {
				fmt.Fprintln(os.Stderr, err)
				os.Exit(2)
			}
			nameserver = conf.Servers[0]
		}

		// nameserver = string([]byte(nameserver)[0:]) // chop off @

		log.Printf("*Debug %v", debug)

		if debug {
			log.Println("*Subscribe to flow")
		}

		cfg := config.Config{}
		err := viper.Unmarshal(&cfg)
		if err != nil {
			log.Printf("unable to decode into struct, %v\n", err)
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
			if len(cfg.Agents) > 1 {
				log.Println("Error: More than one agent in config file, select with --agent")
				os.Exit(1)
			} else {
				agentConfig = cfg.Agents[0]
			}
		}

		/* Start flow creation*/
		if debug {
			log.Printf("-----------------")
			log.Printf("Creating new flow %v", flowNew)
			log.Printf("Resolving FNAA FQDN %v", agentConfig.Fqdn)
		}
		// server, result := serviceResolve(agentConfig.Fqdn+".", "@172.17.0.2")
		server, result := client.ServiceResolve(agentConfig.Fqdn+".", nameserver)

		if !result {
			log.Printf("Error: Could not resolve SRV RR for FQDN %v", agentConfig.Fqdn)
			os.Exit(1)
		}

		// log.Printf("DNS Resolution result: %v:%v", server.Host, server.Port)

		if debug {
			log.Printf("FNAA FQDN Resolved to %v port %v", server.Host, server.Port)
		}
		address, result := client.AddressResolve(server.Host, nameserver)
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

		/* EXECUTING COMMAND SUBSCRIBE FLOW */
		if debug {
			log.Printf("Executing command SUBSCRIBE " + flowNew + " LOCAL " + agentConfig.Prefix + flowNew)
		}
		command := "SUBSCRIBE " + flowNew + " LOCAL " + agentConfig.Prefix + flowNew
		response, err := client.SendCommand(conn, rw, command)

		if err != nil {
			log.Printf("Error: Subscribe to Flow %v in FNAA %v failed, %v", flowNew, server.Host, err)
			os.Exit(1)
		}

		if debug {
			log.Printf("Flow %v subscription created successfully", agentConfig.Prefix+flowNew)
			log.Printf("Server responded: %v", response)

		}
		// /* EXECUTING COMMAND DESCRIBE FLOW */
		// if debug {
		// 	log.Printf("Executing command DESCRIBE FLOW " + flowNew)
		// }
		// command = "DESCRIBE FLOW " + flowNew
		// response, err = client.SendCommand(conn, rw, command)

		// if err != nil {
		// 	log.Printf("Error: Describe Flow %v in FNAA %v failed, %v", flowNew, server.Host, err)
		// 	os.Exit(1)
		// }
		// log.Printf("Flow %v description:\n%v", flowNew, response)

		// log.Printf("Flow %v described successfully", flowNew)

		// // log.Printf("Got: 220 DATA")
		// // log.Printf("Got: 220 OK")
		// /* End flow creation*/

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

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// SubscribeCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	SubscribeCmd.Flags().String("nameserver", "", "Override system nameserver")
	SubscribeCmd.Flags().String("agent", "", "Select FNAA")

}
