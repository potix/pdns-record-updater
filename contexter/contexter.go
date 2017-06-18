package contexter

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/BurntSushi/toml"
	"encoding/json"
	"gopkg.in/yaml.v2"
	"github.com/potix/pdns-record-updater/configurator"
	"sync"
	"bytes"
	"strings"
)

var mutableMutex *sync.Mutex

// Target is config of target
type Target struct {
	Name           string   `json:"name"           yaml:"name"           toml:"name"`           // target名
	Protocol       string   `json:"protocol"       yaml:"protocol"       toml:"protocol"`       // プロトコル icmp, udp, udpRegexp, tcp, tcpRegexp, http, httpRegexp
	Dest           string   `json:"dest"           yaml:"dest"           toml:"dest"`           // 宛先
	TCPTLS         bool     `json:"tcpTls"         yaml:"tcpTls"         toml:"tcpTls"`         // TCPにTLSを使う
	HTTPMethod     string   `json:"httpMethod"     yaml:"httpMethod"     toml:"httpMethod"`     // HTTPメソッド
	HTTPStatusList []string `json:"httpStatusList" yaml:"httpStatusList" toml:"httpStatusList"` // OKとみなすHTTPステータスコード
	Regexp         string   `json:"regexp"         yaml:"regexp"         toml:"regexp"`         // OKとみなす正規表現  
	ResSize        uint32   `json:"resSize"        yaml:"resSize"        toml:"resSize"`        // 受信する最大レスポンスサイズ   
	Retry          uint32   `json:"retry"          yaml:"retry"          toml:"retry"`          // リトライ回数 
	RetryWait      uint32   `json:"retryWait"      yaml:"retryWait"      toml:"retryWait"`      // 次のリトライまでの待ち時間   
	Timeout        uint32   `json:"timeout"        yaml:"timeout"        toml:"timeout"`        // タイムアウトしたとみなす時間  
	TLSSkipVerify  bool     `json:"tlsSkipVerify"  yaml:"tlsSkipVerify"  toml:"tlsSkipVerify"`  // TLSの検証をスキップする 
	alive          bool                                                                         // 生存フラグ 
}

// Validate ins validate target (no lock)
func (t *Target) Validate() (bool) {
	if t.Name == "" || t.Protocol == "" || t.Dest == "" {
		return false
	}
	if t.Protocol == "http" || t.Protocol == "httpRegexp" {
		if t.HTTPMethod == "" || t.HTTPStatusList == nil || len(t.HTTPStatusList) == 0 {
			return false
		}
	}
	return true
}

// SetAlive is set alive
func (t *Target) SetAlive(alive bool) {
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
	t.alive = alive
}

// GetAlive is get alive
func (t *Target) GetAlive() (bool) {
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
	return t.alive
}

type NotifyTrigger string

func (n NotifyTrigger) Validate() (bool) {
	if strings.ToUpper(string(n)) != "CHANGED" && strings.ToUpper(string(n)) != "LATESTDOWN" && strings.ToUpper(string(n)) != "LATESTUP" {
		return false
	}
	return true
}

// DynamicRecord is config of record
type DynamicRecord struct {
	Name                 string          `json:"name"              yaml:"name"              toml:"name"`              // DNSレコード名
	Type                 string          `json:"type"              yaml:"type"              toml:"type"`              // DNSレコードタイプ
	TTL                  int32           `json:"ttl"               yaml:"ttl"               toml:"ttl"`               // DNSレコードTTL 
	Content              string          `json:"content"           yaml:"content"           toml:"content"`           // DNSレコード内容                  
	TargetList           []*Target       `json:"targetList"        yaml:"targetList"        toml:"targetList"`        // ターゲットリスト
	EvalRule             string          `json:"evalRule"          yaml:"evalRule"          toml:"evalRule"`          // 生存を判定する際のターゲットの評価ルール example: "(%(a) && (%(b) || !%(c))) || ((%(d) && %(e)) || !%(f))"  (a,b,c,d,e,f is target name)
	WatchInterval        uint32          `json:"watchInterval"     yaml:"watchInterval"     toml:"watchInterval"`     // 監視する間隔
	currentIntervalCount uint32                                                                                 // 現在の時間                       [mutable]
	progress             bool                                                                                   // 監視中を示すフラグ               [mutable]
	Alive                bool            `json:"alive"             yaml:"alive"             toml:"alive"`             // 生存フラグ                       [mutable]
	ForceDown            bool            `json:"forceDown"         yaml:"forceDown"         toml:"forceDown"`         // 強制的にダウンしたとみなすフラグ [mutable]
	NotifyTriggerList    []NotifyTrigger `json:"notifyTriggerList" yaml:"notifyTriggerList" toml:"notifyTriggerList"` // notifierを送信するトリガー changed, latestDown, latestUp
}

