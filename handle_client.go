package main

func handleClient(args map[string]interface{}) {
	var (
		socketPath = args["--socket"].(string)
	)

	client := NewClient("unix", socketPath)

	// corner case: fingerprint can be obtained without communication with
	// daemon
	if !args["fingerprint"].(bool) {
		err := client.Dial()
		if err != nil {
			fatalh(
				err,
				"unable to dial to %s, is monk daemon running?",
				socketPath,
			)
		}
	}

	var err error
	switch {
	case args["peers"].(bool):
		err = handleClientQuery(
			client,
			args["--fingerprint"].(bool),
		)

	case args["stream"].(bool):
		machine, _ := args["<machine>"].(string)
		err = handleClientStream(
			client,
			machine,
		)

	case args["fingerprint"].(bool):
		err = handleClientFingerprint(args["--data-dir"].(string))

	case args["trust"].(bool):
		err = handleClientTrust(client, args["<machine>"].(string))

	default:
		panic("unexpected action")
	}

	if err != nil {
		fatalln(err)
	}
}
