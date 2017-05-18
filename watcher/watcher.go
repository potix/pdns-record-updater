package watcher

import (
        "github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/potix/pdns-record-updater/configurator"
	"sync/atomic"
        "go/token"
        "go/types"
        "fmt"
)

const (
	tfChanged     uint32 = 0x01
	tfLatestDown uint32 = 0x02
	tfLatestUp   uint32 = 0x04
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
//not implemented "UDPREGEXP":   udpRegexpWatcherNew,
	"TCP":        tcpWatcherNew,
	"TCPREGEXP":   tcpRegexpWatcherNew,
	"HTTP":       httpWatcherNew,
	"HTTPREGEXP":  httpRegexpWatcherNew,
}

func (w *Watcher) targetWatch(task *targetTask) {
	protoWatcherNewFunc, ok := protoWatcherNewFuncMap[strings.ToUpper(task.target.TargetType)]
	if !ok {
		belog.Error("unsupported protocol type (%v)", task.target.TargetType)
		task.target.alive = false
		close(task.waitChan)
		return
	}
	protoWatcher, err := protoWatcherNewFunc(task.target)
	if err != nil {
		belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not create protocol watcher (%v)", task.target.TargetType)))
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
	var triggerFlags uint32
	for _, trigger := range record.NotifyTrigger {
		if strings.ToUpper(record.NotifyTrigger) == "CHANGED" {
			triggerFlags |= tfChanged
		} else if strings.ToUpper(record.NotifyTrigger) == "LATESTDOWN" {
			triggerFlags |= tfLatestUp
		} else if strings.ToUpper(record.NotifyTrigger) == "LATESTUP" {
			triggerFlags |= tfLatestDown
		}
	}
	if (triggerFlags & tfChanged) != 0 && oldAlive != newAlive {
		w.notifier.notify(record, oldAlive, newAlive)
	} else if (triggerFlags & tfLatestDown) != 0 && newAlive == 0 {
		w.notifier.notify(record, oldAlive, newAlive)
	} else if (triggerFlags & tfLatestUp) != 0 && newAlive == 1  {
		w.notifier.notify(record, oldAlive, newAlive)
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
	for task := firstTask; task != nil; task = task.next {
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
		belog.Error("can not evalute (%v)", replacer.Replace(configurator.Record.EvalRule))
		w.updateAlive(record, 0)
	}
	val, ok := constant.BoolVal(tv.Value)
	if !ok {
		belog.Error("can not convert to Bool from Value type (%v)", val)
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
	for atomic.LoadUint32(&w.running) {
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