// Validate is validate dynamic record (no lock)
func (d *DynamicRecord) Validate() (bool) {
	if d.Name == "" || d.Type == "" || d.TTL == 0 || d.Content == "" ||
            d.WatchInterval == 0 || d.EvalRule == "" || d.TargetList == nil {
		return false
	}
	for _, target := range d.TargetList {
		if !target.Validate() {
			return false
		}
	}
	if d.NotifyTriggerList != nil {
		for _, notifyTrigger := range d.NotifyTriggerList {
			if !notifyTrigger.Validate() {
				return false
			}
		}
	}
	return true
}

// GetCurrentIntervalCount is get currentIntervalCount
func (d *DynamicRecord) GetCurrentIntervalCount() (uint32) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	return d.currentIntervalCount
}

// IncrementCurrentIntervalCount is increment currentIntervalCount
func (d *DynamicRecord) IncrementCurrentIntervalCount() {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	d.currentIntervalCount++
}

// ClearCurrentIntervalCount is clear currentIntervalCount
func (d *DynamicRecord) ClearCurrentIntervalCount() {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	d.currentIntervalCount = 0
}

// SetProgress is set progress
func (d *DynamicRecord) SetProgress(progress bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	d.progress = progress
}

// CompareAndSwapProgress is set progress
func (d *DynamicRecord) CompareAndSwapProgress(oldProgress bool, newProgress bool) (bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.progress == oldProgress {
		d.progress = newProgress
		return true
	}
	return false
}

// SwapAlive is swap alive
func (d *DynamicRecord) SwapAlive(newAlive bool) (oldAlive bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	oldAlive = d.Alive
	d.Alive = newAlive
	return oldAlive
}

// GetAlive is get alive
func (d *DynamicRecord) GetAlive() (bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	return d.Alive
}

// SetForceDown is set force down
func (d *DynamicRecord) SetForceDown(forceDown bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	d.ForceDown = forceDown
}

// GetForceDown is get force down
func (d *DynamicRecord) GetForceDown() (bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	return d.Alive
}

// NegativeRecord is negative record
type NegativeRecord struct {
	Name        string `json:"name"    yaml:"name"    toml:"name"`     // DNSレコード名
	Type        string `json:"type"    yaml:"type"    toml:"type"`     // DNSレコードタイプ
	TTL         int32  `json:"ttl"     yaml:"ttl"     toml:"ttl"`      // DNSレコードTTL
	Content     string `json:"content" yaml:"content" toml:"content"`  // DNSレコード内容
}

// Validate is validate negative record (no lock)
func (n *NegativeRecord) Validate() (bool) {
	if n.Name == "" || n.Type == "" || n.TTL == 0 || n.Content == "" {
		return false
	}
	return true
}

// NameServerRecord is static record
type NameServerRecord struct {
	Name        string `json:"name"    yaml:"name"    toml:"name"`    // SOAプライマリ,DNSレコード名
	Type        string `json:"type"    yaml:"type"    toml:"type"`    // DNSレコードタイプ
	TTL         int32  `json:"ttl"     yaml:"ttl"     toml:"ttl"`     // DNSレコードTTL
	Content     string `json:"content" yaml:"content" toml:"content"` // DNSレコード内容
}

// Validate is validate static record (no lock)
func (n *NameServerRecord) Validate() (bool) {
	if n.Name == "" || n.Type == "" || n.TTL == 0 || n.Content == "" {
		return false
	}
	return true
}

// StaticRecord is static record
type StaticRecord struct {
	Name        string `json:"name"    yaml:"name"    toml:"name"`    // DNSレコード名
	Type        string `json:"type"    yaml:"type"    toml:"type"`    // DNSレコードタイプ
	TTL         int32  `json:"ttl"     yaml:"ttl"     toml:"ttl"`     // DNSレコードTTL
	Content     string `json:"content" yaml:"content" toml:"content"` // DNSレコード内容
}

// Validate is validate static record (no lock)
func (s *StaticRecord) Validate() (bool) {
	if s.Name == "" || s.Type == "" || s.TTL == 0 || s.Content == "" {
		return false
	}
	return true
}

// DynamicGroup is dynamicGroup
type DynamicGroup struct {
	DynamicRecordList  []*DynamicRecord  `json:"dynamicRecordList"  yaml:"dynamicRecordList"  toml:"dynamicRecordList"` // 動的レコード                                     [mutable]
	NegativeRecordList []*NegativeRecord `json:"negativeRecordList" yaml:"negativeRecordList" toml:"negativeRecordList"` // 動的レコードが全て死んだ場合に有効になるレコード [mutable]
}

