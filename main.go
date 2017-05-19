package main

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
        "github.com/potix/pdns-record-updater/configurator"
//        "github.com/potix/pdns-record-updater/collector"
//        "github.com/potix/pdns-record-updater/initializer"
//        "github.com/potix/pdns-record-updater/updater"
        "github.com/potix/pdns-record-updater/watcher"
        "github.com/potix/pdns-record-updater/server"
	"flag"
	"strings"
	"os"
	"os/signal"
	"syscall"
)

func runUpdater(config *configurator.Config) (err error) {
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

func runWatcher(config *configurator.Config) (error) {
	watcher := watcher.New(config)
	watcher.Init()
	server := server.New(config)
	err := server.Start()
	if err != nil {
		return err
	}
	watcher.Start()
        sigChan := make(chan os.Signal, 1)
        signal.Notify(sigChan,
                syscall.SIGINT,
                syscall.SIGTERM,
                syscall.SIGQUIT)
Loop:
        for {
                sig := <-sigChan
                switch sig {
                case syscall.SIGTERM:
                        break Loop
                default:
                        belog.Warn("unexpected signal (%v)", sig)
                }
        }
	server.Stop()
	watcher.Stop()
	return nil
}

func main() {
	var err error
	mode := flag.String("mode", "watcher", "run mode (updater|checker)")
	configPath := flag.String("config", "/etc/pdns-record-updater.yml", "config file path")
	configurator, err := configurator.New(*configPath)
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	config, err := configurator.Load()
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	err = belog.SetupLoggers(config.Logger)
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	if (strings.ToUpper(*mode) == "UPDATER") {
		err = runUpdater(config)
	} else if (strings.ToUpper(*mode) == "WATCHER") {
		err = runWatcher(config)
	} else {
		err = errors.New("unexpected run mode")
	}
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
}

