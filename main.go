package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
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

}

func formHandler(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		fmt.Fprintf(w, "ParseForm() err: %v", err)
		return
	}
	//fmt.Fprintf(w, "POST request successful")

	header = r.FormValue("header")
	scanGroups(r.FormValue("groups"))
	var input string
	input = strings.TrimSpace(r.FormValue("date"))
	if input != "" {
		dateRegex := regexp.MustCompile(`[0-3]\d\.[0-1]\d\.(?:19|20)\d\d`)
		if match := dateRegex.FindStringIndex(input); match != nil {
			date, err := time.Parse("02.01.2006", input)
			if err != nil {
				fmt.Fprintf(w, "Error parsing date '%s': %s\n", input, err)
			} else {
				postDateUnix = date.Unix() + 60*60*24 // add a day for security, i.e. if it was posted befor upload was finished
			}
		} else {
			fmt.Fprintf(w, "Error parsing date '%s': Date does not have correct format DD.MM.YYYY!\n", input)
		}
	}
	input = strings.TrimSpace(r.FormValue("days"))
	result, _ := strconv.Atoi(input)
	days = result + 1

	// old main
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
	fmt.Fprintf(w, "A total of %d messages were processed in %v (%d Messages/s)\n", counter, duration, int(perSecond))

}

func init() {

	// load configuration
	if err := loadConfig(); err != nil {
		fmt.Println("Fatal error while loading configuration file!")
		os.Exit(1)
	}

	// flags
	pathFlag := flag.String("path", conf.Path, "the path where the NZB file will be saved to")
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

	// fire up webserver
	fileServer := http.FileServer(http.Dir("./static"))
	http.Handle("/", fileServer)
	http.HandleFunc("/index", formHandler)

	fmt.Printf("Starting server at port 8080\n")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}

}