// Validate is validate dynamic group (no lock)
func (d *DynamicGroup) Validate() (bool) {
	if d.DynamicRecordList != nil {
		for _, dynamicRecord := range d.DynamicRecordList {
			if !dynamicRecord.Validate() {
				return false
			}
		}
	}
	if d.NegativeRecordList != nil {
		for _, negativeRecord := range d.NegativeRecordList {
			if !negativeRecord.Validate() {
				return false
			}
		}
	}
	return true
}

// GetDynamicRecord is get name server
func (d *DynamicGroup) GetDynamicRecord() ([]*DynamicRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.DynamicRecordList == nil {
		d.DynamicRecordList = make([]*DynamicRecord, 0)
	}
	newDynamicRecordList := make([]*DynamicRecord, len(d.DynamicRecordList))
	copy(newDynamicRecordList, d.DynamicRecordList)
	return newDynamicRecordList
}

// FindDynamicRecord is fins name server
func (d *DynamicGroup) FindDynamicRecord(n string, t string, c string) ([]*DynamicRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.DynamicRecordList == nil {
		d.DynamicRecordList = make([]*DynamicRecord, 0)
	}
	newDynamicRecordList := make([]*DynamicRecord, 0, len(d.DynamicRecordList))
	for _, dr := range d.DynamicRecordList {
		if dr.Name == n && dr.Type == t && dr.Content == c {
			newDynamicRecordList = append(newDynamicRecordList, dr)
		}
	}
	return newDynamicRecordList
}

// AddDynamicRecord is add name server
func (d *DynamicGroup) AddDynamicRecord(dynamicRecord *DynamicRecord) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.DynamicRecordList == nil {
		d.DynamicRecordList = make([]*DynamicRecord, 0, 1)
	}
	for _, dr := range d.DynamicRecordList {
		if dr.Name == dynamicRecord.Name && dr.Type == dynamicRecord.Type && dr.Content == dynamicRecord.Content {
			return errors.Errorf("can not add because already exists")
		}
	}
	d.DynamicRecordList = append(d.DynamicRecordList, dynamicRecord)
	return nil
}

// DeleteDynamicRecord is delete name server
func (d *DynamicGroup) DeleteDynamicRecord(n string, t string, c string) (error) {
	deleted := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.DynamicRecordList == nil {
		d.DynamicRecordList = make([]*DynamicRecord, 0)
	}
	newDynamicRecordList := make([]*DynamicRecord, 0, len(d.DynamicRecordList))
	for _, dr := range d.DynamicRecordList {
		if dr.Name == n && dr.Type == t && dr.Content == c {
			deleted = true
			continue
		}
		newDynamicRecordList = append(newDynamicRecordList, dr)
	}
	if !deleted {
		return errors.Errorf("can not delete because not exists")
	}
	d.DynamicRecordList = newDynamicRecordList
	return nil
}

// ReplaceDynamicRecord is replace name server
func (d *DynamicGroup) ReplaceDynamicRecord(n string, t string, c string, dynamicRecord *DynamicRecord) (error) {
	replaced := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.DynamicRecordList == nil {
		d.DynamicRecordList = make([]*DynamicRecord, 0)
	}
	newDynamicRecordList := make([]*DynamicRecord, 0, len(d.DynamicRecordList) - 1)
	for _, dr := range d.DynamicRecordList {
		if dr.Name == n && dr.Type == t && dr.Content == c {
			newDynamicRecordList = append(newDynamicRecordList, dynamicRecord)
			replaced = true
		} else {
			newDynamicRecordList = append(newDynamicRecordList, dr)
		}
	}
	if !replaced {
		return errors.Errorf("can not replace because not exists")
	}
	d.DynamicRecordList = newDynamicRecordList
	return nil
}

// GetNegativeRecord is get name server
func (d *DynamicGroup) GetNegativeRecord() ([]*NegativeRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.NegativeRecordList == nil {
		d.NegativeRecordList = make([]*NegativeRecord, 0)
	}
	newNegativeRecordList := make([]*NegativeRecord, len(d.NegativeRecordList))
	copy(newNegativeRecordList, d.NegativeRecordList)
	return newNegativeRecordList
}

// FindNegativeRecord is fins name server
func (d *DynamicGroup) FindNegativeRecord(n string, t string, c string) ([]*NegativeRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.NegativeRecordList == nil {
		d.NegativeRecordList = make([]*NegativeRecord, 0)
	}
	newNegativeRecordList := make([]*NegativeRecord, 0, len(d.NegativeRecordList))
	for _, nr := range d.NegativeRecordList {
		if nr.Name == n && nr.Type == t && nr.Content == c {
			newNegativeRecordList = append(newNegativeRecordList, nr)
		}
	}
	return newNegativeRecordList
}

