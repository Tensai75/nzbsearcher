package main

import (
	"errors"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Tensai75/nntp"
	"github.com/kennygrant/sanitize"
)

type Message struct {
	messageNo     int
	subject       string
	messageId     string
	from          string
	bytes         int
	date          int64
	header        string
	filename      string
	basefilename  string
	fileNo        int
	totalFiles    int
	segmentNo     int
	totalSegments int
	headerHash    string
	fileHash      string
	group         string
}

var searches sync.WaitGroup
var nzbSaves sync.WaitGroup

func search(group string) error {
	var startMessageID, currentMessageID int
	var currentMessageDate, lastMessageDate time.Time
	fmt.Printf("Switching to group '%s' and retrieving group information from the usenet server\n", group)
	conn, firstMessageID, lastMessageID, err := switchToGroup(group)
	if err != nil {
		return err
	}
	if verbose {
		fmt.Printf("First / last message in group '%s' are: %d | %d\n", group, firstMessageID, lastMessageID)
	}
	if verbose {
		fmt.Printf("Scanning group '%s' for the last message to end the search\n", group)
	}
	lastMessageID, lastMessageDate, err = scanForDate(conn, firstMessageID, lastMessageID, 0, false)
	if err != nil {
		DisconnectNNTP(conn)
		fmt.Printf("Error while scanning group '%s' for the last message: %v\n", group, err)
		return err
	}
	if verbose {
		fmt.Printf("Last message in group '%s' to end the search is %d, uploaded on %s\n", group, lastMessageID, lastMessageDate)
	}
	if verbose {
		fmt.Printf("Scanning group '%s'for the first message to start the search\n", group)
	}
	currentMessageID, currentMessageDate, err = scanForDate(conn, firstMessageID, lastMessageID, -1*days*60*60*24, true)
	if err != nil {
		DisconnectNNTP(conn)
		fmt.Printf("Error while scanning group '%s' for the first message: %v\n", group, err)
		return err
	}
	if verbose {
		fmt.Printf("First message in group '%s' to start the search is %d, uploaded on %s\n", group, currentMessageID, currentMessageDate)
	}
	if currentMessageID >= lastMessageID {
		DisconnectNNTP(conn)
		return errors.New("no messages found within search range")
	}
	DisconnectNNTP(conn)
	startMessageID = currentMessageID
	fmt.Printf("Start searching messages %d to %d from %s to %s in group '%s'\n", startMessageID, lastMessageID, currentMessageDate, lastMessageDate, group)
	for currentMessageID <= lastMessageID {

		var lastMessage int
		if currentMessageID+conf.Step > lastMessageID {
			lastMessage = lastMessageID
		} else {
			lastMessage = currentMessageID + conf.Step
		}
		searches.Add(1)
		go searchMessages(currentMessageID, lastMessage, group)
		// update currentMessageID for next request
		currentMessageID = lastMessage + 1

	}
	searches.Wait()
	fmt.Printf("Finished searching in group '%s'\n", group)
	if verbose {
		fmt.Printf("Messages %d to %d were searched in group '%s'\n", startMessageID, currentMessageID-1, group)
	}
	if _, ok := hits[group]; !ok {
		fmt.Printf("Header not found!\n")
		return nil
	} else {
		mutex.Lock()
		for _, headerMap := range hits[group].headers {
			fmt.Printf("Found header '%s' in group '%s'\n", headerMap.name, hits[group].name)
			if verbose {
				fmt.Printf("Generating NZB file\n")
			}
			nzbSaves.Add(1)
			go saveNZB(headerMap, hits[group].name)
		}
		nzbSaves.Wait()
		mutex.Unlock()
	}
	return nil
}

func saveNZB(headerMap headerMap, group string) error {
	defer nzbSaves.Done()
	nzb := []string{
		`<?xml version="1.0" encoding="UTF-8"?>`,
		`<!DOCTYPE nzb PUBLIC "-//newzBin//DTD NZB 1.1//EN" "http://www.newzbin.com/DTD/nzb/nzb-1.1.dtd">`,
		`<nzb xmlns="http://www.newzbin.com/DTD/2003/nzb">`,
		`<!-- NZB file created by https://github.com/Tensai75/nzbsearcher, coded by Tensai -->`,
		`<head>`,
		`</head>`,
	}
	for _, fileMap := range headerMap.files {
		nzb = append(nzb, fmt.Sprintf(`<file poster="%s" date="%d" subject="%s">`, html.EscapeString(fileMap.poster), fileMap.date, html.EscapeString(fileMap.subject)))
		nzb = append(nzb, `  <groups>`)
		for _, group := range fileMap.groups {
			nzb = append(nzb, fmt.Sprintf(`    <group>%s</group>`, group))
		}
		nzb = append(nzb, `  </groups>`)
		nzb = append(nzb, `  <segments>`)
		for _, message := range fileMap.messages {
			nzb = append(nzb, fmt.Sprintf(`    <segment bytes="%d" number="%d">%s</segment>`, message.bytes, message.segmentNo, html.EscapeString(message.messageId)))
		}
		nzb = append(nzb, `  </segments>`)
		nzb = append(nzb, `</file>`)
	}
	nzb = append(nzb, `</nzb>`)
	filepath := filepath.Join(conf.Path, sanitize.Name(headerMap.name+"_"+group+".nzb"))
	f, err := os.Create(filepath)
	if err != nil {
		fmt.Printf("Error saving NZB file '%s': %v\n", filepath, err)
		f.Close()
		return err
	}
	for _, line := range nzb {
		_, err = fmt.Fprintln(f, line)
		if err != nil {
			fmt.Printf("Error saving NZB file '%s': %v\n", filepath, err)
			f.Close()
			return err
		}
	}
	f.Close()
	fmt.Printf("NZB file '%s' saved to disk\n", filepath)
	return nil
}

