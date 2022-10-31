# xm

eXtreme Monitoring (lol) logs Linux load averages and memory use, and probably more, given time.

## Use

Any of:

* download and run the binary with `./xm`
* (if you have Go installed) clone the repo and `go run .`

**Note:** you may need to become a root user or run with `sudo` for some parts of the application to work correctly.

## How it works

XM will create a log file called `xm.log` in the folder it's run from.  All output is sent to `stdout` and the log file.  The log file is created or truncated on program startup.

If any of the 1, 5 or 15-minute load averages exceeds the number of processors, a record of load average and memory use is logged.

If the in-use memory exceeds 50% of the total memory, a record of load average and memory use is logged.

If the number of active Phusion Passenger workers exceeds 50% of the total number of workers, it's logged.

## History

* **v0.0.1:** initial release. Logs load average and ram use to `stdout` and a log file.
* **v0.0.2:** added Phusion Passenger worker watching (uses `passenger-status`, which requires elevated privileges to get details out of). Also, the entry in the log is suffixed with whichever text triggered it: `CPU` for load average, `RAM` for memory, `WRK` for Passenger workers.