// AddNegativeRecord is add name server
func (d *DynamicGroup) AddNegativeRecord(negativeRecord *NegativeRecord) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.NegativeRecordList == nil {
		d.NegativeRecordList = make([]*NegativeRecord, 0, 1)
	}
	for _, nr := range d.NegativeRecordList {
		if nr.Name == negativeRecord.Name && nr.Type == negativeRecord.Type && nr.Content == negativeRecord.Content {
			errors.Errorf("can not add because already exists");
		}
	}
	d.NegativeRecordList = append(d.NegativeRecordList, negativeRecord)
	return nil
}

// DeleteNegativeRecord is delete name server
func (d *DynamicGroup) DeleteNegativeRecord(n string, t string, c string) (error) {
	deleted := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.NegativeRecordList == nil {
		d.NegativeRecordList = make([]*NegativeRecord, 0)
	}
	newNegativeRecordList := make([]*NegativeRecord, 0, len(d.NegativeRecordList))
	for _, nr := range d.NegativeRecordList {
		if nr.Name == n && nr.Type == t && nr.Content == c {
			deleted = true
			continue
		}
		newNegativeRecordList = append(newNegativeRecordList, nr)
	}
	if !deleted {
		errors.Errorf("can not delete because not exists");
	}
	d.NegativeRecordList = newNegativeRecordList
	return nil
}

// ReplaceNegativeRecord is replace name server
func (d *DynamicGroup) ReplaceNegativeRecord(n string, t string, c string, negativeRecord *NegativeRecord) (error) {
	replaced := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.NegativeRecordList == nil {
		d.NegativeRecordList = make([]*NegativeRecord, 0)
	}
	newNegativeRecordList := make([]*NegativeRecord, 0, len(d.NegativeRecordList) - 1)
	for _, nr := range d.NegativeRecordList {
		if nr.Name == n && nr.Type == t && nr.Content == c {
			newNegativeRecordList = append(newNegativeRecordList, negativeRecord)
			replaced = true
		} else {
			newNegativeRecordList = append(newNegativeRecordList, nr)
		}
	}
	if !replaced  {
		errors.Errorf("can not replace because not exists");
	}
	d.NegativeRecordList = newNegativeRecordList
	return nil
}

// Zone is zone
type Zone struct {
        PrimaryNameServer string                   `json:"email"            yaml:"email"            toml:"email"`            // primary name server [mutable]
        Email             string                   `json:"email"            yaml:"email"            toml:"email"`            // email [mutable]
	NameServerList    []*NameServerRecord      `json:"nameServerList "  yaml:"nameServerList"   toml:"nameServerList"`   // ネームサーバーレコードリスト   [mutable]
	StaticRecordList  []*StaticRecord          `json:"staticRecordList" yaml:"staticRecordList" toml:"staticRecordList"` // 固定レコードリスト             [mutable]
	DynamicGroupMap   map[string]*DynamicGroup `json:"dynamicGroupMap " yaml:"dynamicGroupMap"  toml:"dynamicGroupMap"`  // 動的なレコードグループのリスト [mutable]
}

// Validate is validate zone (no lock)
func (z *Zone) Validate() (bool) {
	if z.PrimaryNameServer == "" || z.Email == "" {
		return false
	}
	if z.NameServerList != nil {
		for _, nameServer := range z.NameServerList {
			if !nameServer.Validate() {
				return false
			}
		}
	}
	if z.StaticRecordList != nil {
		for _, staticRecord := range z.StaticRecordList {
			if !staticRecord.Validate() {
				return false
			}
		}
	}
	if z.DynamicGroupMap != nil {
		for dynamicGroupName, dynamicGroup := range z.DynamicGroupMap {
			if dynamicGroupName == "" || !dynamicGroup.Validate() {
				return false
			}
		}
	}
	return true
}

// GetPrimaryNameServer is get primary name server
func  (z *Zone) GetPrimaryNameServer() (string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	return z.PrimaryNameServer
}

// SetPrimaryNameServer is set primary name server
func  (z *Zone) SetPrimaryNameServer(primaryNameServer string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	z.PrimaryNameServer = primaryNameServer
}

// GetEmail is get email
func  (z *Zone) GetEmail() (string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	return z.Email
}

// SetEmail is set email
func  (z *Zone) SetEmail(email string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	z.Email = email
}

// GetNameServer is get name server
func (z *Zone) GetNameServer() ([]*NameServerRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.NameServerList == nil {
		z.NameServerList = make([]*NameServerRecord, 0)
	}
	newNameServerList := make([]*NameServerRecord, len(z.NameServerList))
	copy(newNameServerList, z.NameServerList)
	return newNameServerList
}

// FindNameServer is fins name server
func (z *Zone) FindNameServer(n string, t string, c string) ([]*NameServerRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.NameServerList == nil {
		z.NameServerList = make([]*NameServerRecord, 0)
	}
	newNameServerList := make([]*NameServerRecord, 0, len(z.NameServerList))
	for _, ns := range z.NameServerList {
		if ns.Name == n && ns.Type == t && ns.Content == c {
			newNameServerList = append(newNameServerList, ns)
		}
	}
	return newNameServerList
}

