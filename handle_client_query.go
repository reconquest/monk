package main

import (
	"fmt"
	"os"
	"text/tabwriter"
)

func handleClientQuery(client *Client, withFingerprint bool) error {
	var peers PacketPeers
	err := client.Query(PacketQueryPeers{}, &peers)
	if err != nil {
		return  err
	}

	formatting := "%s\t%s\t%s\n"
	if withFingerprint {
		formatting = "%s\t%s\t%s\t%s\n"
	}

	writer := tabwriter.NewWriter(os.Stdout, 0, 1, 2, ' ', 0)
	for _, peer := range peers {
		values := []interface{}{
			peer.Machine,
			peer.IP,
		}

		if withFingerprint {
			values = append(values, peer.Fingerprint)
		}

		lastSeen := peer.LastSeen.Format("2006-01-02T15:04:05")

		if peer.Trusted {
			lastSeen += " [trusted]"
		}

		values = append(values, lastSeen)

		fmt.Fprintf(writer, formatting, values...)
	}

	err = writer.Flush()
	if err != nil {
		return err
	}

	return nil
}
