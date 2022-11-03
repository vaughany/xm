package main

import (
	"os/exec"
	"strconv"
)

func (c *config) getPassengerMaxPoolSize() (int, error) {
	var maxPoolSize int

	passenger, err := c.getPassengerOutput()
	if err != nil {
		return 0, err
	}

	matches := c.regex.passengerMaxPoolSize.FindStringSubmatch(passenger)

	maxPoolSize, err = strconv.Atoi(matches[1])
	if err != nil {
		return maxPoolSize, err
	}

	return maxPoolSize, nil
}

func (c *config) getPassengerProcesses(passenger string) (int, error) {
	var processes int

	matches := c.regex.passengerProcesses.FindStringSubmatch(passenger)

	processes, err := strconv.Atoi(matches[1])
	if err != nil {
		return processes, err
	}

	return processes, nil
}

func (c *config) getPassengerRequests(passenger string) (int, error) {
	var requests int

	matches := c.regex.passengerRequests.FindStringSubmatch(passenger)

	requests, err := strconv.Atoi(matches[1])
	if err != nil {
		return requests, err
	}

	return requests, nil
}

func (c *config) getPassengerOutput() (string, error) {
	cmd := exec.Command("passenger-status")
	stdout, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return string(stdout), nil
}