// AddNameServer is add name server
func (z *Zone) AddNameServer(nameServer *NameServerRecord) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.NameServerList == nil {
		z.NameServerList = make([]*NameServerRecord, 0, 1)
	}
	for _, ns := range z.NameServerList {
		if ns.Name == nameServer.Name && ns.Type == nameServer.Type && ns.Content == nameServer.Content {
			return errors.Errorf("can not add because already exists")
		}
	}
	z.NameServerList = append(z.NameServerList, nameServer)
	return nil
}

// DeleteNameServer is delete name server
func (z *Zone) DeleteNameServer(n string, t string, c string) (error) {
	deleted := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.NameServerList == nil {
		z.NameServerList = make([]*NameServerRecord, 0)
	}
	newNameServerList := make([]*NameServerRecord, 0, len(z.NameServerList) - 1)
	for _, ns := range z.NameServerList {
		if ns.Name == n && ns.Type == t && ns.Content == c {
			deleted = true
			continue
		}
		newNameServerList = append(newNameServerList, ns)
	}
	if !deleted {
		return errors.Errorf("can not delete because not exists")
	}
	z.NameServerList = newNameServerList
	return nil
}

// ReplaceNameServer is replace name server
func (z *Zone) ReplaceNameServer(n string, t string, c string, nameServer *NameServerRecord) (error) {
	replaced := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.NameServerList == nil {
		z.NameServerList = make([]*NameServerRecord, 0)
	}
	newNameServerList := make([]*NameServerRecord, 0, len(z.NameServerList) - 1)
	for _, ns := range z.NameServerList {
		if ns.Name == n && ns.Type == t && ns.Content == c {
			newNameServerList = append(newNameServerList, nameServer)
			replaced = true
		} else {
			newNameServerList = append(newNameServerList, ns)
		}
	}
	if !replaced {
		return errors.Errorf("can not replace because not exists")
	}
	z.NameServerList = newNameServerList
	return nil
}

// GetStaticRecord is get name server
func (z *Zone) GetStaticRecord() ([]*StaticRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.StaticRecordList == nil {
		z.StaticRecordList = make([]*StaticRecord, 0)
	}
	newStaticRecordList := make([]*StaticRecord, len(z.StaticRecordList))
	copy(newStaticRecordList, z.StaticRecordList)
	return newStaticRecordList
}

// FindStaticRecord is fins name server
func (z *Zone) FindStaticRecord(n string, t string, c string) ([]*StaticRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.StaticRecordList == nil {
		z.StaticRecordList = make([]*StaticRecord, 0)
	}
	newStaticRecordList := make([]*StaticRecord, 0, len(z.StaticRecordList))
	for _, sr := range z.StaticRecordList {
		if sr.Name == n && sr.Type == t && sr.Content == c {
			newStaticRecordList = append(newStaticRecordList, sr)
		}
	}
	return newStaticRecordList
}

// AddStaticRecord is add name server
func (z *Zone) AddStaticRecord(staticRecord *StaticRecord) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.StaticRecordList == nil {
		z.StaticRecordList = make([]*StaticRecord, 0, 1)
	}
	for _, sr := range z.StaticRecordList {
		if sr.Name == staticRecord.Name && sr.Type == staticRecord.Type && sr.Content == staticRecord.Content {
			return errors.Errorf("can not add because already exists")
		}
	}
	z.StaticRecordList = append(z.StaticRecordList, staticRecord)
	return nil
}

// DeleteStaticRecord is delete name server
func (z *Zone) DeleteStaticRecord(n string, t string, c string) (error) {
	deleted := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.StaticRecordList == nil {
		z.StaticRecordList = make([]*StaticRecord, 0)
	}
	newStaticRecordList := make([]*StaticRecord, 0, len(z.StaticRecordList) - 1)
	for _, sr := range z.StaticRecordList {
		if sr.Name == n && sr.Type == t && sr.Content == c {
			deleted = true
			continue
		}
		newStaticRecordList = append(newStaticRecordList, sr)
	}
	if !deleted {
		return errors.Errorf("can not delete because not exists")
	}
	z.StaticRecordList = newStaticRecordList
	return nil
}

// ReplaceStaticRecord is replace name server
func (z *Zone) ReplaceStaticRecord(n string, t string, c string, staticRecord *StaticRecord) (error) {
	replaced := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.StaticRecordList == nil {
		z.StaticRecordList = make([]*StaticRecord, 0)
	}
	newStaticRecordList := make([]*StaticRecord, 0, len(z.StaticRecordList) - 1)
	for _, sr := range z.StaticRecordList {
		if sr.Name == n && sr.Type == t && sr.Content == c {
			newStaticRecordList = append(newStaticRecordList, staticRecord)
			replaced = true
		} else {
			newStaticRecordList = append(newStaticRecordList, sr)
		}
	}
	if !replaced {
		return errors.Errorf("can not replace because not exists")
	}
	z.StaticRecordList = newStaticRecordList
	return nil
}

