# tracerlogger
Reads data from a connected EPsolar Tracer*BN solar charge controller every 5 seconds and saves it to a SQLite database.

## Requirements
* Go (there are no binary releases yet)

## Installation
Currently there are no binaries to download. Instead you have to install Go,
http://www.golang.org, and run:
```
go install github.com/spagettikod/tracerlogger
```

## Usage
Run `tracelogger` with the following required parameters:
* p - name of the serial port to which the EPsolar Tracer is connected to
* db - SQLite database file

```
tracelogger -p /dev/ttyXRUSB0 -db /home/bali/tracerlogger.db
```
