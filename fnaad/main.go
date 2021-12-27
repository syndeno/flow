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

package main

import (
	"flag"
	"fnaad/config"
	"fnaad/server"

	"log"
	"os"

	"github.com/mitchellh/go-homedir"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
)

func main() {

	//Read config file from command line argument
	cfgFile := flag.String("config", "", "Configuration file")

	flag.Parse()

	//If no config file specified as argument, look for default
	if *cfgFile != "" {
		log.Printf("Using config file: %v", *cfgFile)

		viper.SetConfigFile(*cfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			log.Println(err)
			os.Exit(1)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName("/etc/fnaa/fnaad.yml")
	}

	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err == nil {
		log.Println("\tUsing config file:", viper.ConfigFileUsed())
	}

	//Initialize configuration file
	cfg := config.Config{}
	err := viper.Unmarshal(&cfg)
	if err != nil {
		log.Println("Error:", errors.WithStack(err))
	}

	//Launch server
	err = server.Server(cfg)
	if err != nil {
		log.Println("Error:", errors.WithStack(err))
	}

	log.Println("Server done.")
}

func init() {
	log.SetFlags(log.Lshortfile)
}
