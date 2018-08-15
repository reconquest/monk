package main

import (
	"fmt"
	"strconv"
)

func argInt(args map[string]interface{}, flag string) int {
	defer func() {
		tears := recover()
		if tears != nil {
			panic(fmt.Sprintf("invalid docopt for %s", flag))
		}
	}()

	number, err := strconv.Atoi(args[flag].(string))
	if err != nil {
		fatalh(err, "unable to parse %s value as integer", flag)
	}

	return number
}
