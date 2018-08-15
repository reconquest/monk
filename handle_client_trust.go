package main

import (
	"fmt"
)

func handleClientTrust(client *Client, id string) error {
	var peer PacketPeer
	err := client.Query(PacketTrustPeer{ID: id}, &peer)
	if err != nil {
		return err
	}

	fmt.Printf(
		"Monk %s %s (%s) is trusted now.\n",
		peer.Machine,
		peer.Fingerprint,
		peer.IP,
	)

	return nil
}
