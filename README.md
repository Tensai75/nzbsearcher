# nzbsearcher
 "Proof of Concept" for an NZB search engine that searches for a header directly on the Usenet server and then generates the NZB files.
 
 The programme searches in each specified newsgroup for a configurable number of days backwards from a specified date.
 So if the approximate date the file was posted on Usenet and the newsgroup are known (e.g. from an Usenet forum post), there is a good chance that this program will find the messages and generate the NZB file for download.
 The program is mainly intended to search for very old Usenet posts that are still available on the Omicron backbone but have not been indexed by the known Usenet indexers.

### Installation
 Download the zip file for the desired operating system from the release page and unzip it.
 
 Execute the programme 'nzbsearcher' in a command line. The first time you run it, it will create the default configuration file "config.yml" in the same directory as the executable file.
 Open the "config.yml" and adjust the settings under "Server" according to the settings of your Usenet account. The other settings can be adjusted according to your own requirements. There are some explanations in the comments in the configuration file.

### How to use
 Execute the programme 'nzbsearcher' in a command line. You will be prompted for the header you are looking for and the date you assume the file was posted on Usenet.
 If no default information is given in the configuration file, you will also be prompted to enter the full name of the newsgroup(s) to search in and the number of days you want the program to search backwards from the specified date. If several newsgroups are to be searched, the names must be separated by commas. "alt.binaries." can be abbreviated to "a.b." when entered.
 
 The program will then search all messages in the newsgroup(s) within the specified time period and search the subjects for the specified header. If messages are found, the information is collected and then stored in a corresponding NZB file, either in the same directory as the executable file or in the path specified by the path setting.
 
 All settings in the conf file can also be specified as command line parameters and will then override the config settings. Further information can be found by specifying the `-help` parameter.

### To do
 The parsing of the subject should be improved significantly to better take into account all the very different subject formats used for file posts.

### Change log

#### v0.0.3
 - allow a.b. instead of alt.binaries.
 - show parsing erros only with verbose output (parameter: -verbose)
 - improved parsing of subject 

#### v0.0.2
 - fix to have accurate dates

#### v0.0.1
 - first public version
 