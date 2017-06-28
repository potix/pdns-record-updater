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

// NotifyTrigger is notify trigger
type NotifyTrigger string

func (n NotifyTrigger) validate() (bool) {
	if strings.ToUpper(string(n)) != "CHANGED" && strings.ToUpper(string(n)) != "LATESTDOWN" && strings.ToUpper(string(n)) != "LATESTUP" {
		belog.Error("unexpected trigger")
		return false
	}
	return true
}


// String is string
func (n NotifyTrigger) String() (string) {
	return string(n)
}

// DynamicRecord is config of record
type DynamicRecord struct {
	Name                 string          `json:"name"              yaml:"name"              toml:"name"`              // DNSレコード名
	Type                 string          `json:"type"              yaml:"type"              toml:"type"`              // DNSレコードタイプ
	TTL                  int32           `json:"ttl"               yaml:"ttl"               toml:"ttl"`               // DNSレコードTTL 
	Content              string          `json:"content"           yaml:"content"           toml:"content"`           // DNSレコード内容                  
	TargetNameList       []string        `json:"targetNameList"    yaml:"targetNameList"    toml:"targetNameList"`    // ターゲットリスト
	EvalRule             string          `json:"evalRule"          yaml:"evalRule"          toml:"evalRule"`          // 生存を判定する際のターゲットの評価ルール example: "(%(a) && (%(b) || !%(c))) || ((%(d) && %(e)) || !%(f))"  (a,b,c,d,e,f is target name)
	Alive                bool            `json:"alive"             yaml:"alive"             toml:"alive"`             // 生存フラグ                       [mutable]
	ForceDown            bool            `json:"forceDown"         yaml:"forceDown"         toml:"forceDown"`         // 強制的にダウンしたとみなすフラグ [mutable]
	NotifyTriggerList    []NotifyTrigger `json:"notifyTriggerList" yaml:"notifyTriggerList" toml:"notifyTriggerList"` // notifierを送信するトリガー changed, latestDown, latestUp
}

func (d *DynamicRecord) validate() (bool) {
	if d.Name == "" || d.Type == "" || d.TTL == 0 || d.Content == "" ||
           d.EvalRule == "" || d.TargetNameList == nil {
		belog.Error("no name or no type or no ttl or no content or no watchInterval or no evalRule or no targetList")
		return false
	}
	for _, targetName := range d.TargetNameList {
		if targetName == "" {
			return false
		}
	}
	if d.NotifyTriggerList != nil {
		for _, notifyTrigger := range d.NotifyTriggerList {
			if !notifyTrigger.validate() {
				return false
			}
		}
	}
	return true
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
	return d.ForceDown
}

// NegativeRecord is negative record
type NegativeRecord struct {
	Name        string `json:"name"    yaml:"name"    toml:"name"`     // DNSレコード名
	Type        string `json:"type"    yaml:"type"    toml:"type"`     // DNSレコードタイプ
	TTL         int32  `json:"ttl"     yaml:"ttl"     toml:"ttl"`      // DNSレコードTTL
	Content     string `json:"content" yaml:"content" toml:"content"`  // DNSレコード内容
}

