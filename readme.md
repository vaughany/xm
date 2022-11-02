# xm

eXtreme Monitoring (lol) logs Linux load averages and memory use, and probably more, given time.


## Use

It is assumed that this app would be left unattended on a remote computer by using `screen` or similar.

Either of:

* download and run the binary with `./xm`
* (if you have Go installed) clone the repo and `go run .`

**Note:** you may need to become a root user or run with `sudo` for some parts of the application to work correctly.

You may specify a different logfile with the `--logfile` flag, e.g. `./xm --logfile today.log`


## How it works

XM will create a log file called `xm.log` in the folder it's run from.  All output is sent to `stdout` and the log file.  The log file is created or truncated on program startup.

A line is logged to the logfile in any of these situations:

* if any of the 1, 5 or 15-minute load averages exceeds the number of processors
* if the used memory exceeds 50% of the total memory
* if the number of active Phusion Passenger workers exceeds 50% of the total number of workers, or less than 5


## Logfile

This is a sample of the generated logfile.

```
XM v1.0.3 startup on ubuntu-thinkpad: Tue, 01 Nov 2022 20:32:00 GMT.

CPUs: 8. Any load above this will be recorded.
RAM: 16098 Mb. Less than 50% of this available will be recorded.
Passenger: 20 workers. 10 or more of these used (or less than 5 used) will be recorded.
---
Tue, 01 Nov 2022 20:32:00 GMT: 0.43 0.27 0.48; 3513 / 16098 Mb; 18 / 20 workers [RAM WRK]
```


## History

* **v0.0.1:** initial release. Logs load average and ram use to `stdout` and a log file.
* **v0.0.2:** added Phusion Passenger worker watching (uses `passenger-status`, which requires elevated privileges to get details out of). Also, the entry in the log is suffixed with whichever text triggered it: `CPU` for load average, `RAM` for memory, `WRK` for Passenger workers.
* **v0.0.3:** replaced reading /proc/cpuinfo with runtime.NumCPU(); added hostname to log file; added an hourly 'keep-alive' timestamp to the log file for clarity.
* **v0.0.4:** created per-day logfile name; modified Passenger workers to log 10 or more used; modified CPU use to monitor only the recent minute average, not 5 or 15 minute averages; code to handle Ctrl-C nicely.
