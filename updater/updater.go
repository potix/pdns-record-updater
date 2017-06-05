package updater

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
