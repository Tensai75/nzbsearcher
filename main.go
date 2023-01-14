package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var (
	counter   uint64
	waitGroup sync.WaitGroup

	// search variables
	headerToSearch string
	groups         []string
	postDateUnix   int64
	days           int

	verbose bool
)

func main() {
	start := time.Now()

	guard := make(chan struct{}, conf.ParallelScans)

	for _, group := range groups {
		guard <- struct{}{} // will block if guard channel is already filled
		waitGroup.Add(1)
		go func(group string) {
			defer func() {
				waitGroup.Done()
				<-guard
			}()

			if err := search(group); err != nil {
				fmt.Printf("Error searching in group '%s': %v\n", group, err)
			}
		}(group)
	}
	waitGroup.Wait()

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

	var (
		date       string
		groupsFlag string
		path       string
	)

	// flags
	flag.StringVar(&headerToSearch, "header", "", "the header to search for")
	flag.StringVar(&date, "date", "", "the date the header was posted (in the format DD.MM.YYYY or YYYY-MM-dd)")
	flag.StringVar(&groupsFlag, "groups", conf.Groups, `the group(s) to search in (separated by commas)
if set to an existing file, the groups listed in this file will be scanned (each group name must be on a separate line)
if set to 'ALL' all available groups on the usenet server will be scanned
if set to 'BINARIES' all available alt.binaries.* groups on the usenet server will be scanned`)
	flag.StringVar(&path, "path", conf.Path, "the path where the NZB file will be saved to")
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

	// force user to enter header if not already done
	for headerToSearch == "" {
		fmt.Print("Enter header to search for: ")
		headerToSearch = inputReader()
	}

	// force user to input groups if not already done
	for len(groups) == 0 {
		var err error
		if groupsFlag != "" {
			err = scanGroups(groupsFlag)
		} else {
			fmt.Print("Enter group name(s) to search in: ")
			err = scanGroups(inputReader())
		}
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}

	// force user to input date if not already done
	for {
		if date == "" {
			fmt.Print("Enter the date when the header was posted (DD.MM.YYYY or YYYY-MM-dd): ")
			date = strings.TrimSpace(inputReader())
		}
		d, err := time.Parse("02.01.2006", date)
		if err != nil {
			d, err = time.Parse("2006-01-02", date)
			if err != nil {
				fmt.Printf("Error parsing date '%s': %s\n", date, err)
				continue
			}
		}
		postDateUnix = d.Add(24 * time.Hour).Unix() // add a day for security, i.e. if it was posted before upload was finished
		break
	}

	// force user to enter search range
	days = conf.Days
	for days == 0 {
		var input string
		fmt.Print("Enter the amount of days to search back: ")
		input = strings.TrimSpace(inputReader())
		result, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Error parsing input '%s': %s\n", input, err)
		} else {
			days = result + 1 // add back the day which was added above for security to have full length of back search
		}
	}

	// set path
	if path == "" {
		path = "./"
	}
	if _, err := os.Stat(path); err != nil {
		fmt.Printf("Error for path '%s': %v\n", path, err)
		os.Exit(1)
	}
	if verbose {
		fmt.Printf("Setting path for NZB files to: %s\n", path)
	}
	conf.Path = path
}

func inputReader() string {
	reader := bufio.NewScanner(os.Stdin)
	for reader.Scan() {
		return strings.TrimSpace(reader.Text())
	}
	fmt.Printf("Error reading data: %v\n", reader.Err())
	return ""
}
