package main

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
        "github.com/potix/pdns-record-updater/configurator"
        "github.com/potix/pdns-record-updater/contexter"
        "github.com/potix/pdns-record-updater/api/client"
        "github.com/potix/pdns-record-updater/initializer"
        "github.com/potix/pdns-record-updater/updater"
        "github.com/potix/pdns-record-updater/watcher"
        "github.com/potix/pdns-record-updater/notifier"
        "github.com/potix/pdns-record-updater/api/server"
	"flag"
	"strings"
	"os"
	"os/signal"
	"syscall"
)

func runUpdater(contexter *contexter.Contexter) (err error) {
	client := client.New(contexter.Context.Client)
	initializer := initializer.New(contexter.Context.Initializer, client)
	initializer.Initialize()
	updater := updater.New(contexter.Context.Updater, client)
	updater.Start()
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
	updater.Stop()
	return nil
}

func runWatcher(contexter *contexter.Contexter) (error) {
	notifier := notifier.New(contexter.Context.Notifier)
	watcher := watcher.New(contexter.Context.Watcher, notifier)
	watcher.Init()
	server := server.New(contexter.Context.Server, contexter)
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
	mode := flag.String("mode", "", "run mode (updater|watcher|client)")
	configPath := flag.String("config", "/etc/pdns-record-updater.yml", "config file path")
	flag.Parse()
	configurator, err := configurator.New(*configPath)
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	contexter := contexter.New(configurator)
	err = contexter.LoadConfig()
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	err = belog.SetupLoggers(contexter.Context.Logger)
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	dump, err := contexter.DumpContext("toml")
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	belog.Debug("%v", string(dump))
	if (strings.ToUpper(*mode) == "UPDATER") {
		err = runUpdater(contexter)
	} else if (strings.ToUpper(*mode) == "WATCHER") {
		err = runWatcher(contexter)
	} else if (strings.ToUpper(*mode) == "CLIENT") {
		// TODO
	} else {
		err = errors.New("unexpected run mode")
	}
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
}

