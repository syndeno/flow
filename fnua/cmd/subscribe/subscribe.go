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
package subscribe

import (
	"fmt"
	"fnua/client"
	"fnua/cmd/config"
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

		log.Printf("*Debug %v", debug)

		if debug {
			log.Println("*Subscribe to flow")
		}

		cfg := config.Config{}
		err := viper.Unmarshal(&cfg)
		if err != nil {
			log.Printf("unable to decode into struct, %v\n", err)
		}

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
		server, result := client.ServiceResolve(agentConfig.Fqdn+".", nameserver)

		if !result {
			log.Printf("Error: Could not resolve SRV RR for FQDN %v", agentConfig.Fqdn)
			os.Exit(1)
		}

		if debug {
			log.Printf("FNAA FQDN Resolved to %v port %v", server.Host, server.Port)
		}
		address, result := client.AddressResolve(server.Host, nameserver)
		if !result {
			log.Printf("Error: Could not resolve A RR for FQDN %v", server.Host)
			os.Exit(1)
		}

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

	SubscribeCmd.Flags().String("nameserver", "", "Override system nameserver")
	SubscribeCmd.Flags().String("agent", "", "Select FNAA")

}
