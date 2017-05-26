package watcher

import (
        "github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/potix/pdns-record-updater/contexter"
	"github.com/potix/pdns-record-updater/notifier"
        "go/token"
        "go/types"
        "go/constant"
	"sync/atomic"
        "strings"
        "fmt"
        "time"
)

const (
	tfChanged    uint32 = 0x01
	tfLatestDown uint32 = 0x02
	tfLatestUp   uint32 = 0x04
)

// Watcher is struct of Watcher
type Watcher struct {
	watcherContext *contexter.Watcher
	running         uint32
	notifier        *notifier.Notifier
}

type targetTask struct {
	target   *contexter.Target
	waitChan chan bool
	next     *targetTask
}

type protoWatcherIf interface {
	isAlive() (bool)
}

var protoWatcherNewFuncMap = map[string]func(*contexter.Target) (protoWatcherIf, error) {
	"ICMP":       icmpWatcherNew,
//not implemented "UDP":        udpWatcherNew, 
//not implemented "UDPREGEXP":   udpRegexpWatcherNew,
	"TCP":        tcpWatcherNew,
	"TCPREGEXP":  tcpRegexpWatcherNew,
	"HTTP":       httpWatcherNew,
	"HTTPREGEXP": httpRegexpWatcherNew,
}

func (w *Watcher) targetWatch(task *targetTask) {
	protoWatcherNewFunc, ok := protoWatcherNewFuncMap[strings.ToUpper(task.target.Protocol)]
	if !ok {
		belog.Error("unsupported protocol type (%v)", task.target.Protocol)
		task.target.SetAlive(0)
		close(task.waitChan)
		return
	}
	protoWatcher, err := protoWatcherNewFunc(task.target)
	if err != nil {
		belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not create protocol watcher (%v)", task.target.Protocol)))
		task.target.SetAlive(0)
		close(task.waitChan)
		return
	}
	task.target.SetAlive(protoWatcher.isAlive())
	close(task.waitChan)
}

func (w *Watcher) eval(expr string) (types.TypeAndValue, error) {
	return types.Eval(token.NewFileSet(), nil, token.NoPos, expr)
}

func (w *Watcher) updateAlive(domain string, record *contexter.DynamicRecord, targetResult string, newAlive bool){
	oldAlive := record.SwapAlive(newAlive);
	var triggerFlags uint32
	for _, trigger := range record.NotifyTrigger {
		if strings.ToUpper(trigger) == "CHANGED" {
			triggerFlags |= tfChanged
		} else if strings.ToUpper(trigger) == "LATESTDOWN" {
			triggerFlags |= tfLatestUp
		} else if strings.ToUpper(trigger) == "LATESTUP" {
			triggerFlags |= tfLatestDown
		}
	}
	if (triggerFlags & tfChanged) != 0 && oldAlive != newAlive {
		w.notifier.Notify(domain, record, targetResult, oldAlive, newAlive)
	} else if (triggerFlags & tfLatestDown) != 0 && newAlive == false {
		w.notifier.Notify(domain, record, targetResult, oldAlive, newAlive)
	} else if (triggerFlags & tfLatestUp) != 0 && newAlive == true  {
		w.notifier.Notify(domain, record, targetResult, oldAlive, newAlive)
	}
}

func (w *Watcher) recordWatch(domain string, record *contexter.DynamicRecord) {
	var firstTask *targetTask
	// run target watch task
	for _, target := range record.Target {
		newTask := &targetTask {
			target: target,
		        waitChan: make(chan bool),
			next: nil,
		}
		if firstTask != nil {
			newTask.next = firstTask
		}
		firstTask = newTask
		go w.targetWatch(newTask)
	}
	// wait target watch task
	for task := firstTask; task != nil; task = task.next {
		<-task.waitChan
	}
	// create replacer
	replaceName := make([]string, 0, 2 * len(record.Target))
	targetResult := ""
	for _, target := range record.Target {
		replaceName = append(replaceName, fmt.Sprintf("%%(%v)", target.Name), fmt.Sprintf("%v", target.GetAlive()))
		targetResult = targetResult + fmt.Sprintf("%v %v %v\n", target.Name, target.Dest, target.GetAlive())
	}
        replacer := strings.NewReplacer(replaceName...)

	// exec eval
	tv, err := w.eval(replacer.Replace(record.EvalRule))
	if err != nil {
		belog.Error("can not evalute (%v)", replacer.Replace(record.EvalRule))
		w.updateAlive(domain, record, targetResult, false)
	} else {
		w.updateAlive(domain, record, targetResult, constant.BoolVal(tv.Value))
	}
	record.SetProgress(false)
}

func (w *Watcher) zoneWatch(domain string, zone *contexter.Zone) {
	for _, dynamicGroup := range zone.DynamicGroup {
		for _, record := range dynamicGroup.DynamicRecord {
			if (record.GetCurrentIntervalCount() >= record.WatchInterval) {
				if (record.CompareAndSwapProgress(false, true)) {
					// run record waatch task
					go w.recordWatch(domain, record)
					record.ClearCurrentIntervalCount()
				} else {
					// already progress last record watch task
				}
			}
			record.IncrementCurrentIntervalCount()
		}
	}
}

func (w *Watcher) watchLoop() {
	for atomic.LoadUint32(&w.running) == 1 {
		for domain, zone := range w.watcherContext.Zone {
			go w.zoneWatch(domain, zone)
		}
		time.Sleep(time.Second)
	}
}

// Init is Init
func (w *Watcher) Init() {
	for zoneName, zone := range w.watcherContext.Zone {
		for _, dynamicGroup := range zone.DynamicGroup {
			for _, record := range dynamicGroup.DynamicRecord {
				w.recordWatch(domain, record)
			}
		}
	}
}

// Start is run 
func (w *Watcher) Start() {
	atomic.StoreUint32(&w.running, 1)
	go w.watchLoop()
}

// Stop is stop
func (w *Watcher) Stop() {
	atomic.StoreUint32(&w.running, 0)
}

// New is create Wathcer
func New(context *contexter.Context) (*Watcher) {
	return &Watcher{
		watcherContext: context.Watcher,
		running:	0,
		notifier:       notifier.New(context),
	}
}