// GetDynamicGroupName is get dynamic group name
func (z *Zone) GetDynamicGroupNameList() ([]string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.DynamicGroupMap == nil {
		z.DynamicGroupMap = make(map[string]*DynamicGroup)
	}
	dynamicGroupNameList := make([]string, 0, len(z.DynamicGroupMap))
	for n := range z.DynamicGroupMap {
		dynamicGroupNameList = append(dynamicGroupNameList, n)
	}
	return dynamicGroupNameList
}

// GetDynamicGroup is get dynamicGroup
func (z *Zone) GetDynamicGroup(dynamicGroupName string) (*DynamicGroup, error) {
        mutableMutex.Lock()
        defer mutableMutex.Unlock()
	if z.DynamicGroupMap == nil {
		z.DynamicGroupMap = make(map[string]*DynamicGroup)
	}
        dynamicGroup, ok := z.DynamicGroupMap[dynamicGroupName]
	if !ok {
		return nil, errors.Errorf("not exist synamic group")
	}
	return dynamicGroup, nil
}

// AddDynamicGroup is get force down
func (z *Zone) AddDynamicGroup(dynamicGroupName string) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.DynamicGroupMap == nil {
		z.DynamicGroupMap = make(map[string]*DynamicGroup)
	}
	_, ok := z.DynamicGroupMap[dynamicGroupName]
	if ok {
		return errors.Errorf("already exists dynamic group name")
	}
	newDynamicGroup := &DynamicGroup {
		DynamicRecordList:  make([]*DynamicRecord, 0),
		NegativeRecordList: make([]*NegativeRecord, 0),
	}
	z.DynamicGroupMap[dynamicGroupName] = newDynamicGroup
	return nil
}

// DeleteDynamicGroup is delete dynamicGroup
func (z *Zone) DeleteDynamicGroup(dynamicGroupName string) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if z.DynamicGroupMap == nil {
		z.DynamicGroupMap = make(map[string]*DynamicGroup)
	}
	dynamicGroup, ok := z.DynamicGroupMap[dynamicGroupName]
	if !ok {
		return errors.Errorf("not exist dynamic group name")
	}
	if len(dynamicGroup.DynamicRecord) == 0 && len(dynamicGroup.NegativeRecord) == 0 {
		delete(z.DynamicGroupMap, dynamicGroupName)
		return nil
	}
	return errors.Errorf("not empty dynamic group")
}

// Watcher is watcher
type Watcher struct {
	ZoneMap       map[string]*Zone `json:"zoneMap"      yaml:"zoneMap"        toml:"zoneMap"`       // ゾーン [mutable]
	NotifySubject string           `json:"notifySybject" yaml:"notifySybject" toml:"notifySybject"` // Notifyの題名テンプレート
	NotifyBody    string           `json:"notifyBody"    yaml:"notifyBody"    toml:"notifyBody"`    // Notifyの本文テンプレート
}

// Validate is validate Wacther (no lock)
func (z *Watcher) Validate() (bool) {
	if z.ZoneMap != nil {
		for domain, zone := range z.ZoneMap {
			if domain == "" || !zone.Validate() {
				return false
			}
		}
	}
	return true
}

// GetDomain is get domain
func (w *Watcher) GetDomain() ([]string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if w.ZoneMap == nil {
		w.ZoneMap = make(map[string]*Zone)
	}
	domain := make([]string, 0, len(w.ZoneMap))
	for d := range w.ZoneMap {
		domain = append(domain, d)
	}
	return domain
}

// GetZone is get zone
func (w *Watcher) GetZone(domain string) (*Zone, error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if w.ZoneMap == nil {
		w.ZoneMap = make(map[string]*Zone)
	}
	zone, ok := w.ZoneMap[domain]
	if !ok {
		return nil, errors.Errorf("not exist domain")
	}
	return zone, nil
}

// AddZone is get force down
func (w *Watcher) AddZone(domain string, email string, primaryNameServer string) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if w.ZoneMap == nil {
		w.ZoneMap = make(map[string]*Zone)
	}
	_, ok := w.ZoneMap[domain]
	if ok {
		return errors.Errorf("already exist domain")
	}
	newZone := &Zone {
		Email:              email,
		PrimaryNameServer:  primaryNameServer,
		NameServerList:     make([]*NameServerRecord, 0),
		StaticRecordList:   make([]*StaticRecord, 0),
		DynamicGroupList:   make(map[string]*DynamicGroup),
	}
	w.ZoneMap[domain] = newZone
	return nil
}

