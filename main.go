package main
import (
        "github.com/potix/pdns-record-updater/configurator"
        "github.com/potix/pdns-record-updater/collector"
        "github.com/potix/pdns-record-updater/initializer"
        "github.com/potix/pdns-record-updater/updater"
        "github.com/potix/pdns-record-updater/server"
        "github.com/potix/pdns-record-updater/checker"
	"flag"
	"fmt"
)

func updater(config *configurator.Config) (err error) {
	collector := collector.New(config)
	err := collector.Run()
	if err != nil {
		return err
	}
	initializer := initializer.New(config, colletor)
	updater := initializer.New(config, collector)
	initializer.initialize()
	for {
		updator.update()
	}
	return nil
}

func checker(config *configurator.Config) (err error) {
	checker := checkerr.New(config)
	err := checker.Run()
	if err != nil {
		return err
	}
	server := server.New(config, checker)
	err := server.Start()
	if err != nil {
		return err
	}
	for {
		checker.check()
	}
	return nil
}

func main() {
	mode := flag.String("mode", nil, "run mode (updater|checker)")
	configPath := flag.String("config", "/etc/pdns-record-updater.conf", "config file path")
	configurator := configurator.New(configPath)
	config, err := configurator.Load()
	if err != nil {
		return err
	}
	if (mode == "updater") {
		err := updator(config)
	} else if (mode == "checker") {
		err := checker(config)
	} else {
		err := errors.New("unexpected run mode")
	}
	if err != nil {
		fmt.Printf("%v\n", err)
	}
}

