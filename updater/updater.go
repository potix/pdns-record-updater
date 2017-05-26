package updater

// New is create updater
func New(updater *contexter.Updater, client client.Client) (*Updater) {
        return &Updater {
                client:    client,
                pnsServer: Updater.PdnsServer,
                pnsApiKey: Updater.PdnsApiKey,
        }
}
