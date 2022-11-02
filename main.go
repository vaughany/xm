package main

import (
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
	startup time.Time
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

	cfg.log.file = fmt.Sprintf("xm_%s.log", time.Now().Format("2006-01-02"))
	cfg.regex.memAvailable = regexp.MustCompile(`MemAvailable: {1,}([0-9]{1,}) kB`)
	cfg.regex.memTotal = regexp.MustCompile(`MemTotal: {1,}([0-9]{1,}) kB`)
	cfg.regex.passengerMaxPoolSize = regexp.MustCompile(`Max pool size {1,}: ([0-9]{1,})`)
	cfg.regex.passengerProcesses = regexp.MustCompile(`Processes {1,}: ([0-9]{1,})`)
	cfg.version = "1.0.4"
	cfg.startup = time.Now()

	flag.StringVar(&cfg.log.file, "logfile", cfg.log.file, "logfile path and name")
	flag.Parse()

	// Creates a new log file, wiping the existing file if it exists.
	cfg.log.fileHandle, err = os.Create(cfg.log.file)
	if err != nil {
		log.Fatal(err)
	}
	defer cfg.log.fileHandle.Close()

	hostname, err := os.Hostname()
	if err != nil {
		log.Fatal(err)
	}
	cfg.recordIt(fmt.Sprintf("XM v%s startup on %s: %s.\n", cfg.version, hostname, time.Now().Format(time.RFC1123)))

	cfg.total.CPUs = runtime.NumCPU()
	cfg.recordIt(fmt.Sprintf("CPUs: %d. Any load above this will be recorded.", cfg.total.CPUs))

	cfg.total.RAM, err = cfg.getTotalRAM()
	if err != nil {
		fmt.Println(err)
	}
	cfg.recordIt(fmt.Sprintf("RAM: %d Mb. Less than 50%% available will be recorded.", cfg.total.RAM))

	cfg.total.passengerPool, err = cfg.getPassengerMaxPoolSize()
	if err != nil {
		fmt.Println(err)
	}
	cfg.recordIt(fmt.Sprintf("Passenger: %d workers. Less than 5 used, or 10 or more used, will be recorded.", cfg.total.passengerPool))

	cfg.recordIt("---")

	go cfg.keepAlive()

	for {
		logThis := false
		triggers := ""

		// Load average testing.
		la1, la5, la15, err := cfg.getLoadAverages()
		if err != nil {
			fmt.Println(err)
		}
		if la1 > float64(cfg.total.CPUs) {
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
		if passengerProcesses < 5 || passengerProcesses >= 10 {
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

func (c *config) keepAlive() {
	for {
		time.Sleep(time.Hour)

		c.recordIt(fmt.Sprintf("App has been running for %s.", time.Since(c.startup).Round(time.Minute)))
	}
}

func (c *config) cleanup() {
	c.recordIt(fmt.Sprintf("%s: app shutdown. Bye!", time.Now().Format(time.RFC1123)))
}
