package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

func scanGroups(groupsString string) error {

	if _, err := os.Stat(groupsString); err == nil {
		if verbose {
			if verbose {
				fmt.Printf("Reading groups file '%s'\n", groupsString)
			}
		}
		err := readGroups(groupsString)
		if err != nil {
			fmt.Printf("Error while reading groups file '%s': %v\n", groupsString, err)
		}
	} else if groupsString == "BINARIES" || groupsString == "ALL" {
		if verbose {
			fmt.Println("Connecting to usenet server to get groups list")
		}
		conn, err := ConnectNNTP()
		defer DisconnectNNTP(conn)
		if err != nil {
			fmt.Printf("Error while connecting to usenet server: %v\n", err)
		}
		filter := ""
		if groupsString == "BINARIES" {
			filter = "alt.binaries.*"
		}
		groupsList, err := conn.List("ACTIVE", filter)
		if err != nil {
			fmt.Printf("Error while requesting list of groups: %v\n", err)
		}
		DisconnectNNTP(conn)
		if verbose {
			fmt.Println("Processing the groups")
		}
		for _, group := range groupsList {
			groupData := strings.Split(string(group), " ")
			groups = append(groups, groupData[0])
		}
	} else {
		groups = strings.Split(groupsString, ",")
		for i, group := range groups {
			groups[i] = strings.TrimSpace(group)
		}
	}
	if len(groups) == 0 {
		return errors.New("no groups found")
	}
	return nil
}

func readGroups(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		groups = append(groups, scanner.Text())
	}
	return scanner.Err()
}