// DeleteZone is delete zone
func (w *Watcher) DeleteZone(domain string) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if w.ZoneMap == nil {
		w.ZoneMap = make(map[string]*Zone)
	}
	zone, ok := w.ZoneMap[domain]
	if !ok {
		return errors.Errorf("not exist domain")
	}
	if len(zone.NameServer) == 0 && len(zone.StaticRecord) == 0 && len(zone.DynamicGroup) == 0 {
		delete(w.ZoneMap, domain)
		return nil
	}
	return errors.Errorf("not empty zone")
}

// Mail is Mail
type Mail struct {
	HostPort      string `json:"hostPort"      yaml:"hostPort"      toml:"hostPort"`      // smtp接続先ホストとポート
	Username      string `json:"username"      yaml:"username"      toml:"username"`      // ユーザ名
	Password      string `json:"password"      yaml:"password"      toml:"password"`      // パスワード
	To            string `json:"to"            yaml:"to"            toml:"to"`            // 宛先メールアドレス 複数書く場合は,で区切る
	From          string `json:"from"          yaml:"from"          toml:"from"`          // 送り元メールアドレス
	AuthType      string `json:"authType"      yaml:"authType"      toml:"authType"`      // 認証タイプ  cram-md5, plain
	UseStartTLS   bool   `json:"useStartTls"   yaml:"useStartTls"   toml:"useStartTls"`   // startTLSの使用フラグ
	UseTLS        bool   `json:"useTls"        yaml:"useTls"        toml:"useTls"`        // TLS接続の使用フラグ
	TLSSkipVerify bool   `json:"tlsSkipVerify" yaml:"tlsSkipVerify" toml:"tlsSkipVerify"` // TLSの検証をスキップする
}

// Validate is validate Mail (no lock)
func (m *Mail) Validate() (bool) {
	if m.HostPort == nil || m.To == nil || m.From == nil {
		return false
	}
	return true
}

// Notifier is Notifier
type Notifier struct {
	MailList []*Mail `json:"mailList" yaml:"mailList" toml:"mailList"` // メールリスト
}

// Validate is validate notifier (no lock)
func (n *Notifier) Validate() (bool) {
	if n.MailList != nil {
		for _, mail := range n.MailList {
			if !mail.Validate() {
				return false
			}
		}
	}
	return true
}

// Listen is listen
type Listen struct {
	AddrPort string `json:"addrPort" yaml:"addrPort" toml:"addrPort"` // リッスンするアドレスとポート
	UseTLS   bool   `json:"useTls"   yaml:"useTls"   toml:"useTls"`   // TLSを使うかどうか
	CertFile string `json:"certFile" yaml:"certFile" toml:"certFile"` // 証明書ファイルパス
	KeyFile  string `json:"keyFile"  yaml:"keyFile"  toml:"keyFile"`  // プライベートキーファイルパス
}

// Validate is validate listen (no lock)
func (l *Listen) Validate() (bool) {
	if AddrPort == nil {
		return false
	}
	return true
}

// Server is server
type Server struct {
	Debug        bool      `json:"debug"      yaml:"debug"      toml:"debug"`      // デバッグモードにする
	ListenList   []*Listen `json:"listenList  yaml:"listenList" toml:"listenList"` // リッスンリスト
	Username     string    `json:"username"   yaml:"username"   toml:"username"`   // ユーザー名
	Password     string    `json:"password"   yaml:"password"   toml:"password"`   // パスワード
	StaticPath   string    `json:"staticPath" yaml:"staticPath" toml:"staticPath"` // Staticリソースのパス
}

// Validate is validate Server (no lock)
func (s *Server) Validate() (bool) {
	if s.ListenList == nil || len(s.ListenList) == 0 {
		return false
	}
	for _, listen := range s.ListenList {
		if !listen.Validate() {
			return falase
		}
	}
	return true
}

type ServerURL string

// Validate is validate Server url (no lock)
func (s ServerURL) validate() {
	if s == "" {
		return false
	}
	return true
}

// Client is server
type Client struct {
	ServerURLList []ServerURL `json:"serverURLList" yaml:"serverURLList" toml:"serverURLList"` // server url list
	Username      string      `json:"username"      yaml:"username"      toml:"username"`      // ユーザー名
	Password      string      `json:"password"      yaml:"password"      toml:"password"`      // パスワード
	TLSSkipVerify bool        `json:"tlsSkipVerify" yaml:"tlsSkipVerify" toml:"tlsSkipVerify"` // TLSのverifyをスキップルするかどうか
	Retry         uint32      `json:"retry"         yaml:"retry"         toml:"retry"`         // retry回数
	RetryWait     uint32      `json:"retryWait"     yaml:"retryWait"     toml:"retryWait"`     // retry時のwait時間
	Timeout       uint32      `json:"timeout"       yaml:"timeout"       toml:"timeout"`       // タイムアウト
}

