# Sonar CDR Submission

Gathers CDR information from a CSV file on the local host, parses, and then submits to your Sonar instance.

### Installation

This requires go installed unless you download the pre-built binary within the build directory. Currently only built for linux 64bit

###### Setup with pre-built file
Setup initial .env file. Within the directory the pre-built file is located at perform the following.

```sh
$ cp env.example .env
$ vi .env
```
Adjust the information accordingly to your setup needs.

### Development

Want to contribute? Great! Fork the repository and perform your changes and submit a PR.