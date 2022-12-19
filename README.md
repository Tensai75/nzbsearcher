# nzbsearcher
 "Proof of Concept" for an NZB search engine that searches for a header directly on the Usenet server and then generates the NZB files.
 
 The program searches in each specified newsgroup for a configurable number of days backwards from a specified date.
 So if the approximate date the file was posted on Usenet and the newsgroup are known (e.g. from an Usenet forum post), there is a good chance that this program will find the messages and generate the NZB file for download.
 The program is mainly intended to search for very old Usenet posts that are still available on the Omicron backbone but have not been indexed by the known Usenet indexers.

## About this fork
This fork tries to convert the original command line tool into a web based, dockered version.

![Alt text](/doc/screenshot.png?raw=true "web interface screenshot")

It is far from perfect and I advice you to use it with caution.

### Docker instructions
Copy the config.yml.example to config.yml. Open the "config.yml" and adjust the settings under "Server" according to the settings of your Usenet account. The other settings can be adjusted according to your own requirements. There are some explanations in the comments in the configuration file.

Build the image with:
```
docker build --tag nzbsearcher .
```

Run the container with:
```
docker run -d \
  --publish 8080:8080 \
  -v ./config.yml:/app/config.yml \
  -v ./nzbdrop:/nzb \
  --name nzbsearcher \
  nzbsearcher
```

Explanation of the parameters:
  - --publish 8080:8080 makes sure that you can reach the container at the given port
  - -v ./config.yml:/app/config.yml makes your configuration available to the container
  - -v ./nzbdrop:/app/nzb links nzbsearcher with a usenet downloader

### How to use
Run the docker container as described above. Then navigate to http://localhost:8080 with your browser to open the web interface.

# Changelog
  - added a web interface
  - added docker capabilities
  - added [nzblnk parser from hnz101](https://github.com/Tensai75/nzbsearcher/pull/1/commits/a3299e89d8481c98f731c13ebd0e6df08f55d064)
  
# Credits
All credits go to tensai75: [Link to Repo](https://github.com/Tensai75/nzbsearcher)
CSS theme source: https://codepen.io/Godex/pen/DLgQbg