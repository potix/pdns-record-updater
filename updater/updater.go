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

// Start is start
func (u Updater) Start() {

}

// Stop is stop
func (u Updater) Stop() {

}

// New is create updater
func New(updaterContext *contexter.Updater, client *client.Client) (*Updater) {
        return &Updater {
                client:    client,
                updaterContext: updaterContext,
        }
}
