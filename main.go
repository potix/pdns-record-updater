package main

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
        "github.com/potix/pdns-record-updater/configurator"
        "github.com/potix/pdns-record-updater/collector"
        "github.com/potix/pdns-record-updater/initializer"
        "github.com/potix/pdns-record-updater/updater"
        "github.com/potix/pdns-record-updater/watcher"
        "github.com/potix/pdns-record-updater/server"
	"flag"
	"fmt"
)

func updater(config *configurator.Config) (err error) {
//	collector := collector.New(config)
//	err := collector.Run()
//	if err != nil {
//		return err
//	}
//	initializer := initializer.New(config, colletor)
//	updater := initializer.New(config, collector)
//	initializer.Initialize()
//	for {
//		updator.Update()
//	}
	return nil
}

func watcher(config *configurator.Config) (err error) {
	watcher := watcher.New(config)
	err := watcher.Run()
	if err != nil {
		return err
	}
	server := server.New(config, watcher)
	err := server.Start()
	if err != nil {
		return err
	}
	for {
		// XXX TODO finish
		// XXX TODO loop break
		watcher.Watch()
	}
	return nil
}

func main() {
	mode := flag.String("mode", nil, "run mode (updater|checker)")
	configPath := flag.String("config", "/etc/pdns-record-updater.yml", "config file path")
	configurator := configurator.New(configPath)
	config, err := configurator.Load()
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	belog.SetupLoggers(config.Logger)

	if (mode == "updater") {
		err := updator(config)
	} else if (mode == "checker") {
		err := checker(config)
	} else {
		err := errors.New("unexpected run mode")
	}
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
}

