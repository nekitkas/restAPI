package main

import (
	"flag"
	"github.com/nekitkas/restAPI/internal/app/apiserver"
	"log"
)

var configPath string

func init() {
	flag.StringVar(&configPath, "config-path", "configs/config.json", "path to config")
}

func main() {
	flag.Parse()

	config := apiserver.NewConfig()
	err := config.ReadConfig(configPath)
	if err != nil {
		log.Fatalf("Error reading config file: %s\n", err)
	}

	s := apiserver.New(config)
	log.Fatal(s.Start())
}
