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
	context         *contexter.Context
	running         uint32
	notifier	*notifier.Notifier
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

func (w Watcher) notify(watcherContext *contexter.Watcher, domain string, groupName string, record *contexter.DynamicRecord, targetResult string, newAlive bool, oldAlive bool) {
	var triggerFlags uint32
	for _, trigger := range record.NotifyTriggerList {
		if strings.ToUpper(trigger.String()) == "CHANGED" {
			triggerFlags |= tfChanged
		} else if strings.ToUpper(trigger.String()) == "LATESTDOWN" {
			triggerFlags |= tfLatestDown
		} else if strings.ToUpper(trigger.String()) == "LATESTUP" {
			triggerFlags |= tfLatestUp
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
	subject := watcherContext.NotifySubject
	if subject == "" {
		subject = "%(hostname) %(domain) %(groupName) %(name) %(content): old alive = %(oldAlive) -> new alive = %(newAlive)"
	}
	body := watcherContext.NotifyBody
	if body == "" {
		body = "hostname: %(hostname)\ndomain: %(domain)\ngroupName: %(groupName)\nrecord: %(name) %(type) %(content)\n%(time) old alive = %(oldAlive) -> new alive = %(newAlive)\n\n-----\n%(detail)\n"
	}
	if (triggerFlags & tfChanged) != 0 && oldAlive != newAlive {
		belog.Debug("notify changed")
		w.notifier.Notify(replacer, subject, body)
	} else if (triggerFlags & tfLatestDown) != 0 && !newAlive {
		belog.Debug("notify latestdown")
		w.notifier.Notify(replacer, subject, body)
	} else if (triggerFlags & tfLatestUp) != 0 && newAlive {
		belog.Debug("notify latestup")
		w.notifier.Notify(replacer, subject, body)
	}
}

func (w *Watcher) updateAlive(watcherContext *contexter.Watcher, domain string, groupName string, record *contexter.DynamicRecord, targetResult string, newAlive bool){
	oldAlive := record.SwapAlive(newAlive);
	belog.Debug("%v %v %v: new alive = %v, old alive = %v", record.Name, record.Type, record.Content, newAlive, oldAlive)
	if record.NotifyTriggerList != nil {
		w.notify(watcherContext, domain, groupName, record, targetResult, newAlive, oldAlive)
	}
}

func (w *Watcher) recordWatch(watcherContext *contexter.Watcher ,domain string, groupName string, record *contexter.DynamicRecord) {
	var firstTask *targetTask
	// run target watch task
	for _, target := range record.TargetList {
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
	replaceNameList := make([]string, 0, 2 * len(record.TargetList))
	targetResult := ""
	for _, target := range record.TargetList {
		replaceNameList = append(replaceNameList, fmt.Sprintf("%%(%v)", target.Name), fmt.Sprintf("%v", target.GetAlive()))
		targetResult = targetResult + fmt.Sprintf("%v %v %v %v %v %v %v %v\n",
			domain, groupName, record.Name, record.Type, record.Content, target.Name, target.Dest, target.GetAlive())
	}
        replacer := strings.NewReplacer(replaceNameList...)

	// exec eval
	evalString := replacer.Replace(record.EvalRule)
	belog.Debug("%v %v %v: eval = %v", record.Name, record.Type, record.Content, evalString)
	tv, err := w.eval(evalString)
	if err != nil {
		belog.Error("can not evalute (%v)", replacer.Replace(record.EvalRule))
		w.updateAlive(watcherContext, domain, groupName, record, targetResult, false)
	} else {
		w.updateAlive(watcherContext, domain, groupName, record, targetResult, constant.BoolVal(tv.Value))
	}
	record.SetProgress(false)
}

func (w *Watcher) zoneWatch(watcherContext *contexter.Watcher, domain string, zone *contexter.Zone) {
	dynamicGroupNameList := zone.GetDynamicGroupNameList()
	for _, dynamicGroupName := range dynamicGroupNameList {
		dynamicGroup, err := zone.GetDynamicGroup(dynamicGroupName)
		if err != nil {
			belog.Notice("%v", err)
			continue
		}
		for _, record := range dynamicGroup.GetDynamicRecordList() {
			if (record.GetCurrentIntervalCount() >= record.WatchInterval) {
				if (record.CompareAndSwapProgress(false, true)) {
					// run record waatch task
					go w.recordWatch(watcherContext, domain, dynamicGroupName, record)
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
		watcherContext := w.context.GetWatcher()
		domainList := watcherContext.GetDomainList()
		for _, domain := range domainList {
			zone, err := watcherContext.GetZone(domain)
			if err != nil {
				belog.Notice("%v", err)
				continue
			}
			go w.zoneWatch(watcherContext, domain, zone)
		}
		time.Sleep(time.Second)
	}
}

// Init is Init
func (w *Watcher) Init() {
	watcherContext := w.context.GetWatcher()
	domainList := watcherContext.GetDomainList()
	for _, domain := range domainList {
		zone, err := watcherContext.GetZone(domain)
		if err != nil {
			belog.Notice("%v", err)
			continue
		}
		dynamicGroupNameList := zone.GetDynamicGroupNameList()
		for _, dynamicGroupName := range dynamicGroupNameList {
			dynamicGroup, err := zone.GetDynamicGroup(dynamicGroupName)
			if err != nil {
				belog.Notice("%v", err)
				continue
			}
			for _, record := range dynamicGroup.GetDynamicRecordList() {
				w.recordWatch(watcherContext, domain, dynamicGroupName, record)
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
func New(context *contexter.Context, notifier *notifier.Notifier) (*Watcher) {
        hostname, err := os.Hostname()
        if err != nil {
                hostname = "unknown"
        }
	return &Watcher{
		hostname: hostname,
		context:  context,
		running:  0,
		notifier: notifier,
	}
}
