package main

import (
	"crypto/md5"
	"encoding/hex"
	"errors"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

type header struct {
	name        string
	hash        string
	filesByHash map[string]*file
}

type message struct {
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
}

type file struct {
	name     string
	hash     string
	poster   string
	subject  string
	date     int64
	groups   []string
	number   int
	messages []message
}

var (
	headersByGroupAndHeaderHash = make(map[string]map[string]*header)
	mutex                       sync.Mutex
)

var (
	// TODO: much better parsing to better account for all the very different subjects formats used for file posts...
	// TODO: rename the variables to better reflect what they try to match
	pattern1 = regexp.MustCompile(`^(?P<reminder>.+)(?:[\[\(] *(?P<segmentNo>\d+) *(?:/|of|von) *(?P<totalSegments>\d+) *[\)\]])`)
	pattern2 = regexp.MustCompile(`^(?P<header>.*?)?(?:(?:[\[\(]|File|Datei)? *(?P<segmentNo>\d+) *(?:/|of|von) *(?P<totalSegments>\d+) *[\)\]]?)(?P<reminder>.*)?`)
	pattern3 = regexp.MustCompile(`(?i)^(?P<header>.*?)?[- ]*"(?P<filename>(?P<basefilename>.*?)\.(?P<extension>(?:vol\d+\+\d+\.par2?|part\d+\.[^ "\.]*|[^ "\.]*\.\d+|[^ "\.]*))")`)
	pattern4 = regexp.MustCompile(`(?i)^(?P<filename>(?P<basefilename>.*?)\.(?P<extension>(?:vol\d+\+\d+\.par2?|part\d+\.[^ "\.]*|[^ "\.]*\.\d+|[^ "\.]*))(?:[" ]|$))`)
)

func parseSubject(msg *message, group string) error {
	searchPattern := regexp.MustCompile("(?i)" + regexp.QuoteMeta(headerToSearch))
	if match := searchPattern.Match([]byte(msg.subject)); !match {
		return nil
	}
	var matches map[string]string
	if matches = findNamedMatches(pattern1, msg.subject); matches == nil {
		return errors.New("subject did not match")
	}
	msg.segmentNo, _ = strconv.Atoi(matches["segmentNo"])
	msg.totalSegments, _ = strconv.Atoi(matches["totalSegments"])
	reminder := matches["reminder"]
	if matches := findNamedMatches(pattern2, reminder); matches != nil {
		msg.fileNo, _ = strconv.Atoi(matches["segmentNo"])
		msg.totalFiles, _ = strconv.Atoi(matches["totalSegments"])
		msg.header = strings.Trim(matches["header"], " -")
		reminder = matches["reminder"]
	}
	if matches := findNamedMatches(pattern3, reminder); matches != nil {
		if matches["header"] != "" {
			if msg.header == "" {
				msg.header = strings.Trim(matches["header"], " -")
			} else {
				msg.header = msg.header + " " + strings.Trim(matches["header"], " -")
			}
		}
		msg.filename = strings.Trim(matches["filename"], " -")
		msg.basefilename = strings.Trim(matches["basefilename"], " -")
	} else if matches := findNamedMatches(pattern4, reminder); matches != nil {
		msg.filename = strings.Trim(matches["filename"], " -")
		msg.basefilename = strings.Trim(matches["basefilename"], " -")
	}
	if msg.basefilename != "" {
		if msg.header == "" {
			msg.header = msg.basefilename
		} else {
			msg.header = msg.header + " - " + msg.basefilename
		}
	}
	if msg.header == "" {
		return errors.New("no header found")
	}
	headerHash := getMD5Hash(msg.header + msg.from + strconv.Itoa(msg.totalFiles))
	if msg.filename == "" {
		return errors.New("no filename found")
	}
	fileHash := getMD5Hash(headerHash + msg.filename + strconv.Itoa(msg.totalSegments))
	mutex.Lock()
	headersByHash, ok := headersByGroupAndHeaderHash[group]
	if !ok {
		headersByHash = make(map[string]*header)
		headersByGroupAndHeaderHash[group] = headersByHash
	}
	hdr, ok := headersByHash[headerHash]
	if !ok {
		hdr = &header{
			name:        msg.header + " " + msg.basefilename,
			hash:        headerHash,
			filesByHash: make(map[string]*file),
		}
		headersByHash[headerHash] = hdr
	}
	f, ok := hdr.filesByHash[fileHash]
	if !ok {
		f = &file{
			name:     msg.filename,
			hash:     fileHash,
			poster:   msg.from,
			number:   msg.fileNo,
			date:     msg.date,
			subject:  msg.subject,
			groups:   []string{group},
			messages: make([]message, 0, 1),
		}
		hdr.filesByHash[fileHash] = f
	}
	f.messages = append(f.messages, *msg)
	mutex.Unlock()
	return nil
}

func findNamedMatches(regex *regexp.Regexp, str string) map[string]string {
	match := regex.FindStringSubmatch(str)
	if match == nil {
		return nil
	}
	results := map[string]string{}
	for i, name := range match {
		results[regex.SubexpNames()[i]] = name
	}
	return results
}

func getMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
