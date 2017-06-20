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
	"runtime"
	"fmt"
)

func runUpdater(contexter *contexter.Contexter) (error) {
	client := client.New(contexter.Context.APIClient)
	initializer := initializer.New(contexter.Context.Initializer, client)
	err := initializer.Initialize()
	if err != nil {
		return err
	}
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
	server := server.New(contexter.Context.APIServer, contexter)
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
	runtime.GOMAXPROCS(runtime.NumCPU())
	mode := flag.String("mode", "", "run mode (updater|watcher|manager)")
	configPath := flag.String("config", "/etc/pdns-record-updater.yml", "config file path")
	flag.Parse()
	if *mode == "" || *configPath == "" {
		fmt.Printf("usage: %v -mode <updater|watcher|client> -config <config path>\n", os.Args[0])
		os.Exit(1)
	}
	configurator, err := configurator.New(*configPath)
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	contexter := contexter.New(*mode, configurator)
	err = contexter.LoadConfig()
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	if contexter.Context.Logger != nil {
		err = belog.SetupLoggers(contexter.Context.Logger)
		if err != nil {
			belog.Error("%v", err)
			os.Exit(1);
		}
	}
	dump, err := contexter.DumpContext("toml")
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
	belog.Debug("%v", string(dump))
	_, err = syscall.Setsid()
	if err != nil {
		belog.Notice("%v", err)
	}
	if (strings.ToUpper(*mode) == "UPDATER") {
		err = runUpdater(contexter)
	} else if (strings.ToUpper(*mode) == "WATCHER") {
		err = runWatcher(contexter)
	} else if (strings.ToUpper(*mode) == "MANAGER") {
		// TODO
	} else {
		err = errors.New("unexpected run mode")
	}
	if err != nil {
		belog.Error("%v", err)
                os.Exit(1);
	}
}

