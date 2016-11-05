package main

func assert(statement bool, message string) {
	if !statement {
		panic("assertion failure: " + message)
	}
}