func (n *NegativeRecord) validate() (bool) {
	if n.Name == "" || n.Type == "" || n.TTL == 0 || n.Content == "" {
		belog.Error("no name or no type or no ttl or no content")
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

func (n *NameServerRecord) validate() (bool) {
	if n.Name == "" || n.Type == "" || n.TTL == 0 || n.Content == "" {
		belog.Error("no name or no type or no ttl or no content")
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

func (s *StaticRecord) validate() (bool) {
	if s.Name == "" || s.Type == "" || s.TTL == 0 || s.Content == "" {
		belog.Error("no name or no type or no ttl or no content")
		return false
	}
	return true
}

// DynamicGroup is dynamicGroup
type DynamicGroup struct {
	DynamicRecordList  []*DynamicRecord  `json:"dynamicRecordList"  yaml:"dynamicRecordList"  toml:"dynamicRecordList"`  // 動的レコード                                     [mutable]
	NegativeRecordList []*NegativeRecord `json:"negativeRecordList" yaml:"negativeRecordList" toml:"negativeRecordList"` // 動的レコードが全て死んだ場合に有効になるレコード [mutable]
}

func (d *DynamicGroup) validate() (bool) {
	if d.DynamicRecordList != nil {
		for _, dynamicRecord := range d.DynamicRecordList {
			if !dynamicRecord.validate() {
				return false
			}
		}
	}
	if d.NegativeRecordList != nil {
		for _, negativeRecord := range d.NegativeRecordList {
			if !negativeRecord.validate() {
				return false
			}
		}
	}
	return true
}

func (d *DynamicGroup) isEmpty() (bool) {
	if (d.DynamicRecordList != nil && len(d.DynamicRecordList) != 0) ||
           (d.NegativeRecordList != nil && len(d.NegativeRecordList) != 0) {
                return false
        }
	return true
}

// GetDynamicRecordList is get name server
func (d *DynamicGroup) GetDynamicRecordList() ([]*DynamicRecord) {
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
	if !dynamicRecord.validate() {
		return errors.Errorf("invalid dynamic record")
	}
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
	if !dynamicRecord.validate() {
		return errors.Errorf("invalid dynamic record")
	}
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

// GetNegativeRecordList is get name server
func (d *DynamicGroup) GetNegativeRecordList() ([]*NegativeRecord) {
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
	if !negativeRecord.validate() {
		return errors.Errorf("invalid negative record")
	}
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
	if !negativeRecord.validate() {
		return errors.Errorf("invalid negative record")
	}
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
        PrimaryNameServer string                   `json:"primaryNameServer" yaml:"primaryNameServer" toml:"primaryNameServer"` // primary name server [mutable]
        Email             string                   `json:"email"             yaml:"email"             toml:"email"`             // email [mutable]
	NameServerList    []*NameServerRecord      `json:"nameServerList"    yaml:"nameServerList"    toml:"nameServerList"`    // ネームサーバーレコードリスト   [mutable]
	StaticRecordList  []*StaticRecord          `json:"staticRecordList"  yaml:"staticRecordList"  toml:"staticRecordList"`  // 固定レコードリスト             [mutable]
	DynamicGroupMap   map[string]*DynamicGroup `json:"dynamicGroupMap"  yaml:"dynamicGroupMap"    toml:"dynamicGroupMap"`   // 動的なレコードグループのリスト [mutable]
}

func (z *Zone) validate() (bool) {
	if z.PrimaryNameServer == "" || z.Email == "" {
		belog.Error("no primaryNameServer or no email")
		return false
	}
	if z.NameServerList != nil {
		for _, nameServer := range z.NameServerList {
			if !nameServer.validate() {
				return false
			}
		}
	}
	if z.StaticRecordList != nil {
		for _, staticRecord := range z.StaticRecordList {
			if !staticRecord.validate() {
				return false
			}
		}
	}
	if z.DynamicGroupMap != nil {
		for dynamicGroupName, dynamicGroup := range z.DynamicGroupMap {
			if dynamicGroupName == "" {
				belog.Error("invalid dynamicGroupName")
				return false
			}
			if !dynamicGroup.validate() {
				return false
			}
		}
	}
	return true
}

func (z *Zone) isEmpty() (bool) {
        if (z.NameServerList != nil && len(z.NameServerList) != 0)  ||
           (z.StaticRecordList != nil && len(z.StaticRecordList) != 0) ||
           (z.DynamicGroupMap != nil && len(z.DynamicGroupMap) != 0) {
		return false
        }
	return true

}

// GetPrimaryNameServerAndEmail is get primary name server and email
func  (z *Zone) GetPrimaryNameServerAndEmail() (string, string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	return z.PrimaryNameServer, z.Email
}

// SetPrimaryNameServerAndEmail is set primary name server and email
func  (z *Zone) SetPrimaryNameServerAndEmail(primaryNameServer string, email string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	z.PrimaryNameServer = primaryNameServer
	z.Email = email
}

// GetNameServerList is get name server
func (z *Zone) GetNameServerList() ([]*NameServerRecord) {
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
	if !nameServer.validate() {
		return errors.Errorf("invalid name server")
	}
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
	if !nameServer.validate() {
		return errors.Errorf("invalid name server")
	}
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

// GetStaticRecordList is get name server
func (z *Zone) GetStaticRecordList() ([]*StaticRecord) {
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
	if !staticRecord.validate() {
		return errors.Errorf("invalid static record")
	}
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
	if !staticRecord.validate() {
		return errors.Errorf("invalid static record")
	}
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

// GetDynamicGroupNameList is get dynamic group name
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
	if dynamicGroupName == "" {
		return errors.Errorf("invalid dynamic group name")
	}
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
	if !dynamicGroup.isEmpty() {
		return errors.Errorf("not empty dynamic group")
	}
	delete(z.DynamicGroupMap, dynamicGroupName)
	return nil
}

// Target is config of target
type Target struct {
	Protocol             string   `json:"protocol"       yaml:"protocol"       toml:"protocol"`       // プロトコル icmp, udp, udpRegexp, tcp, tcpRegexp, http, httpRegexp
	Dest                 string   `json:"dest"           yaml:"dest"           toml:"dest"`           // 宛先
	TCPTLS               bool     `json:"tcpTls"         yaml:"tcpTls"         toml:"tcpTls"`         // TCPにTLSを使う
	HTTPMethod           string   `json:"httpMethod"     yaml:"httpMethod"     toml:"httpMethod"`     // HTTPメソッド
	HTTPStatusList       []string `json:"httpStatusList" yaml:"httpStatusList" toml:"httpStatusList"` // OKとみなすHTTPステータスコード
	Regexp               string   `json:"regexp"         yaml:"regexp"         toml:"regexp"`         // OKとみなす正規表現  
	ResSize              uint32   `json:"resSize"        yaml:"resSize"        toml:"resSize"`        // 受信する最大レスポンスサイズ   
	Retry                uint32   `json:"retry"          yaml:"retry"          toml:"retry"`          // リトライ回数 
	RetryWait            uint32   `json:"retryWait"      yaml:"retryWait"      toml:"retryWait"`      // 次のリトライまでの待ち時間   
	Timeout              uint32   `json:"timeout"        yaml:"timeout"        toml:"timeout"`        // タイムアウトしたとみなす時間  
	TLSSkipVerify        bool     `json:"tlsSkipVerify"  yaml:"tlsSkipVerify"  toml:"tlsSkipVerify"`  // TLSの検証をスキップする 
	WatchInterval        uint32   `json:"watchInterval"  yaml:"watchInterval"  toml:"watchInterval"`  // 監視する間隔
	currentIntervalCount uint32                                                                       // 現在の時間                       [mutable]
	progress             bool                                                                         // 監視中を示すフラグ               [mutable]
	alive                bool     `json:"alive"          yaml:"alive"          toml:"alive"`          // 生存フラグ                       [mutable]
}

func (t *Target) validate() (bool) {
	if t.Protocol == "" || t.Dest == "" || t.WatchInterval == 0 {
		belog.Error("no name or no protocol or no dest")
		return false
	}
	if t.Protocol == "http" || t.Protocol == "httpRegexp" {
		if t.HTTPMethod == "" || t.HTTPStatusList == nil || len(t.HTTPStatusList) == 0 {
			belog.Error("no httpMethod or no httpStatusList")
			return false
		}
	}
	return true
}

// GetCurrentIntervalCount is get currentIntervalCount
func (t *Target) GetCurrentIntervalCount() (uint32) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	return t.currentIntervalCount
}

// IncrementCurrentIntervalCount is increment currentIntervalCount
func (t *Target) IncrementCurrentIntervalCount() {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	t.currentIntervalCount++
}

// ClearCurrentIntervalCount is clear currentIntervalCount
func (t *Target) ClearCurrentIntervalCount() {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	t.currentIntervalCount = 0
}

// SetProgress is set progress
func (t *Target) SetProgress(progress bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	t.progress = progress
}

// CompareAndSwapProgress is set progress
func (t *Target) CompareAndSwapProgress(oldProgress bool, newProgress bool) (bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if t.progress == oldProgress {
		t.progress = newProgress
		return true
	}
	return false
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

// Update is update target 
func (t *Target) Update(newTarget *Target)  {
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
	t.Protocol = newTarget.Protocol
	t.Dest = newTarget.Dest
	t.TCPTLS = newTarget.TCPTLS
	t.HTTPMethod = newTarget.HTTPMethod
	t.HTTPStatusList = newTarget.HTTPStatusList
	t.Regexp = newTarget.Regexp
	t.ResSize = newTarget.ResSize
	t.Retry = newTarget.Retry
	t.RetryWait = newTarget.RetryWait
	t.Timeout = newTarget.Timeout
	t.TLSSkipVerify = newTarget.TLSSkipVerify
}

// TargetName is target name
type TargetName string

func (t TargetName) validate() (bool) {
        if t == "" {
                belog.Error("invalid target name")
                return false
        }
        return true
}

// String is string
func (t TargetName) String() (string) {
        return string(t)
}

// Watcher is watcher
type Watcher struct {
	ZoneMap       map[string]*Zone       `json:"zoneMap"      yaml:"zoneMap"        toml:"zoneMap"`       // ゾーン [mutable]
	TargetMap     map[string]*Target `json:"targetMap"    yaml:"targetMap"      toml:"targetMap"`     // ゾーン [mutable]
	NotifySubject string                 `json:"notifySybject" yaml:"notifySybject" toml:"notifySybject"` // Notifyの題名テンプレート 
	NotifyBody    string                 `json:"notifyBody"    yaml:"notifyBody"    toml:"notifyBody"`    // Notifyの本文テンプレート
}

func (w *Watcher) validate() (bool) {
	if w.ZoneMap != nil {
		for domain, zone := range w.ZoneMap {
			if domain == "" {
				belog.Error("invalid domain")
				return false
			}
			if !zone.validate() {
				return false
			}
		}
	}
	if w.TargetMap != nil {
		for targetName, target := range w.TargetMap {
			if targetName == "" {
				return false
			}
			if !target.validate() {
				return false
			}
		}
	}
	return true
}

// GetDomainList is get domain
func (w *Watcher) GetDomainList() ([]string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if w.ZoneMap == nil {
		w.ZoneMap = make(map[string]*Zone)
	}
	domainList := make([]string, 0, len(w.ZoneMap))
	for d := range w.ZoneMap {
		domainList = append(domainList, d)
	}
	return domainList
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
func (w *Watcher) AddZone(domain string, newZone *Zone) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if domain == "" || !newZone.validate() {
		return errors.Errorf("invalid zone")
	}
	if w.ZoneMap == nil {
		w.ZoneMap = make(map[string]*Zone)
	}
	_, ok := w.ZoneMap[domain]
	if ok {
		return errors.Errorf("already exist domain")
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
	if !zone.isEmpty() {
		return errors.Errorf("not empty zone")
	}
	delete(w.ZoneMap, domain)
	return nil
}

// GetTargetNameList is get domain
func (w *Watcher) GetTargetNameList() ([]string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if w.TargetMap == nil {
		w.TargetMap = make(map[string]*Target)
	}
	targetNameList := make([]string, 0, len(w.TargetMap))
	for tn := range w.TargetMap {
		targetNameList = append(targetNameList, tn)
	}
	return targetNameList
}

// GetTarget is get target
func (w *Watcher) GetTarget(targetName string) (*Target, error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if w.TargetMap == nil {
		w.TargetMap = make(map[string]*Target)
	}
	target, ok := w.TargetMap[targetName]
	if !ok {
		return nil, errors.Errorf("not exist domain")
	}
	return target, nil
}

// AddTarget is get force down
func (w *Watcher) AddTarget(targetName string, target *Target) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if targetName == "" || !target.validate() {
		return errors.Errorf("invalid zone")
	}
	if w.TargetMap == nil {
		w.TargetMap = make(map[string]*Target)
	}
	_, ok := w.TargetMap[targetName]
	if ok {
		return errors.Errorf("already exist domain")
	}
	w.TargetMap[targetName] = target
	return nil
}

// DeleteTarget is delete target
func (w *Watcher) DeleteTarget(targetName string) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if w.TargetMap == nil {
		w.TargetMap = make(map[string]*Target)
	}
	_, ok := w.TargetMap[targetName]
	if !ok {
		return errors.Errorf("not exist domain")
	}
	delete(w.TargetMap, targetName)
	return nil
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

func (m *Mail) validate() (bool) {
	if m.HostPort == "" || m.To == "" || m.From == "" {
		belog.Error("no hostPort or no to or no from")
		return false
	}
	return true
}

// Notifier is Notifier
type Notifier struct {
	MailList []*Mail `json:"mailList" yaml:"mailList" toml:"mailList"` // メールリスト
}

func (n *Notifier) validate() (bool) {
	if n.MailList != nil {
		for _, mail := range n.MailList {
			if !mail.validate() {
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

func (l *Listen) validate() (bool) {
	if l.AddrPort == "" {
		belog.Error("no addrPort")
		return false
	}
	if l.UseTLS {
		if l.CertFile == "" || l.KeyFile == "" {
			belog.Error("no certFile of no keyFile")
		}
	}
	return true
}

// APIServer is api server
type APIServer struct {
	Debug           bool      `json:"debug"           yaml:"debug"           toml:"debug"`           // デバッグモードにする
	ListenList      []*Listen `json:"listenList"      yaml:"listenList"      toml:"listenList"`      // リッスンリスト
	APIKey          string    `json:"apiKey"          yaml:"apiKey"          toml:"apiKey"`          // api key
	LetsEncryptPath string    `json:"letsEncryptPath" yaml:"letsEncryptPath" toml:"letsEncryptPath"` // Staticリソースのパス
}

func (a *APIServer) validate() (bool) {
	if a.ListenList == nil || len(a.ListenList) == 0 || a.APIKey == "" {
		belog.Error("no listenList or no apiKey")
		return false
	}
	for _, listen := range a.ListenList {
		if !listen.validate() {
			return false
		}
	}
	return true
}

// APIServerURL is watcher url
type APIServerURL string

func (a APIServerURL) validate() (bool) {
	if a == "" {
		belog.Error("invalid api server url")
		return false
	}
	return true
}

// String is string
func (a APIServerURL) String() (string) {
	return string(a)
}

// APIClient is server
type APIClient struct {
	APIServerURLList []APIServerURL `json:"apiServerUrlList" yaml:"apiServerUrlList" toml:"apiServerUrlList"` // api server url list
	APIKey           string         `json:"apiKey"           yaml:"apiKey"           toml:"apiKey"`           // api key
	TLSSkipVerify    bool           `json:"tlsSkipVerify"    yaml:"tlsSkipVerify"    toml:"tlsSkipVerify"`    // TLSのverifyをスキップルするかどうか
	Retry            uint32         `json:"retry"            yaml:"retry"            toml:"retry"`            // retry回数
	RetryWait        uint32         `json:"retryWait"        yaml:"retryWait"        toml:"retryWait"`        // retry時のwait時間
	Timeout          uint32         `json:"timeout"          yaml:"timeout"          toml:"timeout"`          // タイムアウト
}

func (a *APIClient) validate() (bool) {
	if a.APIServerURLList == nil || len(a.APIServerURLList) == 0 || a.APIKey == "" {
		belog.Error("no apiServerUrlList or no apiKey")
		return false
	}
	for _, apiServerURL := range a.APIServerURLList {
		if !apiServerURL.validate() {
			return false
		}
	}
	return true
}

// Initializer is initializer
type Initializer struct {
	PdnsSqlitePath string `json:"pdnsSqlitePath" yaml:"pdnsSqlitePath" toml:"pdnsSqlitePath"` // power dns sqlite path
	SoaMinimumTTL  int32  `json:"soaMinimumTTL"  yaml:"soaMinimumTTL"  toml:"soaMinimumTTL"`  // soa mininum ttl
}

func (i *Initializer) validate() (bool) {
	if i.PdnsSqlitePath == ""  {
		belog.Error("no pdnsSqlitePath")
		return false
	}
	if i.SoaMinimumTTL < 0 {
                belog.Error("invali soaMinimumTTL")
                return false
	}
	return true
}

// Updater is updater
type Updater struct {
	UpdateInterval uint32 `json:"updateInterval" yaml:"updateInterval" toml:"updateInterval"` // updateInterval
	PdnsServer     string `json:"pdnsServer"     yaml:"pdnsServer"     toml:"pdnsServer"`     // power dns server url
        PdnsAPIKey     string `json:"pdnsApiKey"     yaml:"pdnsApiKey"     toml:"pdnsApiKey"`     // power dns api key
        SoaMinimumTTL  int32  `json:"soaMinimumTTL"  yaml:"soaMinimumTTL"  toml:"soaMinimumTTL"`  // soa minimum ttl
}

func (u *Updater) validate() (bool) {
	if u.UpdateInterval == 0 || u.PdnsServer == "" || u.PdnsAPIKey == "" {
		belog.Error("no updateInterval or no pdnsServer or no pdnsApiKey")
		return false
	}
	if u.SoaMinimumTTL < 0 {
                belog.Error("invali soaMinimumTTL")
                return false
	}
	return true
}

// Manager is manager
type Manager struct {
	Debug           bool      `json:"debug"           yaml:"debug"           toml:"debug"`           // デバッグモードにする
	ListenList      []*Listen `json:"listenList"      yaml:"listenList"      toml:"listenList"`      // リッスンリスト
	Username        string    `json:"username"        yaml:"username"        toml:"username"`        // ユーザー名
	Password        string    `json:"password"        yaml:"password"        toml:"password"`        // パスワード
	LetsEncryptPath string    `json:"letsEncryptPath" yaml:"letsEncryptPath" toml:"letsEncryptPath"` // Staticリソースのパス
}

func (m *Manager) validate() (bool) {
	if m.ListenList == nil || len(m.ListenList) == 0 {
		belog.Error("no listenList")
		return false
	}
	for _, listen := range m.ListenList {
		if !listen.validate() {
			return false
		}
	}
	return true
}

// Context is context
type Context struct {
	Watcher     *Watcher             `json:"watcher"     yaml:"watcher"     toml:"watcher"`     // 監視設定         [mutable]
	Notifier    *Notifier            `json:"notifier"    yaml:"notifier"    toml:"notifier"`    // 通知設定         [mutable]
	APIServer   *APIServer           `json:"apiServer"   yaml:"apiServer"   toml:"apiServer"`   // サーバー設定     [mutable]
	APIClient   *APIClient           `json:"apiClient"   yaml:"apiClient"   toml:"apiClient"`   // クライアント設定 [mutable]
	Initializer *Initializer         `json:"initializer" yaml:"initializer" toml:"initializer"` // Initializer設定  [mutable]
	Updater     *Updater             `json:"updater"     yaml:"updater"     toml:"updater"`     // Updater設定      [mutable]
	Manager     *Manager             `json:"manager"     yaml:"manager"     toml:"manager"`     // マネージャー     [mutable]
	Logger      *belog.ConfigLoggers `json:"logger"      yaml:"logger"      toml:"logger"`      // ログ設定         [mutable]
}

func (c *Context) validate(mode string) (bool) {
	switch strings.ToUpper(mode) {
	case "WATCHER":
		if c.Watcher == nil || c.APIServer == nil  {
			return false
		}
	case "UPDATER":
		if c.APIClient  == nil || c.Initializer == nil || c.Updater == nil {
			return false
		}
	case "MANAGER":
		if c.APIClient  == nil || c.Manager == nil {
			return false
		}
	default:
		panic("not reached")
	}
	if c.Watcher != nil && !c.Watcher.validate() {
		return false
	}
	if c.Notifier != nil && !c.Notifier.validate() {
		return false
	}
	if c.APIClient != nil && !c.APIClient.validate() {
		return false
	}
	if c.APIServer != nil && !c.APIServer.validate() {
		return false
	}
	if c.Initializer != nil && !c.Initializer.validate() {
		return false
	}
	if c.Updater != nil && !c.Updater.validate() {
		return false
	}
	if c.Manager != nil && !c.Manager.validate() {
		return false
	}
	if c.Logger != nil {
		err :=  belog.ValidateLoggers(c.Logger)
		if err != nil {
			return false
		}
	}
	return true
}

// GetWatcher is get watcher
func (c *Context) GetWatcher() (*Watcher) {
	mutableMutex.Lock()
        mutableMutex.Unlock()
	return c.Watcher
}

// GetNotifier is get notifier
func (c *Context) GetNotifier() (*Notifier) {
	mutableMutex.Lock()
        mutableMutex.Unlock()
	return c.Notifier
}

// GetAPIServer is get api server
func (c *Context) GetAPIServer() (*APIServer) {
	mutableMutex.Lock()
        mutableMutex.Unlock()
	return c.APIServer
}

// GetAPIClient is get api client
func (c *Context) GetAPIClient() (*APIClient) {
	mutableMutex.Lock()
        mutableMutex.Unlock()
	return c.APIClient
}

// GetInitializer is get initializer
func (c *Context) GetInitializer() (*Initializer) {
	mutableMutex.Lock()
        mutableMutex.Unlock()
	return c.Initializer
}

// GetUpdater is get updater
func (c *Context) GetUpdater() (*Updater) {
	mutableMutex.Lock()
        mutableMutex.Unlock()
	return c.Updater
}

// GetManager is get manager
func (c *Context) GetManager() (*Manager) {
	mutableMutex.Lock()
        mutableMutex.Unlock()
	return c.Manager
}

// GetLogger is get logger
func (c *Context) GetLogger() (*belog.ConfigLoggers) {
	mutableMutex.Lock()
        mutableMutex.Unlock()
	return c.Logger
}

// Contexter is contexter
type Contexter struct {
	mode string
	Context *Context
	configurator *configurator.Configurator
}

func (c *Contexter) replaceContext (newContext *Context) {
	if (c.Context == nil ) {
		c.Context = newContext
		return
	}
	c.Context.Watcher = newContext.Watcher
	c.Context.Notifier = newContext.Notifier
	c.Context.APIServer = newContext.APIServer
	c.Context.APIClient = newContext.APIClient
	c.Context.Initializer = newContext.Initializer
	c.Context.Updater = newContext.Updater
	c.Context.Manager = newContext.Manager
	c.Context.Logger = newContext.Logger
}

// LoadConfig is load config
func (c *Contexter) LoadConfig() (error){
	newContext := new(Context)
	err := c.configurator.Load(newContext)
	if err != nil {
		return err
	}
	if !newContext.validate(c.mode) {
		return errors.Errorf("invalid config")
	}
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
	c.replaceContext(newContext)
	return nil
}

// SaveConfig is save config
func (c *Contexter) SaveConfig() (error) {
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
	return c.configurator.Save(c.Context)
}

// GetContext is get context
func (c *Contexter) GetContext(format string) ([]byte, error) {
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

// PutContext is put context
func (c *Contexter) PutContext(newContext *Context) (error) {
	if !newContext.validate(c.mode) {
		return errors.Errorf("invalid config")
	}
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
	c.replaceContext(newContext)
	return nil
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


