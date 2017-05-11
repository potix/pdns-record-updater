package main
import (
        "github.com/potix/pdns-record-updater/configurator"
        "github.com/potix/pdns-record-updater/collector"
        "github.com/potix/pdns-record-updater/initializer"
        "github.com/potix/pdns-record-updater/updater"
	"flag"
)

func main() {
	configPath := flag.String("config", "/etc/pdns-record-updater.conf", "config file path")
	configurator := configurator.New(configPath)
	config, err := configurator.Load()
	if err != nil {
		os.Printf("%s", err)
		os.Exit(1)
	}
	collector := collector.New(config)
	err := collector.Run()
	if err != nil {
		os.Printf("%s", err)
		os.Exit(1)
	}
	initializer := initializer.New(config, colletor)
	updater := initializer.New(config, collector)
	initializer.initialize()
	for {
		updator.update()
	}
}

