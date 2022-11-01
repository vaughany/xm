package main

import (
	"fmt"
	"os"
)

func (c *config) getLoadAverages() (float64, float64, float64, error) {
	var (
		la1  float64
		la5  float64
		la15 float64
	)

	fileContents, err := os.ReadFile("/proc/loadavg")
	if err != nil {
		return la1, la5, la15, err
	}

	fmt.Sscanf(string(fileContents), "%f %f %f", &la1, &la5, &la15)

	return la1, la5, la15, nil
}
