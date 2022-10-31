package main

import (
	"os"
	"strconv"
)

func (c *config) getTotalRAM() (int, error) {
	var out int

	memInfo, err := c.getMeminfo()
	if err != nil {
		return out, err
	}

	matches := c.regex.memTotal.FindStringSubmatch(memInfo)

	ramInKb, err := strconv.Atoi(matches[1])
	if err != nil {
		return out, err
	}

	return ramInKb / 1000, nil
}

func (c *config) getUsedRAM() (int, error) {
	var out int

	memInfo, err := c.getMeminfo()
	if err != nil {
		return out, err
	}

	matches := c.regex.memAvailable.FindStringSubmatch(memInfo)

	ramInKb, err := strconv.Atoi(matches[1])
	if err != nil {
		return out, err
	}

	return ramInKb / 1000, nil
}

func (c *config) getMeminfo() (string, error) {
	contents, err := os.ReadFile("/proc/meminfo")
	if err != nil {
		return "", err
	}

	return string(contents), nil
}
