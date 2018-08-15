package main

import (
	"io/ioutil"
	"math/rand"
	"strings"
	"time"

	"github.com/reconquest/karma-go"
)

func init() {
	rand.Seed(time.Now().UnixNano())
}

func getMachineID() (string, error) {
	contents, err := ioutil.ReadFile("/etc/machine-id")
	if err != nil {
		return "", karma.Format(
			err,
			"unable to read machine-id",
		)
	}

	return strings.TrimSpace(string(contents)), nil
}
