package main

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
        "github.com/potix/pdns-record-updater/configurator"
        "github.com/potix/pdns-record-updater/contexter"
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

func runUpdater(context *contexter.Context) (err error) {
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

func runWatcher(context *contexter.Context) (error) {
	watcher := watcher.New(context)
	watcher.Init()
	server := server.New(context)
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
                case syscall.SIGINT:
			fallthrough
                case syscall.SIGQUIT:
			fallthrough
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
	mode := flag.String("mode", "", "run mode (updater|checker)")
	configPath := flag.String("config", "/etc/pdns-record-updater.yml", "config file path")
	flag.Parse()
	configurator, err := configurator.New(*configPath)
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	context, err := configurator.Load()
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	err = belog.SetupLoggers(context.Logger)
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	context.Dump()
	if (strings.ToUpper(*mode) == "UPDATER") {
		err = runUpdater(context)
	} else if (strings.ToUpper(*mode) == "WATCHER") {
		err = runWatcher(context)
	} else {
		err = errors.New("unexpected run mode")
	}
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
}

