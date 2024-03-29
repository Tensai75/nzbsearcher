package main

func defaultConfig() string {
	return `# Usenet server settings
Server:
  Host: "news-eu.newshosting.com"
  Port: 119
  SSL: false
  User: ""
  Password: ""
  Connections: 50

# Groups to be scanned
# Possible values:
# - group names separated by a comma, e.g. "alt.binaries.u-4all,alt.binaries.highspeed"
# - path to an existing text file with each group name on a separate line, e.g. "./groups.txt"
# - "ALL" -> all available groups on the usenet server will be scanned
# - "BINARIES" -> all available alt.binaries.* groups on the usenet server will be scanned
# If left empty or commented out, the program will ask for the group names
Groups: ""

# Path to the folder where the NZB file will be saved.
# The path must exist. If left empty or commented out or if the path does not exist
# the file will be saved in the program's folder
Path: ""

# Number of days the search will go back from the date the header was posted
# If left empty or commented out, the program will ask for the amount of days
Days:

# Number of groups to scan in parallel
ParallelScans: 200

# Number of message headers to retrieve in one header overview request
# Be careful when increasing this value as it will increase memory consumption without significantly increasing processing speed
# Step x ParallelScans is the max number of message headers held in memory at one time, with each header consuming around 1 Kilobit of memory
Step: 20000

# If set to true, additional information will be outputted
Verbose: false`
}
