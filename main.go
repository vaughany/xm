package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	_ "runtime/cgo" // Required for cgo / static compilation for Ubuntu 18.04.
	"strings"
	"time"
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
	}
	version string
}

func main() {
	var (
		cfg config
		err error
	)

	cfg.log.file = "xm.log"
	cfg.regex.memAvailable = regexp.MustCompile(`MemAvailable: {1,}([0-9]{1,}) kB`)
	cfg.regex.memTotal = regexp.MustCompile(`MemTotal: {1,}([0-9]{1,}) kB`)
	cfg.regex.passengerMaxPoolSize = regexp.MustCompile(`Max pool size {1,}: ([0-9]{1,})`)
	cfg.regex.passengerProcesses = regexp.MustCompile(`Processes {1,}: ([0-9]{1,})`)
	cfg.version = "1.0.2"

	flag.StringVar(&cfg.log.file, "logfile", cfg.log.file, "logfile path and name")
	flag.Parse()

	// Creates a new log file, wiping the existing file if it exists.
	cfg.log.fileHandle, err = os.Create(cfg.log.file)
	if err != nil {
		log.Fatal(err)
	}
	defer cfg.log.fileHandle.Close()

	cfg.recordIt(fmt.Sprintf("XM v%s startup: %s.\n", cfg.version, time.Now().Format(time.RFC1123)))

	cfg.total.CPUs, err = cfg.getNumCPUs()
	if err != nil {
		fmt.Println(err)
	}
	cfg.recordIt(fmt.Sprintf("CPUs: %d. Any load above this will be recorded.", cfg.total.CPUs))

	cfg.total.RAM, err = cfg.getTotalRAM()
	if err != nil {
		fmt.Println(err)
	}
	cfg.recordIt(fmt.Sprintf("RAM: %d Mb. Less than 50%% of this available will be recorded.", cfg.total.RAM))

	cfg.total.passengerPool, err = cfg.getPassengerMaxPoolSize()
	if err != nil {
		fmt.Println(err)
	}
	cfg.recordIt(fmt.Sprintf("Passenger: %d workers. More than 50%% of these used will be recorded.", cfg.total.passengerPool))

	cfg.recordIt("---")

	for {
		logThis := false
		triggers := ""

		// Load average testing.
		la1, la5, la15, err := cfg.getLoadAverages()
		if err != nil {
			fmt.Println(err)
		}
		if la1 > float64(cfg.total.CPUs) || la5 > float64(cfg.total.CPUs) || la15 > float64(cfg.total.CPUs) {
			logThis = true
			triggers += "CPU "
		}

		// RAM use testing.
		availableRAM, err := cfg.getUsedRAM()
		if err != nil {
			fmt.Println(err)
		}
		if availableRAM < (cfg.total.RAM / 2) {
			logThis = true
			triggers += "RAM "
		}

		// Passenger processes.
		passengerProcesses, err := cfg.getPassengerProcesses()
		if err != nil {
			fmt.Println(err)
		}
		if passengerProcesses > (cfg.total.passengerPool / 2) {
			logThis = true
			triggers += "WRK "
		}

		// Log it all if required.
		if logThis {
			record := fmt.Sprintf("%s: %.2f %.2f %.2f; %d / %d Mb; %d / %d workers [%s]", time.Now().Format(time.RFC1123),
				la1, la5, la15, availableRAM, cfg.total.RAM, passengerProcesses, cfg.total.passengerPool, strings.Trim(triggers, " "))

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
