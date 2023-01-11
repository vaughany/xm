# xm

eXtreme Monitoring (lol) logs Linux load averages and memory use, and [Phusion Passenger](https://www.phusionpassenger.com/) stats.


## Get it

Either of:

* download the [latest binary](https://github.com/vaughany/xm/releases)
* clone this repo with `git clone git@github.com:vaughany/xm.git`


## Use it

If you downloaded the binary, run it with `./xm`.  If you cloned the repo, run the code with `go run .`.

**Note:** You may need to become the root user (e.g. `su -s`) or run with `sudo` for some parts of the application to work correctly, e.g. `sudo ./xm`.

The program will run and create a logfile in the current directory, and if any trigger points are met or exceeded, a row is added to the log.  The logfile created is based on today's date in 'YYYY-MM-DD' format, as well as the computer's hostname.

**Note:** If a logfile exists with the same name, it will be truncated.

Run `./xm -h` for flags. Currently:

* `-v` to print version info then exit
* specify a different logfile with the `--logfile` flag, e.g. `./xm --logfile today.log`
* `--nopassenger` to prevent checks against Phusion Passenger.


## Workflow

This app was designed to be left running on remote servers, so I have the following workflow to achieve this.

* build the app (if you haven't downloaded the binary) using the provided 'build.sh' script: `./build.sh`
* copy the binary to the remote computer, e.g.: `scp bin/xm 10.10.10.10:/tmp/`
* log in to the remote computer, e.g.: `ssh 10.10.10.10`
* become root, e.g.: `su -` or `sudo -s`
* copy the binary to the root user's home folder, e.g.: `cp /tmp/xm .`
* run screen, e.g.: `screen`
* test-run the application: `./xm -v`
* run the application properly: `./xm` - output is sent to the logfile and stdout, so you should see some startup text
* detatch screen by pressing `Ctrl-A` then `d`, then log out of the computer with e.g.: `Ctrl-D`


## How it works

xm checks the computer's load averages, used memory, and optionally Phusion Passenger workers and requests, every five seconds. (This seems to be the frequency at which the load averages in `/etc/loadavg` are updated, despite the other details in the file being updated near-constantly.)

A line is logged to the logfile in any of these situations:

* if the 1-minute load average exceeds the number of processors
* if the used memory exceeds 50% of the total memory
* if the number of active Phusion Passenger workers equals or exceeds 10, or is less than 5
* if the number of queued Phusion Passenger requests exceeds 0

Each line in the logfile is tagged with which check was triggered.


## Logfile

This is a sample of the generated logfile.

```
XM v0.0.6 startup on ubuntu: Fri, 04 Nov 2022 22:29:24 GMT.

CPUs: 4. Any load above this will be recorded.
RAM: 32941 Mb. Less than 50% available will be recorded.
Passenger: 30 workers. Less than 5 available, or 10 or more used, will be recorded.
Passenger: queued requests. More than 0 will be recorded. 
---
```


## History

* **v0.0.1:** initial release. Logs load average and ram use to `stdout` and a log file.
* **v0.0.2:** added Phusion Passenger worker watching (uses `passenger-status`, which requires elevated privileges to get details out of). Also, the entry in the log is suffixed with whichever text triggered it: `CPU` for load average, `RAM` for memory, `WRK` for Passenger workers.
* **v0.0.3:** replaced reading /proc/cpuinfo with runtime.NumCPU(); added hostname to log file; added an hourly 'keep-alive' timestamp to the log file for clarity.
* **v0.0.4:** created per-day logfile name; modified Passenger workers to log 10 or more used; modified CPU use to monitor only the recent minute average, not 5 or 15 minute averages; code to handle Ctrl-C nicely.
* **v0.0.5:** version flag; passed output from Phusion Passenger to functions that require it, meaning we call `passenger-status` once for multiple Passenger tests; test for queued Passenger requests (tagged `REQ`).
* **v0.0.6:** added a flag to skip Passenger-related checks, and an in-code check to also set that flag if the binary's not available; added colour to the output to better show which check triggered the log.
* **v0.0.7:** Now doesn't show Passenger details (all 0's) in the output if `-nopassenger` is set.
