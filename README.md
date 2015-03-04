# tracerlogger
Reads data from a connected EPsolar Tracer*BN solar charge controller every 5 seconds and saves it to AWS SimpleDB.

## Requirements
* Go (there are no binary releases yet)
* Account at Amazon Web Service to access SimpleDB. The access key and secret key is required.

## Installation
Currently there are no binaries to download. Instead you have to install Go,
http://www.golang.org, and run the following commands.

```
go get github.com/spagettikod/tracerlogger
go install github.com/spagettikod/tracerlogger
```

## Usage
Run `tracelogger` with the following required parameters:
* p - name of the serial port to which the EPsolar Tracer is connected to
* a - Amazon Web Services Access Key
* s - Amazon Web Services Secret Key

```
tracelogger -p COM16 -a AWS_ACCESS_KEY -s AWS_SECRET_KEY
```
