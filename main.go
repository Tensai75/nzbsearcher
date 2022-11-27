package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
)

var counter int

var waitGroups sync.WaitGroup

// search variables
var header string
var groups []string
var postDateUnix int64
var days int

var verbose bool

func main() {

	counter = 0
	start := time.Now()

	guard := make(chan struct{}, conf.ParallelScans)

	for _, group := range groups {
		guard <- struct{}{} // will block if guard channel is already filled
		waitGroups.Add(1)
		go func(group string) {
			defer func() {
				waitGroups.Done()
				<-guard
			}()

			if err := search(group); err != nil {
				fmt.Printf("Error searching in group '%s': %v\n", group, err)
			}

		}(group)
	}

	waitGroups.Wait()

	duration := time.Since(start)
	perSecond := float64(counter) / duration.Seconds()
	fmt.Printf("A total of %d messages were processed in %v (%d Messages/s)\n", counter, duration, int(perSecond))

}

func init() {

	// load configuration
	if err := loadConfig(); err != nil {
		fmt.Println("Fatal error while loading configuration file!")
		os.Exit(1)
	}

	// flags
	headerFlag := flag.String("header", "", "the header to search for")
	dateFlag := flag.String("date", "", "the date the header was posted (in the format DD.MM.YYYY)")
	groupsFlag := flag.String("groups", conf.Groups, `the group(s) to search in (separated by commas)
if set to an existing file, the groups listed in this file will be scanned (each group name must be on a separate line)
if set to 'ALL' all available groups on the usenet server will be scanned
if set to 'BINARIES' all available alt.binaries.* groups on the usenet server will be scanned
`)
	pathFlag := flag.String("path", conf.Path, "the path where the NZB file will be saved to")
	flag.IntVar(&conf.Days, "days", conf.Days, "the number of days to search back from the post date")
	flag.StringVar(&conf.Server.Host, "host", conf.Server.Host, "the usenet server host name")
	flag.IntVar(&conf.Server.Port, "port", conf.Server.Port, "the port for the usenet server")
	flag.BoolVar(&conf.Server.SSL, "ssl", conf.Server.SSL, "connect via SSL")
	flag.StringVar(&conf.Server.User, "user", conf.Server.User, "the username to login to the usenet server")
	flag.StringVar(&conf.Server.Password, "pass", conf.Server.Password, "the password to login to the usenet server")
	flag.IntVar(&conf.Server.Connections, "conn", conf.Server.Connections, "the number of connections to use")
	flag.IntVar(&conf.ParallelScans, "scans", conf.ParallelScans, "the number of groups to scan in parallel")
	flag.IntVar(&conf.Step, "step", conf.Step, "the number of message headers to retrieve in one header overview request")
	flag.BoolVar(&verbose, "verbose", conf.Verbose, "show verbose output")
	flag.Parse()

	// set header
	for header == "" {
		if *headerFlag != "" {
			header = *headerFlag
		} else {
			fmt.Print("Enter header to search for: ")
			header = inputReader()
		}
	}

	// set groups
	for len(groups) == 0 {
		var err error
		if *groupsFlag != "" {
			err = scanGroups(*groupsFlag)
			*groupsFlag = ""
		} else {
			fmt.Print("Enter group name(s) to search in: ")
			err = scanGroups(inputReader())
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	// set post date
	dateRegex := regexp.MustCompile(`[0-3]\d\.[0-1]\d\.(?:19|20)\d\d`)
	for postDateUnix == 0 {
		var input string
		if *dateFlag != "" {
			input = *dateFlag
			*dateFlag = ""
		} else {
			fmt.Print("Enter the date when the header was posted (DD.MM.YYYY): ")
			input = strings.TrimSpace(inputReader())
		}
		if input != "" {
			if match := dateRegex.FindStringIndex(input); match != nil {
				date, err := time.Parse("02.01.2006", input)
				if err != nil {
					fmt.Printf("Error parsing date '%s': %s\n", input, err)
				} else {
					postDateUnix = date.Unix() + 60*60*24 // add a day for security, i.e. if it was posted befor upload was finished
				}
			} else {
				fmt.Printf("Error parsing date '%s': Date does not have correct format DD.MM.YYYY!\n", input)
			}
		}
	}

	// set days to search
	days = conf.Days
	for days == 0 {
		var input string
		fmt.Print("Enter the amount of days to search back: ")
		input = strings.TrimSpace(inputReader())
		result, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Error parsing input '%s': %s\n", input, err)
		} else {
			days = result
		}
	}

	// set path
	var input string
	for input == "" {
		var err error
		if *pathFlag != "" {
			input = *pathFlag
			*pathFlag = ""
		} else {
			input = "./"
		}
		if _, err = os.Stat(input); err == nil {
			if verbose {
				fmt.Printf("Setting path for NZB files to: %s\n", input)
			}
			conf.Path = input
		} else if err != nil {
			fmt.Printf("Error for path '%s': %v\n", input, err)
			input = ""
		}
	}

}

func inputReader() string {
	reader := bufio.NewReader(os.Stdin)
	input, err := reader.ReadString('\n')
	if err != nil {
		if errors.Is(err, io.EOF) { // prefered way by GoLang doc
			os.Exit(0)
		}
		fmt.Println("An error occurred while reading input. Please try again", err)
		return ""
	}
	return strings.TrimSpace(input)
}