// Validate is validate client (no lock)
func (c *Client) Validate() (bool) {
	if c.ServerURLList == nil || len(c.ServerURLList) == 0 {
		return false
	}
	for _, serverURL := range c.ServerURLList {
		if !serverURL.Validate() {
			return false
		}
	}
	return true
}

// Updater is updater
type Updater struct {
	PdnsServer string `json:"pdnsServer" yaml:"pdnsServer" toml:"pdnsServer"` // power dns server url
        PdnsAPIKey string `json:"pdnsApiKey" yaml:"pdnsApiKey" toml:"pdnsApiKey"` // power dns api key
}

// Validate is validate updater (no lock)
func (u *Updater) Validate() (bool) {
	if u.PdnsServer == "" || u.PdnsAPIKey == "" {
		return false
	}
	return true
}

// Initializer is initializer
type Initializer struct {
	PdnsSqlitePath string `json:"pdnsSqlitePath" yaml:"pdnsSqlitePath" toml:"pdnsSqlitePath"` // power dns sqlite path
}

// Validate is validate initializer (no lock)
func (i *Initializer) Validate() (bool) {
	if i.PdnsSqlitePath == "" {
		return false
	}
	return true
}

// Context is context
type Context struct {
	Watcher     *Watcher             `json:"watcher"     yaml:"watcher"     toml:"watcher"`     // 監視設定
	Notifier    *Notifier            `json:"notifier"    yaml:"notifier"    toml:"notifier"`    // 通知設定
	Server      *Server              `json:"server"      yaml:"server"      toml:"server"`      // サーバー設定
	Client      *Client              `json:"client"      yaml:"client"      toml:"client"`      // クライアント設定
	Initializer *Initializer         `json:"initializer" yaml:"initializer" toml:"initializer"` // Initializer設定
	Updater     *Updater             `json:"updater"     yaml:"updater"     toml:"updater"`     // Updater設定
	Logger      *belog.ConfigLoggers `json:"logger"      yaml:"logger"      toml:"logger"`      // ログ設定
}

// Validate is validate Context (no lock)
func (c *Context) Validate(mode string) (bool) {
	switch strings.ToUpdater(mode) {
	case "WATCHER":
		if c.Watcher == nil || c.Server == nil  {
			return false
		}
		if !c.Watcher.Validate() || !c.Server.Validate() {
			return false
		}
	case "UPDATER":
		if c.Client  == nil || c.Initializer == nil || Updater == nil {
			return false
		}
		if !c.Client.Validate() || !c.Initializer.Validate() || !Updater.Validate() {
                        return false
                }
	case "CLIENT":
		if c.Client  == nil {
			return false
		}
		if !c.Client.Validate() {
                        return false
                }
	default:
		return false
	}
	return true
}

// Contexter is contexter
type Contexter struct {
	mode string
	Context *Context
	configurator *configurator.Configurator
}

// Lock is lock context
func (c *Contexter) Lock() {
	mutableMutex.Lock()
}

// Unlock is lock context
func (c *Contexter) Unlock() {
        mutableMutex.Unlock()
}

// LoadConfig is load config
func (c *Contexter) LoadConfig() (error){
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
	newContext := new(Context)
	err := c.configurator.Load(newContext)
	if err != nil {
		return err
	}
	if !newContext.Validate(c.mode) {
		return errors.Errorf("invalid config")
	}
	c.Context = newContext
	return nil
}

// SaveConfig is save config
func (c *Contexter) SaveConfig() (error) {
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
	return c.configurator.Save(c.Context)
}

// DumpContext is dump context
func (c *Contexter) DumpContext(format string) ([]byte, error) {
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
        switch format {
        case "toml":
                var buffer bytes.Buffer
                encoder := toml.NewEncoder(&buffer)
                err := encoder.Encode(c.Context)
                if err != nil {
                        return nil, errors.Wrap(err, "can not encode with toml")
                }
                return buffer.Bytes(), nil
        case "yaml":
                y, err := yaml.Marshal(c.Context)
                if err != nil {
                        return nil, errors.Wrap(err, "can not encode with yaml")
                }
		return y, nil
        case "json":
                j, err := json.Marshal(c.Context)
                if err != nil {
                        return nil, errors.Wrap(err, "can not encode with json")
                }
		return j, nil
        default:
                return nil, errors.Errorf("unexpected format (%v)", format)
        }
}

// New is create new contexter
func New(mode string, configurator *configurator.Configurator) (*Contexter) {
	return &Contexter {
		mode: mode,
		Context: nil,
		configurator: configurator,
	}
}

func init() {
	mutableMutex = new(sync.Mutex)
}


