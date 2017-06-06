package updater

import (
        "github.com/potix/pdns-record-updater/contexter"
        "github.com/potix/pdns-record-updater/api/client"
)

// Updater is updater
type Updater struct {
	client *client.Client
	updaterContext *contexter.Updater
}

// New is create updater
func New(updaterContext *contexter.Updater, client *client.Client) (*Updater) {
        return &Updater {
                client:    client,
                updaterContext: updaterContext,
        }
}
