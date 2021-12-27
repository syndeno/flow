/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

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
package cmd

import (
	"flow/cmd/config"
	"flow/cmd/create"
	"flow/cmd/describe"
	"flow/cmd/get"
	"flow/cmd/set"
	"flow/cmd/subscribe"
	"log"

	"os"

	"github.com/spf13/cobra"

	homedir "github.com/mitchellh/go-homedir"
	"github.com/spf13/viper"
)

var cfgFile string

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "flow",
	Short: "Brief description",
	Long:  `Long description`,
	// Uncomment the following line if your bare application
	// has an action associated with it:
	//	Run: func(cmd *cobra.Command, args []string) { },
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("*Executing rootCmd")
		debug, _ := cmd.Flags().GetBool("debug")
		log.Printf("*Debug %v", debug)

		// fmt.Println("viper.AllKeys()")
		// fmt.Println(viper.AllKeys())

		config := config.Config{}
		err := viper.Unmarshal(&config)
		if err != nil {
			log.Printf("unable to decode into struct, %v\n", err)
		} else {
			// fmt.Printf("Config loaded: %v\n", config)
		}
		// fmt.Println("viper.GetString(\"agents\"))")
		for _, agent := range config.Agents {
			log.Printf("Agent name: %v\n", agent.Name)
			log.Printf("\tfqdn: %v\n", agent.Fqdn)
			log.Printf("\tusername: %v\n", agent.Username)
			log.Printf("\tpassword: %v\n", agent.Password)
		}

		for _, namespace := range config.Namespaces {
			log.Printf("Namespace name: %v\n", namespace.Name)
			log.Printf("\tagent: %v\n", namespace.AgentName)

		}

	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func init() {
	log.Println("Initializing rootCmd")
	cobra.OnInitialize(initConfig)

	/****** FLOW COMMANDS ******/
	rootCmd.AddCommand(create.CreateCmd)
	rootCmd.AddCommand(get.GetCmd)
	rootCmd.AddCommand(set.SetCmd)
	rootCmd.AddCommand(describe.DescribeCmd)
	rootCmd.AddCommand(subscribe.SubscribeCmd)

	// Here you will define your flags and configuration settings.
	// Cobra supports persistent flags, which, if defined here,
	// will be global for your application.

	rootCmd.PersistentFlags().StringVarP(&cfgFile, "config", "c", "", "config file (default is $HOME/.flow.yml)")
	rootCmd.PersistentFlags().BoolP("debug", "d", false, "Enable debug")

	// Cobra also supports local flags, which will only run
	// when this action is called directly.
	// rootCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
	rootCmd.Flags().BoolP("version", "v", false, "Flow version 0.01")

}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	log.Println("Initializing initConfig")

	if cfgFile != "" {
		// Use config file from the flag.
		// fmt.Println("Setting config file from flag --config " + cfgFile)
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".flow" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".flow.yml")
	}

	viper.AutomaticEnv() // read in environment variables that match
	// log.Println(viper.ReadInConfig())
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("\tUsing config file:", viper.ConfigFileUsed())
	}

}
