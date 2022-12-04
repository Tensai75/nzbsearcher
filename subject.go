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

type groupMap struct {
	name    string
	headers map[string]headerMap
}

type headerMap struct {
	hash  string
	name  string
	group string
	files map[string]fileMap
}

type fileMap struct {
	hash     string
	name     string
	poster   string
	subject  string
	date     int64
	groups   []string
	number   int
	messages []Message
}

var hits map[string]groupMap

// var headers map[string]headerMap
var mutex = &sync.RWMutex{}

func parseSubject(message *Message, group string) error {

	if hits == nil {
		hits = make(map[string]groupMap)
	}

	searchPattern := regexp.MustCompile("(?i)" + regexp.QuoteMeta(header))

	if match := searchPattern.Match([]byte(message.subject)); match {

		// TODO: much better parsing to better account for all the very different subjects formats used for file posts...
		pattern1 := regexp.MustCompile(`^(?P<reminder>.+)(?:[\[\(] *(?P<segmentNo>\d+) *(?:/|of|von) *(?P<totalSegments>\d+) *[\)\]])`)
		pattern2 := regexp.MustCompile(`^(?P<header>.*?)?(?:(?:[\[\(]|File|Datei)? *(?P<segmentNo>\d+) *(?:/|of|von) *(?P<totalSegments>\d+) *[\)\]]?)(?P<reminder>.*)?`)
		pattern3 := regexp.MustCompile(`(?i)^(?P<header>.*?)?[- ]*"(?P<filename>(?P<basefilename>.*?)\.(?P<extension>(?:vol\d+\+\d+\.par2?|part\d+\.[^ "\.]*|[^ "\.]*\.\d+|[^ "\.]*))")`)
		pattern4 := regexp.MustCompile(`(?i)^(?P<filename>(?P<basefilename>.*?)\.(?P<extension>(?:vol\d+\+\d+\.par2?|part\d+\.[^ "\.]*|[^ "\.]*\.\d+|[^ "\.]*))(?:[" ]|$))`)

		if matches := findNamedMatches(pattern1, message.subject); matches != nil {
			message.segmentNo, _ = strconv.Atoi(matches["segmentNo"])
			message.totalSegments, _ = strconv.Atoi(matches["totalSegments"])
			reminder := matches["reminder"]
			if matches := findNamedMatches(pattern2, reminder); matches != nil {
				message.fileNo, _ = strconv.Atoi(matches["segmentNo"])
				message.totalFiles, _ = strconv.Atoi(matches["totalSegments"])
				message.header = strings.Trim(matches["header"], " -")
				reminder = matches["reminder"]
			}
			if matches := findNamedMatches(pattern3, reminder); matches != nil {
				if matches["header"] != "" {
					if message.header == "" {
						message.header = strings.Trim(matches["header"], " -")
					} else {
						message.header = message.header + " " + strings.Trim(matches["header"], " -")
					}
				}
				message.filename = strings.Trim(matches["filename"], " -")
				message.basefilename = strings.Trim(matches["basefilename"], " -")
			} else if matches := findNamedMatches(pattern4, reminder); matches != nil {
				message.filename = strings.Trim(matches["filename"], " -")
				message.basefilename = strings.Trim(matches["basefilename"], " -")
			}
			if message.basefilename != "" {
				if message.header == "" {
					message.header = message.basefilename
				} else {
					message.header = message.header + " - " + message.basefilename
				}
			}
			if message.header != "" {
				message.headerHash = GetMD5Hash(message.header + message.from + strconv.Itoa(message.totalFiles))
			} else {
				return errors.New("no header found")
			}
			if message.filename != "" {
				message.fileHash = GetMD5Hash(message.headerHash + message.filename + strconv.Itoa(message.totalSegments))
			} else {
				return errors.New("no filename found")
			}
		} else {
			return errors.New("subject did not match")
		}
		mutex.Lock()
		if _, ok := hits[group]; !ok {
			hits[group] = groupMap{
				name:    group,
				headers: make(map[string]headerMap),
			}
		}
		if _, ok := hits[group].headers[message.headerHash]; !ok {
			hits[group].headers[message.headerHash] = headerMap{
				hash:  message.headerHash,
				name:  message.header + " " + message.basefilename,
				group: group,
				files: make(map[string]fileMap),
			}
		}
		if _, ok := hits[group].headers[message.headerHash].files[message.fileHash]; !ok {
			hits[group].headers[message.headerHash].files[message.fileHash] = fileMap{
				hash:     message.fileHash,
				name:     message.filename,
				poster:   message.from,
				number:   message.fileNo,
				date:     message.date,
				subject:  message.subject,
				groups:   []string{group},
				messages: make([]Message, 0, 1),
			}
		}
		if files, ok := hits[group].headers[message.headerHash].files[message.fileHash]; ok {
			files.messages = append(files.messages, *message)
			hits[group].headers[message.headerHash].files[message.fileHash] = files
		}
		mutex.Unlock()

	}
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

func GetMD5Hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}
