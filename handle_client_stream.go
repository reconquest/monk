package main

import (
	"io"
	"os"

	"github.com/reconquest/karma-go"
)

func handleClientStream(
	client *Client,
	machine string,
) error {
	var stream PacketStream
	err := client.Query(PacketQueryStream{ID: machine}, &stream)
	if err != nil {
		return karma.Format(
			err,
			"unable to start chat",
		)
	}

	pipe, err := ConnectPipe(stream.Pipe)
	if err != nil {
		return karma.Format(
			err,
			"unable to connect to %s", stream.Pipe,
		)
	}

	communicate(
		pipe, os.Stdout,
		os.Stdin, pipe,
	)

	return nil
}

func communicate(
	outputFrom io.ReadCloser,
	outputTo io.WriteCloser,
	inputFrom io.ReadCloser,
	inputTo io.WriteCloser,
) {
	close := func() {
		outputFrom.Close()
		outputTo.Close()
		inputFrom.Close()
		inputTo.Close()
	}

	closed := false
	withWait(
		func() {
			_, err := io.Copy(outputTo, outputFrom)
			if err != nil {
				if closed {
					return
				}

				errorh(err, "unable to io/copy to output from pipe")
			}

			closed = true
			close()
		},
		func() {
			_, err := io.Copy(inputTo, inputFrom)
			if err != nil {
				errorh(err, "unable to io/copy from input to pipe")
			}

			closed = true
			close()
		},
	)
}
