package watcher

import (
        "github.com/pkg/errors"
        "github.com/potix/belog"
	"sync/atomic"
        "go/token"
        "go/types"
        "fmt"
)

// Watcher is struct of Watcher
type Watcher struct {
	watcherConfig *configurator.Watcher
	runnning      uint32
	notifier      *notifier.Notifier
}

type targetTask struct {
	target   *configurator.Target
	waitChan chan bool
	next     *targetTask
}

type protoWatcherIf interface {
	isAlive(*configurator.Target) uint32
}

var protoWatcherNewFuncMap = map[string]func() (protoWatcherIf, error) {
	"ICMP":       icmpWatcherNew,
//not implemented "UDP":        udpWatcherNew, 
//not implemented "UDPREGEX":   udpRegexWatcherNew,
	"TCP":        tcpWatcherNew,
	"TCPREGEX":   tcpRegexWatcherNew,
	"HTTP":       httpWatcherNew,
	"HTTPREGEX":  httpRegexWatcherNew,
}

func (w *Watcher) targetWatch(task *targetTask) {
	protoWatcherNewFunc, ok := protoWatcherNewFuncMap[strings.ToUpper(task.target.TargetType)]
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
	atomic.StoreUint32(&task.target.Alive, protoWatcher.isAlive())
	close(task.waitChan)
}

func (w *Watcher) eval(expr string) (types.TypeAndValue, error) {
	return types.Eval(token.NewFileSet(), nil, token.NoPos, expr)
}

func (w *Watcher) updateAlive(record *configurator.Record, newAlive uint32){
	oldAlive := atomic.SwapUint32(&record.Alive, newAlive);
	if strings.ToUpper(record.NotifyTrigger) == "CHANGED" {
		if oldAlive != newAlive {
			w.notifier.notify(record, oldAlive, newAlive)
		}
	} else if strings.ToUpper(record.NotifyTrigger) == "LATESTDOWN" {
		if newAlive == 0 {
			w.notifier.notify(record, oldAlive, newAlive)
		}
	} else if strings.ToUpper(record.NotifyTrigger) == "LATESTUP" {
		if newAlive == 1 {
			w.notifier.notify(record, oldAlive, newAlive)
		}
	}
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
	// create replacer
	var replaceName []string
	for _, target := range record.targets {
		replaceName = append(replaceName, fmt.Sprintf("%%(%v)", target.Name), fmt.Sprintf("%v", (atomic.LoadUint32(&target.Alive) != 0)))
	}
        replacer := strings.NewReplacer(replaceName...)

	// exec eval
	tv, err := eval(replacer.Replace(configurator.Record.EvalRule))
	if err != nil {
		// eval failure
		w.updateAlive(record, 0)
	}
	val, ok := constant.BoolVal(tv.Value)
	if !ok {
		// convert failure
		w.updateAlive(record, 0)
	}
	if val {
		w.updateAlive(record, 1)
	} else  {
		w.updateAlive(record, 0)
	}
	atomic.StoreUint32(&record.progress, 0)
}

func (w *Watcher) watchLoop() {
	for (atomic.LoadUint32(&w.running)) {
		if (record.currentIntervalCount >= record.WatchInterval) {
			for _, record := range w.watcherConfig.records {
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
func New(config *configurator.Config, notifier *nofifier.Notifier) (*Watcher) {
	return &Watcher{
		wathcerConfig:  config.Watcher,
		runnning:	0,
		notifier:       notifier,
	}
}
