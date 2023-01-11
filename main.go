package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"regexp"
	"runtime"
	_ "runtime/cgo" // Required for cgo / static compilation for Ubuntu 18.04.
	"strings"
	"syscall"
	"time"

	"github.com/TwiN/go-color"
)

type config struct {
	log struct {
		file       string
		fileHandle *os.File
	}
	total struct {
		CPUs          int
		RAM           int
		passengerPool int
	}
	regex struct {
		memAvailable         *regexp.Regexp
		memTotal             *regexp.Regexp
		passengerMaxPoolSize *regexp.Regexp
		passengerProcesses   *regexp.Regexp
		passengerRequests    *regexp.Regexp
	}
	noPassenger bool
	version     string
	startup     time.Time
}

func main() {
	var (
		cfg config
		err error
	)

	sig := make(chan os.Signal)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sig
		cfg.cleanup()
		os.Exit(0)
	}()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}

	cfg.log.file = fmt.Sprintf("xm_%s_%s.log", time.Now().Format("2006-01-02"), hostname)
	cfg.regex.memAvailable = regexp.MustCompile(`MemAvailable: {1,}([0-9]{1,}) kB`)
	cfg.regex.memTotal = regexp.MustCompile(`MemTotal: {1,}([0-9]{1,}) kB`)
	cfg.regex.passengerMaxPoolSize = regexp.MustCompile(`Max pool size {1,}: ([0-9]{1,})`)
	cfg.regex.passengerProcesses = regexp.MustCompile(`Processes {1,}: ([0-9]{1,})`)
	cfg.regex.passengerRequests = regexp.MustCompile(`Requests in queue: ([0-9]{1,})`)
	cfg.version = "0.0.7"
	cfg.startup = time.Now()

	flag.StringVar(&cfg.log.file, "logfile", cfg.log.file, "logfile path and name")
	flag.BoolVar(&cfg.noPassenger, "nopassenger", cfg.noPassenger, "does not run Phusion Passenger checks")
	version := flag.Bool("v", false, "gets version info and exits")
	flag.Parse()

	if *version {
		fmt.Printf("XM v%s.\n", cfg.version)
		os.Exit(0)
	}

	// Creates a new log file, wiping the existing file if it exists.
	cfg.log.fileHandle, err = os.Create(cfg.log.file)
	if err != nil {
		log.Fatal(err)
	}
	defer cfg.log.fileHandle.Close()

	cfg.recordIt(fmt.Sprintf("XM v%s startup on %s: %s.\n", cfg.version, hostname, time.Now().Format(time.RFC1123)))

	cfg.total.CPUs = runtime.NumCPU()
	cfg.recordIt(fmt.Sprintf("CPUs: %d. Any load above this will be recorded.", cfg.total.CPUs))

	cfg.total.RAM, err = cfg.getTotalRAM()
	if err != nil {
		fmt.Println(err)
	}
	cfg.recordIt(fmt.Sprintf("RAM: %d Mb. Less than 50%% available will be recorded.", cfg.total.RAM))

	// Checking for Phusion Passenger's presence.
	_, err = os.Stat("/usr/sbin/passenger-status")
	if err != nil && errors.Is(err, os.ErrNotExist) {
		cfg.noPassenger = true
	}
	if cfg.noPassenger {
		cfg.recordIt("Not checking Phusion Passenger as it is not present or the --nopassenger flag was set.")
	} else {
		cfg.total.passengerPool, err = cfg.getPassengerMaxPoolSize()
		if err != nil {
			fmt.Println(err)
		}
		cfg.recordIt(fmt.Sprintf("Passenger: %d workers. Less than 5 used, or 10 or more used, will be recorded.", cfg.total.passengerPool))
		cfg.recordIt("Passenger: queued requests. More than 0 will be recorded.")
	}

	cfg.recordIt("---")

	go cfg.keepAlive()

	for {
		var (
			logThis            = false
			triggers           = ""
			passengerProcesses = 0
			passengerRequests  = 0
			laOutput           = ""
			ramOutput          = ""
			procOutput         = "0"
			reqOutput          = "0"
		)

		// Load average testing.
		la1, la5, la15, err := cfg.getLoadAverages()
		if err != nil {
			fmt.Println(err)
		}
		laOutput = fmt.Sprintf("%.2f %.2f %.2f", la1, la5, la15)
		if la1 > float64(cfg.total.CPUs) {
			logThis = true
			laOutput = color.Colorize(color.Red, laOutput)
			triggers += color.Colorize(color.Red, "CPU") + " "
		}

		// RAM use testing.
		availableRAM, err := cfg.getUsedRAM()
		if err != nil {
			fmt.Println(err)
		}
		ramOutput = fmt.Sprintf("%d", availableRAM)
		if availableRAM < (cfg.total.RAM / 2) {
			logThis = true
			ramOutput = color.Colorize(color.Yellow, ramOutput)
			triggers += color.Colorize(color.Yellow, "RAM") + " "
		}

		// Passenger things, if possible and not turned off.
		if !cfg.noPassenger {
			// Get the output from passenger once for both 'processes' and 'requests' checks.
			passengerOutput, err := cfg.getPassengerOutput()
			if err != nil {
				fmt.Println(err)
			}

			// Passenger processes.
			passengerProcesses, err = cfg.getPassengerProcesses(passengerOutput)
			if err != nil {
				fmt.Println(err)
			}
			procOutput = fmt.Sprintf("%d", passengerProcesses)
			if passengerProcesses < 5 || passengerProcesses >= 10 {
				logThis = true
				procOutput = color.Colorize(color.Purple, procOutput)
				triggers += color.Colorize(color.Purple, "WRK") + " "
			}

			// Passenger requests in queue.
			passengerRequests, err = cfg.getPassengerRequests(passengerOutput)
			if err != nil {
				fmt.Println(err)
			}
			reqOutput = fmt.Sprintf("%d", passengerRequests)
			if passengerRequests > 0 {
				logThis = true
				reqOutput = color.Colorize(color.Cyan, reqOutput)
				triggers += color.Colorize(color.Purple, "REQ") + " "
			}
		}

		// Log it all if required.
		if logThis {
			var record string
			if cfg.noPassenger {
				record = fmt.Sprintf("%s: %s; %s / %d Mb [%s]", time.Now().Format(time.RFC1123),
					laOutput, ramOutput, cfg.total.RAM, strings.Trim(triggers, " "))

			} else {
				record = fmt.Sprintf("%s: %s; %s / %d Mb; %s / %d workers; %s in queue [%s]", time.Now().Format(time.RFC1123),
					laOutput, ramOutput, cfg.total.RAM, procOutput, cfg.total.passengerPool, reqOutput, strings.Trim(triggers, " "))
			}

			cfg.recordIt(record)
		}

		time.Sleep(5 * time.Second)
	}
}

func (c *config) recordIt(record string) {
	fmt.Println(record)
	c.writeToDisk(record)
}

func (c *config) writeToDisk(record string) {
	_, err := c.log.fileHandle.WriteString(record + "\n")
	if err != nil {
		log.Panic(err)
	}
}

func (c *config) keepAlive() {
	for {
		time.Sleep(time.Hour)

		c.recordIt(fmt.Sprintf("App has been running for %s.", time.Since(c.startup).Round(time.Minute)))
	}
}

func (c *config) cleanup() {
	c.recordIt(fmt.Sprintf("%s: app shutdown. Bye!", time.Now().Format(time.RFC1123)))
}
