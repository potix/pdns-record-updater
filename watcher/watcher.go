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
        "time"
        "os"
        "fmt"
)

const (
	tfChanged    uint32 = 0x01
	tfLatestDown uint32 = 0x02
	tfLatestUp   uint32 = 0x04
)

// Watcher is struct of Watcher
type Watcher struct {
	hostname        string
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
		task.target.SetAlive(false)
		close(task.waitChan)
		return
	}
	protoWatcher, err := protoWatcherNewFunc(task.target)
	if err != nil {
		belog.Error("%v", errors.Wrap(err, fmt.Sprintf("can not create protocol watcher (%v)", task.target.Protocol)))
		task.target.SetAlive(false)
		close(task.waitChan)
		return
	}
	task.target.SetAlive(protoWatcher.isAlive())
	close(task.waitChan)
}

func (w *Watcher) eval(expr string) (types.TypeAndValue, error) {
	return types.Eval(token.NewFileSet(), nil, token.NoPos, expr)
}

func (w *Watcher) updateAlive(domain string, groupName string, record *contexter.DynamicRecord, targetResult string, newAlive bool){
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
	t := time.Now()
        replacer := strings.NewReplacer(
                "%(hostname)", w.hostname,
                "%(time)", t.Format("2006-01-02 15:04:05"),
                "%(domain)", domain,
                "%(groupName)", groupName,
                "%(name)", record.Name,
                "%(type)", record.Type,
                "%(content)", record.Content,
                "%(oldAlive)", fmt.Sprintf("%v", oldAlive),
                "%(newAlive)", fmt.Sprintf("%v", newAlive),
                "%(detail)", targetResult)
	subject := w.watcherContext.NotifySubject
	if subject == "" {
		subject = "%(hostname) %(domain) %(groupName) %(name) %(content): old alive = %(oldAlive) -> new alive = %(newAlive)"
	}
	body := w.watcherContext.NotifyBody
	if body == "" {
		body = "hostname: %(hostname)\ndomain: %(domain)\ngroupName: %(groupName)\nrecord: %(name) %(type) %(content)\n%(time) old alive = %(oldAlive) -> new alive = %(newAlive)\n\n-----\n%(detail)\n-----"
	}
	if (triggerFlags & tfChanged) != 0 && oldAlive != newAlive {
		w.notifier.Notify(replacer, subject, body)
	} else if (triggerFlags & tfLatestDown) != 0 && newAlive == false {
		w.notifier.Notify(replacer, subject, body)
	} else if (triggerFlags & tfLatestUp) != 0 && newAlive == true  {
		w.notifier.Notify(replacer, subject, body)
	}
}

func (w *Watcher) recordWatch(domain string, groupName string, record *contexter.DynamicRecord) {
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
		targetResult = targetResult + fmt.Sprintf("%v %v %v %v %v %v %v %v\n",
			domain, groupName, record.Name, record.Type, record.Content, target.Name, target.Dest, target.GetAlive())
	}
        replacer := strings.NewReplacer(replaceName...)

	// exec eval
	tv, err := w.eval(replacer.Replace(record.EvalRule))
	if err != nil {
		belog.Error("can not evalute (%v)", replacer.Replace(record.EvalRule))
		w.updateAlive(domain, groupName, record, targetResult, false)
	} else {
		w.updateAlive(domain, groupName, record, targetResult, constant.BoolVal(tv.Value))
	}
	record.SetProgress(false)
}

func (w *Watcher) zoneWatch(domain string, zone *contexter.Zone) {
	dynamicGroupName := zone.GetDynamicGroupName()
	for _, dgname := range dynamicGroupName {
		dynamicGroup, err := zone.GetDynamicGroup(dgname)
		if err != nil {
			belog.Notice("%v", err)
			continue
		}
		belog.Debug("%d", len(dynamicGroup.GetDynamicRecord()))
		for _, record := range dynamicGroup.GetDynamicRecord() {
			if (record.GetCurrentIntervalCount() >= record.WatchInterval) {
				if (record.CompareAndSwapProgress(false, true)) {
					// run record waatch task
					go w.recordWatch(domain, dgname, record)
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
		domain := w.watcherContext.GetDomain()
		for _, d := range domain {
			zone, err := w.watcherContext.GetZone(d)
			if err != nil {
				belog.Notice("%v", err)
				continue
			}
			go w.zoneWatch(d, zone)
		}
		time.Sleep(time.Second)
	}
}

// Init is Init
func (w *Watcher) Init() {
	domain := w.watcherContext.GetDomain()
	for _, d := range domain {
		zone, err := w.watcherContext.GetZone(d)
		if err != nil {
			belog.Notice("%v", err)
			continue
		}
		dynamicGroupName := zone.GetDynamicGroupName()
		for _, dgname := range dynamicGroupName {
			dynamicGroup, err := zone.GetDynamicGroup(dgname)
			if err != nil {
				belog.Notice("%v", err)
				continue
			}
			for _, record := range dynamicGroup.GetDynamicRecord() {
				w.recordWatch(d, dgname, record)
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
func New(watcherContext *contexter.Watcher, notifier *notifier.Notifier) (*Watcher) {
        hostname, err := os.Hostname()
        if err != nil {
                hostname = "unknown"
        }
	return &Watcher{
		hostname:       hostname,
		watcherContext: watcherContext,
		running:	0,
		notifier:       notifier,
	}
}