func searchMessages(firstMessage int, lastMessage int, group string) error {
	defer searches.Done()
	conn, firstMessageID, lastMessageID, err := switchToGroup(group)
	if err != nil {
		return err
	}
	if firstMessage < firstMessageID {
		firstMessage = firstMessageID
	}
	if lastMessage > lastMessageID {
		lastMessage = lastMessageID
	}
	if verbose {
		fmt.Printf("Loading message overview from messages %d to %d in group '%s'\n", firstMessage, lastMessage, group)
	}
	results, err := conn.Overview(firstMessage, lastMessage)
	DisconnectNNTP(conn)
	if err != nil {
		fmt.Printf("Error retrieving message overview from the usenet server while searching in group '%s': %v\n", group, err)
		return err
	}
	for _, overview := range results {
		currentDate := overview.Date.Unix()
		if currentDate >= postDateUnix {
			return nil
		}
		var message Message
		message.messageNo = overview.MessageNumber
		message.subject = html.UnescapeString(strings.ToValidUTF8(overview.Subject, ""))
		message.messageId = strings.Trim(overview.MessageId, "<>")
		message.from = strings.ToValidUTF8(overview.From, "")
		message.bytes = overview.Bytes
		if date := overview.Date.Unix(); date < 0 {
			message.date = 0
		} else {
			message.date = date
		}
		message.fileNo = 1
		message.totalFiles = 1
		message.segmentNo = 1
		message.totalSegments = 1
		message.group = group
		if err := parseSubject(&message, group); err != nil {
			// message probably did not contain a yEnc encoded file?
			if verbose {
				fmt.Printf("Parsing error while searching in group '%s': %v\n", group, err)
			}
		}
		counter = counter + 1
	}
	return nil
}

func scanForDate(conn *nntp.Conn, firstMessageID int, lastMessageID int, interval int, first bool) (int, time.Time, error) {
	currentMessageID := firstMessageID
	endMessageID := lastMessageID
	scanStep := endMessageID - currentMessageID
	for currentMessageID <= endMessageID {
		step := 0
		if currentMessageID == firstMessageID {
			step = 2000
		}
		if scanStep < 1000 {
			results, err := conn.Overview(currentMessageID-1000, currentMessageID+1000)
			if err != nil {
				return 0, time.Time{}, err
			}
			for _, overview := range results {
				if overview.Date.Unix() > postDateUnix+int64(interval) {
					return overview.MessageNumber, overview.Date, nil
				}
			}
			return results[len(results)-1].MessageNumber, results[len(results)-1].Date, nil
		} else {
			if verbose {
				fmt.Printf("Scanning message no.: %d| ScanStep: %d\n", currentMessageID, scanStep)
			}
			results, err := conn.Overview(currentMessageID, currentMessageID+step)
			if err != nil {
				return 0, time.Time{}, err
			}
			if len(results) == 0 {
				return 0, time.Time{}, errors.New("Overview results are empty")
			}
			overview := results[0]
			currentDate := overview.Date.Unix()
			scanStep = scanStep / 2
			if first && currentMessageID == firstMessageID && currentDate > postDateUnix+int64(interval) {
				return overview.MessageNumber, overview.Date, nil
			} else if !first && currentMessageID == firstMessageID && currentDate > postDateUnix+int64(interval) {
				return 0, time.Time{}, errors.New("post date is older than oldest message of this group")
			}
			if currentDate < postDateUnix+int64(interval) {
				currentMessageID = currentMessageID + scanStep
			}
			if currentDate > postDateUnix+int64(interval) {
				currentMessageID = currentMessageID - scanStep
			}
		}

	}
	return 0, time.Time{}, errors.New("no messages found within search range")
}

func switchToGroup(group string) (*nntp.Conn, int, int, error) {
	conn, err := ConnectNNTP()
	if err != nil {
		fmt.Printf("Error connecting to the usenet server: %v\n", err)
		return nil, 0, 0, err
	}
	_, firstMessageID, lastMessageID, err := conn.Group(group)
	if err != nil {
		fmt.Printf("Error retrieving group information for group '%s' from the usenet server: %v\n", group, err)
		return nil, 0, 0, err
	}
	return conn, firstMessageID, lastMessageID, nil
}
