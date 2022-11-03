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

* if the 1-minute load average exceeds the number of processors
* if the used memory exceeds 50% of the total memory
* if the number of active Phusion Passenger workers equals or exceeds 10, or is less than 5
* if the number of queued Phusion Passenger requests exceeds 0


## Logfile

This is a sample of the generated logfile.

```
XM v1.0.5 startup on btsdigital-devops: Thu, 03 Nov 2022 00:52:15 GMT.

CPUs: 8. Any load above this will be recorded.
RAM: 16098 Mb. Less than 50% available will be recorded.
Passenger: 30 workers. Less than 5 used, or 10 or more used, will be recorded.
Passenger: queued requests. More than 0 will be recorded.
---
Thu, 03 Nov 2022 00:57:16 GMT: 4.89 3.97 2.89; 7545 / 16426 Mb; 18 / 30 workers; 1 in queue [CPU RAM WRK REQ]
```


## History

* **v0.0.1:** initial release. Logs load average and ram use to `stdout` and a log file.
* **v0.0.2:** added Phusion Passenger worker watching (uses `passenger-status`, which requires elevated privileges to get details out of). Also, the entry in the log is suffixed with whichever text triggered it: `CPU` for load average, `RAM` for memory, `WRK` for Passenger workers.
* **v0.0.3:** replaced reading /proc/cpuinfo with runtime.NumCPU(); added hostname to log file; added an hourly 'keep-alive' timestamp to the log file for clarity.
* **v0.0.4:** created per-day logfile name; modified Passenger workers to log 10 or more used; modified CPU use to monitor only the recent minute average, not 5 or 15 minute averages; code to handle Ctrl-C nicely.
* **v0.0.5:** version flag; passed output from Phusion Passenger to functions that require it, meaning we call `passenger-status` once for multiple Passenger tests; test for queued Passenger requests.
