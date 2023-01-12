package main

import (
	"errors"
	"fmt"
	"html"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/Tensai75/nntp"
	"github.com/kennygrant/sanitize"
)

const (
	secondsPerDay = 60 * 60 * 24
)

func search(group string) error {
	fmt.Printf("Switching to group '%s' and retrieving group information from the usenet server\n", group)
	conn, firstMessageID, lastMessageID, err := switchToGroup(group)
	if err != nil {
		return err
	}
	defer DisconnectNNTP(conn)
	if verbose {
		fmt.Printf("First / last message in group '%s' are: %d | %d\n", group, firstMessageID, lastMessageID)
		fmt.Printf("Scanning group '%s' for the last message to end the search\n", group)
	}
	lastMessageID, lastMessageDate, err := scanForDate(conn, firstMessageID, lastMessageID, 0, false)
	if err != nil {
		fmt.Printf("Error while scanning group '%s' for the last message: %v\n", group, err)
		return err
	}
	if verbose {
		fmt.Printf("Last message in group '%s' to end the search is %d, uploaded on %s\n", group, lastMessageID, lastMessageDate)
		fmt.Printf("Scanning group '%s'for the first message to start the search\n", group)
	}
	currentMessageID, currentMessageDate, err := scanForDate(conn, firstMessageID, lastMessageID, -secondsPerDay*days, true)
	if err != nil {
		fmt.Printf("Error while scanning group '%s' for the first message: %v\n", group, err)
		return err
	}
	DisconnectNNTP(conn)
	if verbose {
		fmt.Printf("First message in group '%s' to start the search is %d, uploaded on %s\n", group, currentMessageID, currentMessageDate)
	}
	if currentMessageID >= lastMessageID {
		return errors.New("no messages found within search range")
	}
	startMessageID := currentMessageID
	fmt.Printf("Start searching messages %d to %d from %s to %s in group '%s'\n", startMessageID, lastMessageID, currentMessageDate, lastMessageDate, group)
	var wg sync.WaitGroup
	for currentMessageID <= lastMessageID {
		lastMessage := int(math.Min(float64(currentMessageID+conf.Step), float64(lastMessageID)))
		wg.Add(1)
		go func(currentMessageID int) {
			defer wg.Done()
			searchMessages(currentMessageID, lastMessage, group)
		}(currentMessageID)
		// update currentMessageID for next request
		currentMessageID = lastMessage + 1
	}
	wg.Wait()
	fmt.Printf("Finished searching in group '%s'\n", group)
	if verbose {
		fmt.Printf("Messages %d to %d were searched in group '%s'\n", startMessageID, currentMessageID-1, group)
	}
	headers, ok := headersByGroupAndHeaderHash[group]
	if !ok {
		fmt.Printf("Header not found in group %s!\n", group)
		return nil
	}
	for _, hdr := range headers {
		fmt.Printf("Found header '%s' in group '%s'\n", hdr.name, group)
		if verbose {
			fmt.Printf("Generating NZB file\n")
		}
		saveNZB(hdr, group)
	}
	return nil
}

const (
	maxFilenameLength = 255

	nzbHeader = `<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE nzb PUBLIC "-//newzBin//DTD NZB 1.1//EN" "http://www.newzbin.com/DTD/nzb/nzb-1.1.dtd">
<nzb xmlns="http://www.newzbin.com/DTD/2003/nzb">
<!-- NZB file created by https://github.com/Tensai75/nzbsearcher, coded by Tensai -->
<head>
</head>
`
)

func saveNZB(hdr *header, group string) error {
	var nzb strings.Builder
	nzb.WriteString(nzbHeader)
	for _, fileMap := range hdr.filesByHash {
		nzb.WriteString(fmt.Sprintf(`<file poster="%s" date="%d" subject="%s">`, html.EscapeString(fileMap.poster), fileMap.date, html.EscapeString(fileMap.subject)))
		nzb.WriteByte('\n')
		nzb.WriteString("  <groups>\n")
		for _, group := range fileMap.groups {
			nzb.WriteString(fmt.Sprintf("    <group>%s</group>\n", group))
		}
		nzb.WriteString("  </groups>\n")
		nzb.WriteString("  <segments>\n")
		for _, message := range fileMap.messages {
			nzb.WriteString(fmt.Sprintf(`    <segment bytes="%d" number="%d">%s</segment>`, message.bytes, message.segmentNo, html.EscapeString(message.messageId)))
			nzb.WriteByte('\n')
		}
		nzb.WriteString("  </segments>\n")
		nzb.WriteString("</file>\n")
	}
	nzb.WriteString("</nzb>\n")
	filename := sanitize.Name(hdr.name + "_" + group + ".nzb")
	if len(filename) > maxFilenameLength {
		filename = filename[len(filename)-maxFilenameLength:]
	}
	filepath := filepath.Join(conf.Path, filename)
	f, err := os.Create(filepath)
	if err != nil {
		fmt.Printf("Error creating file '%s' to save NZB: %v\n", filepath, err)
		return err
	}
	defer f.Close()
	if _, err := io.Copy(f, strings.NewReader(nzb.String())); err != nil {
		fmt.Printf("Error writing NZB to file '%s': %v\n", filepath, err)
		return err
	}
	fmt.Printf("NZB file '%s' saved to disk\n", filepath)
	return nil
}

func searchMessages(firstMessage int, lastMessage int, group string) error {
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
		var message message
		message.messageNo = overview.MessageNumber
		message.subject = html.UnescapeString(strings.ToValidUTF8(overview.Subject, ""))
		message.messageId = strings.Trim(overview.MessageId, "<>")
		message.from = strings.ToValidUTF8(overview.From, "")
		message.bytes = overview.Bytes
		if date := overview.Date.Unix(); date > 0 {
			message.date = date
		}
		message.fileNo = 1
		message.totalFiles = 1
		message.segmentNo = 1
		message.totalSegments = 1
		if err := parseSubject(&message, group); err != nil {
			// message probably did not contain a yEnc encoded file?
			if verbose {
				fmt.Printf("Parsing error while searching in group '%s': %v\n", group, err)
			}
		}
		atomic.AddUint64(&counter, 1)
	}
	return nil
}

func scanForDate(conn *nntp.Conn, firstMessageID int, lastMessageID int, interval int, first bool) (int, time.Time, error) {
	currentMessageID := firstMessageID
	endMessageID := lastMessageID
	scanStep := lastMessageID - firstMessageID
	endTimestamp := postDateUnix + int64(interval)
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
				if overview.Date.Unix() > endTimestamp {
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
			currentTimestamp := overview.Date.Unix()
			scanStep = scanStep / 2
			if first && currentMessageID == firstMessageID && currentTimestamp > endTimestamp {
				return overview.MessageNumber, overview.Date, nil
			} else if !first && currentMessageID == firstMessageID && currentTimestamp > endTimestamp {
				return 0, time.Time{}, errors.New("post date is older than oldest message of this group")
			}
			if currentTimestamp < endTimestamp {
				currentMessageID = currentMessageID + scanStep
			}
			if currentTimestamp > endTimestamp {
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
