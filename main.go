package main

import (
	"fmt"
	"log"
	"os"
	"regexp"
	_ "runtime/cgo" // Required for cgo / static compilation for Ubuntu 18.04.
	"time"
)

type config struct {
	logFile           string
	logFileHandle     *os.File
	totalCPUs         int
	totalRAM          int
	regexMemAvailable *regexp.Regexp
	regexMemTotal     *regexp.Regexp
}

func main() {
	var (
		cfg config
		err error
	)

	cfg.logFile = "xm.log"
	cfg.regexMemAvailable = regexp.MustCompile(`MemAvailable: {1,}([0-9]{1,}) kB`)
	cfg.regexMemTotal = regexp.MustCompile(`MemTotal: {1,}([0-9]{1,}) kB`)

	// Creates a new file, and wipes the existing file.
	cfg.logFileHandle, err = os.Create(cfg.logFile)
	if err != nil {
		log.Fatal(err)
	}
	defer cfg.logFileHandle.Close()

	cfg.recordIt(fmt.Sprintf("App startup: %s.", time.Now().Format(time.RFC1123)))

	cfg.totalCPUs, err = cfg.getNumCPUs()
	if err != nil {
		fmt.Println(err)
	}
	cfg.recordIt(fmt.Sprintf("CPUs: %d. Any load above this will be recorded.", cfg.totalCPUs))

	cfg.totalRAM, err = cfg.getTotalRAM()
	if err != nil {
		fmt.Println(err)
	}
	cfg.recordIt(fmt.Sprintf("RAM: %d Mb. More than 50%% of this used will be recorded.\n", cfg.totalRAM))

	for {
		logThis := false

		// Load average testing.
		la1, la5, la15, err := cfg.getLoadAverages()
		if err != nil {
			fmt.Println(err)
		}
		if la1 > float64(cfg.totalCPUs) || la5 > float64(cfg.totalCPUs) || la15 > float64(cfg.totalCPUs) {
			logThis = true
		}

		availableRAM, err := cfg.getUsedRAM()
		if err != nil {
			fmt.Println(err)
		}
		if availableRAM < (cfg.totalRAM / 2) {
			logThis = true
		}

		// Log it all if required.
		if logThis {
			record := fmt.Sprintf("%s: %.2f %.2f %.2f; %d / %d Mb", time.Now().Format(time.RFC1123), la1, la5, la15, availableRAM, cfg.totalRAM)

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
	_, err := c.logFileHandle.WriteString(record + "\n")
	if err != nil {
		log.Panic(err)
	}
}
