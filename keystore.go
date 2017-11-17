package main

import "fmt"

//Mote promotes or demotes this server
func (k *KeyStore) Mote(master bool) error {
	entries := k.serverGetter.getServers()
	for _, entry := range entries {
		meta := k.serverStatusGetter.getStatus(entry)

		if meta.GetVersion() > k.Meta.GetVersion() {
			return fmt.Errorf("We're too behind to be master")
		}
	}

	return nil
}
