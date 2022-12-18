package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// Configurations
type Configurations struct {
	Server struct {
		Host        string
		Port        int
		SSL         bool
		User        string
		Password    string
		Connections int
	}
	Groups        string
	NzbFilename   string
	ParallelScans int
	Step          int
	Days          int
	Path          string
	Verbose       bool
}

var conf Configurations

func loadConfig() error {

	// Set the file name of the configurations file
	viper.SetConfigName("config")

	// Set the path to look for the configurations file
	viper.AddConfigPath(".")

	// Set config type to yaml
	viper.SetConfigType("yaml")

	if err := viper.ReadInConfig(); err != nil {
		if strings.Contains(err.Error(), "Not Found") {
			fmt.Printf("Config file \"config.yml\" not found. Creating config file...\n")
			defaultConfig := []byte(defaultConfig())
			if err := os.WriteFile("./config.yml", defaultConfig, 0644); err != nil {
				fmt.Printf("Error creating configuration file: %s\n", err)
				return err
			} else {
				fmt.Printf("Config file \"config.yml\" created. Please edit default values.\n")
				os.Exit(0)
			}
		} else {
			fmt.Printf("Error reading configuration file: %s\n", err)
			return err
		}
	}

	if err := viper.Unmarshal(&conf); err != nil {
		fmt.Printf("Unable to decode configure structure, %v\n", err)
		return err
	}

	if verbose {
		fmt.Println("Configuration loaded")
	}

	return nil
}
