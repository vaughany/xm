# xm

eXtreme Monitoring (lol) logs Linux load averages and memory use, and probably more, given time.

## Use

Any of:

* download and run the binary with `./xm`
* (if you hvae Go installed) clone the repo and `go run .`

## How it works

XM will create a log file called `xm.log` in the folder it's run from.  All output is sent to `stdout` and the log file.  The log file is created or truncated on program startup.

If the load average exceeds the number of processors, a record of load average and memory use is logged.

If the in-use memory exceeds 50% of the total memory, a record of load average and memory use is logged.

## History

**v0.0.1:** initial release. Logs load average and ram use to `stdout` and a log file.
