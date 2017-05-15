package watcher

import (
        "github.com/pkg/errors"
	"sync/atomic"
)

// Watcher is struct of Watcher
type Watcher struct {
	config   *configurator.Config
	runnning uint32
}

type targetTask struct {
	target   *configurator.Target
	waitChan chan bool
	next     *targetTask
}

type protoWatcherIf interface {
	check(*configurator.Target) bool
}

var protoWatcherNewFuncMap = map[string]func() (protoWatcherIf, error) {
	"icmp":       icmpWatcherNew,
	"udp":        udpWatcherNew,
	"udpRegex":   udpRegexWatcherNew,
	"tcp":        tcpWatcherNew,
	"tcpRegex":   tcpRegexWatcherNew,
	"http":       httpWatcherNew,
	"httpRegex":  httpRegexWatcherNew,
}

func (w *Watcher) targetWatch(task *targetTask) {
	protoWatcherNewFunc, ok := protoWatcherNewFuncMap[task.target.TargetType]
	if !ok {
		// unsupported protocol type
		task.target.alive = false
		close(task.waitChan)
		return
	}
	protoWatcher, err := protoWatcherNewFunc(task.target)
	if err != nil {
		// can not create protocol watcher
		task.target.alive = false
		close(task.waitChan)
		return
	}
	task.target.alive = protoWatcher.check()
	close(task.waitChan)
}

func (w *Watcher) recordWatch(record *configurator.Record) {
	var firstTask *targetTask
	// run target watch task
	for _, target := range record.targets {
		newTask := &targetTask {
			target: target,
		        finishWaitChan: make(chan bool),
			next: nil,
		}
		if firstTask != nil {
			newTask.next = firstTask
		}
		firstTask = newTask
		go targetWatch(newTask)
	}
	// wait target watch task
	for task := firstTask; task != nil; task := task.next {
		<-task.waitChan
	}
	atomic.StoreUint32(&record.progress, 0)
}

func (w *Watcher) watchLoop() {
	for (atomic.LoadUint32(&w.running)) {
		if (record.currentIntervalCount >= record.WatchInterval) {
			for _, record := range w.config.records {
				if (!atomic.CompareAndSwapUint32(&record.progress, 0, 1)) {
					// run record waatch task
					go recordWatch(record)
				} else {
					// alresy progress last record watch task
				}
			}
		}
		record.currentIntervalCount++
		time.Sleep(time.Second)
	}
}

// Run is run 
func (w *Watcher) Run() {
	atomic.StoreUint32(&w.running, 1)
	go Watcher.watchLoop()
}

// Stop is stop
func (w *Watcher) Stop() {
	atomic.StoreUint32(&w.running, 0)
}

// New is create Wathcer
func New(config *configurator.Config) (*Watcher) {
	return &Watcher{
		config:         config,
		runnning:	0,
	}
}
