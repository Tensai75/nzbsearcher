package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
)

const (
	allGroups       = "ALL"
	allBinaryGroups = "BINARIES"
)

var (
	ErrNoGroups = errors.New("no groups found")
)

func scanGroups(groupsString string) error {
	if groupsString == allBinaryGroups || groupsString == allGroups {
		if verbose {
			fmt.Println("Connecting to usenet server to get groups list")
		}
		conn, err := ConnectNNTP()
		defer DisconnectNNTP(conn)
		if err != nil {
			fmt.Printf("Error while connecting to usenet server: %v\n", err)
			return ErrNoGroups
		}
		filter := ""
		if groupsString == allBinaryGroups {
			filter = "alt.binaries.*"
		}
		groupsList, err := conn.List("ACTIVE", filter)
		if err != nil {
			fmt.Printf("Error while requesting list of groups: %v\n", err)
			return ErrNoGroups
		}
		if verbose {
			fmt.Println("Processing the groups")
		}
		for _, group := range groupsList {
			groupData := strings.Split(string(group), " ")
			groups = append(groups, groupData[0])
		}
	} else if _, err := os.Stat(groupsString); err == nil {
		if verbose {
			fmt.Printf("Reading groups file '%s'\n", groupsString)
		}
		err := readGroups(groupsString)
		if err != nil {
			fmt.Printf("Error while reading groups file '%s': %v\n", groupsString, err)
			return ErrNoGroups
		}
	} else {
		groups = strings.Split(groupsString, ",")
		for i, group := range groups {
			groups[i] = strings.Replace(strings.TrimSpace(group), "a.b.", "alt.binaries.", 1)
		}
	}
	if len(groups) == 0 {
		return ErrNoGroups
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
		groups = append(groups, strings.Replace(strings.TrimSpace(scanner.Text()), "a.b.", "alt.binaries.", 1))
	}
	return scanner.Err()
}
