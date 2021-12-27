package main

// install libpam0g-dev in ubuntu

import (
	"flag"
	"flow-agent/config"
	"flow-agent/server"
	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

/*
## Main

Main starts either a client or a server, depending on whether the `connect`
flag is set. Without the flag, the process starts as a server, listening
for incoming requests. With the flag the process starts as a client and connects
to the host specified by the flag value.

Try "localhost" or "127.0.0.1" when running both processes on the same machine.

*/
// const (
// 	Port = ":61000"
// )

// main
func main() {

	cfgFile := flag.String("config", "", "Configuration file")

	// flag.String("nameserver", "", "Nameserver to use")
	// flag.String("user", "", "Nameserver to use")
	// flag.String("password", "", "Nameserver to use")

	flag.Parse()

	if *cfgFile != "" {
		// Use config file from the flag.
		// fmt.Println("Setting config file from flag --config " + cfgFile)
		log.Printf("Using config file: %v", *cfgFile)

		viper.SetConfigFile(*cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		// Search config in home directory with name ".flow" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName("/etc/fnaa/fnaad.yml")
	}

	viper.AutomaticEnv() // read in environment variables that match
	// log.Println(viper.ReadInConfig())
	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		log.Println("\tUsing config file:", viper.ConfigFileUsed())
	}
	cfg := config.Config{}
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Println("Error:", errors.WithStack(err))
	}

	// Go into server mode.
	err = server.Server(cfg)
	if err != nil {
		log.Println("Error:", errors.WithStack(err))
	}

	log.Println("Server done.")
}

// The Lshortfile flag includes file name and line number in log messages.
func init() {
	log.SetFlags(log.Lshortfile)
}
