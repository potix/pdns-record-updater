package contexter

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/BurntSushi/toml"
	"sync"
	"bytes"
)

var mutableMutex *sync.Mutex

// Target is config of target
type Target struct {
	Name          string   // target名
	Protocol      string   // プロトコル icmp, udp, udpRegexp, tcp, tcpRegexp, http, httpRegexp
	Dest          string   // 宛先
	TCPTLS        bool     // TCPにTLSを使う
	HTTPMethod    string   // HTTPメソッド
	HTTPStatus    []string // OKとみなすHTTPステータスコード
	Regexp        string   // OKとみなす正規表現  
	ResSize       uint32   // 受信する最大レスポンスサイズ   
	Retry         uint32   // リトライ回数 
	RetryWait     uint32   // 次のリトライまでの待ち時間   
	Timeout       uint32   // タイムアウトしたとみなす時間  
	TLSSkipVerify bool     // TLSの検証をスキップする 
	alive         bool     // 生存フラグ 
}

//SetAlive is set alive
func (t *Target) SetAlive(alive bool) {
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
	t.alive = alive
}

//GetAlive is get alive
func (t *Target) GetAlive() (bool) {
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
	return t.alive
}

// StaticRecord is negative record
type StaticRecord struct {
	Name        string  // DNSレコード名
	Type        string  // DNSレコードタイプ
	TTL         uint32  // DNSレコードTTL
	Content     string  // DNSレコード内容
}

// DynamicRecord is config of record
type DynamicRecord struct {
	Name                 string    // DNSレコード名
	Type                 string    // DNSレコードタイプ
	TTL                  uint32    // DNSレコードTTL 
	Content              string    // DNSレコード内容                  
	Target               []*Target // ターゲットリスト
	EvalRule             string    // 生存を判定する際のターゲットの評価ルール example: "(%(a) && (%(b) || !%(c))) || ((%(d) && %(e)) || !%(f))"  (a,b,c,d,e,f is target name)
	WatchInterval        uint32    // 監視する間隔
	currentIntervalCount uint32    // 現在の時間                       [mutable]
	progress             bool      // 監視中を示すフラグ               [mutable]
	Alive                bool      // 生存フラグ                       [mutable]
	ForceDown            bool      // 強制的にダウンしたとみなすフラグ [mutable]
	NotifyTrigger        []string  // notifierを送信するトリガー changed, latestDown, latestUp
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
type NegativeRecord StaticRecord


// DynamicGroup is dynamicGroup
type DynamicGroup struct {
	dynamicRecord  []*DynamicRecord  // 動的レコード                                     [mutable]
	negativeRecord []*NegativeRecord // 動的レコードが全て死んだ場合に有効になるレコード [mutable]
}

// GetDynamicRecord is get name server
func (d *DynamicGroup) GetDynamicRecord() ([]*DynamicRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newDynamicRecord := make([]*StaticRecord, 0, len(z.dynamicRecord))
	copy(newDynamicRecord, d.dynamicRecord)
	return newDynamicRecord
}

// FindDynamicRecord is fins name server
func (d *DynamicGroup) FindDynamicRecord(n string, t string, c string) ([]*DynamicRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newDynamicRecord := make([]*StaticRecord, 0, len(z.dynamicRecord))
	for _, dr := range d.DynamicRecord {
		if dr.Name == n && dr.Type == t && dr.Content == c {
			newDynamicRecord := append(newDynamicRecord, dr)
		}
	}
	return newDynamicRecord
}

// AddDynamicRecord is add name server
func (d *DynamicGroup) AddDynamicRecord(dynamicRecord *DynamicRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	d.DynamicRecord = append(d.DynamicRecord, dynamicRecord)
}

// DeleteDynamicRecord is delete name server
func (d *DynamicGroup) DeleteDynamicRecord(n string, t string, c string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newDynamicRecord := make([]*StaticRecord, 0, len(z.dynamicRecord) - 1)
	for _, dr := range d.DynamicRecord {
		if dr.Name == n && dr.Type == t && dr.Content == c {
			continue
		}
		newDynamicRecord := append(newDynamicRecord, dr)
	}
	d.dynamicRecord = newDynamicRecord
}

// ReplaceDynamicRecord is replace name server
func (d *DynamicGroup) ReplaceDynamicRecord(n string, t string, c string, dynamicRecord *DynamicRecord) (replaced bool) {
	replaced = false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newDynamicRecord := make([]*StaticRecord, 0, len(z.dynamicRecord) - 1)
	for _, dr := range d.DynamicRecord {
		if dr.Name == n && dr.Type == t && dr.Content == c {
			newDynamicRecord := append(newDynamicRecord, dynamicRecord)
			replaced = true
		} else {
			newDynamicRecord := append(newDynamicRecord, dr)
		}
	}
	d.dynamicRecord = newDynamicRecord
	return replaced
}

// GetNegativeRecord is get name server
func (d *DynamicGroup) GetNegativeRecord() ([]*NegativeRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNegativeRecord := make([]*StaticRecord, 0, len(z.negativeRecord))
	copy(newNegativeRecord, d.negativeRecord)
	return newNegativeRecord
}

// FindNegativeRecord is fins name server
func (d *DynamicGroup) FindNegativeRecord(n string, t string, c string) ([]*NegativeRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNegativeRecord := make([]*StaticRecord, 0, len(z.negativeRecord))
	for _, nr := range d.NegativeRecord {
		if nr.Name == n && nr.Type == t && nr.Content == c {
			newNegativeRecord := append(newNegativeRecord, nr)
		}
	}
	return newNegativeRecord
}

// AddNegativeRecord is add name server
func (d *DynamicGroup) AddNegativeRecord(negativeRecord *NegativeRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	d.NegativeRecord = append(d.NegativeRecord, negativeRecord)
}

// DeleteNegativeRecord is delete name server
func (d *DynamicGroup) DeleteNegativeRecord(n string, t string, c string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNegativeRecord := make([]*StaticRecord, 0, len(z.negativeRecord) - 1)
	for _, nr := range d.NegativeRecord {
		if nr.Name == n && nr.Type == t && nr.Content == c {
			continue
		}
		newNegativeRecord := append(newNegativeRecord, nr)
	}
	d.negativeRecord = newNegativeRecord
}

// ReplaceNegativeRecord is replace name server
func (d *DynamicGroup) ReplaceNegativeRecord(n string, t string, c string, negativeRecord *NegativeRecord) (replaced bool) {
	replaced = false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNegativeRecord := make([]*StaticRecord, 0, len(z.negativeRecord) - 1)
	for _, nr := range d.NegativeRecord {
		if nr.Name == n && nr.Type == t && nr.Content == c {
			newNegativeRecord := append(newNegativeRecord, negativeRecord)
			replaced = true
		} else {
			newNegativeRecord := append(newNegativeRecord, nr)
		}
	}
	d.negativeRecord = newNegativeRecord
	return replaced
}

// Zone is zone
type Zone struct {
	nameServer     []*StaticRecord           // ネームサーバーレコードリスト   [mutable]
	staticRecord   []*StaticRecord           // 固定レコードリスト             [mutable]
	dynamicGroup   map[string]*DynamicGroup  // 動的なレコードグループのリスト [mutable]
}

// GetNameServer is get name server
func (z *Zone) GetNameServer() ([]*StaticRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNameServer := make([]*StaticRecord, 0, len(z.nameServer))
	copy(newNameServer, z.nameServer)
	return newNameServer
}

// FindNameServer is fins name server
func (z *Zone) FindNameServer(n string, t string, c string) ([]*StaticRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNameServer := make([]*StaticRecord, 0, len(z.nameServer))
	for _, ns := range z.NameServer {
		if ns.Name == n && ns.Type == t && ns.Content == c {
			newNameServer := append(newNameServer, ns)
		}
	}
	return newNameServer
}

// AddNameServer is add name server
func (z *Zone) AddNameServer(nameServer *StaticRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	z.NameServer = append(z.NameServer, nameServer)
}

// DeleteNameServer is delete name server
func (z *Zone) DeleteNameServer(n string, t string, c string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNameServer := make([]*StaticRecord, 0, len(z.nameServer) - 1)
	for _, ns := range z.NameServer {
		if ns.Name == n && ns.Type == t && ns.Content == c {
			continue
		}
		newNameServer := append(newNameServer, ns)
	}
	z.nameServer = newNameServer
}

// ReplaceNameServer is replace name server
func (z *Zone) ReplaceNameServer(n string, t string, c string, nameServer *StaticRecord) (replaced bool) {
	replaced = false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNameServer := make([]*StaticRecord, 0, len(z.nameServer) - 1)
	for _, ns := range z.NameServer {
		if ns.Name == n && ns.Type == t && ns.Content == c {
			newNameServer := append(newNameServer, nameServer)
			replaced = true
		} else {
			newNameServer := append(newNameServer, ns)
		}
	}
	z.nameServer = newNameServer
	return rplaced
}

// GetStaticRecord is get name server
func (z *Zone) GetStaticRecord() ([]*StaticRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newStaticRecord := make([]*StaticRecord, 0, len(z.staticRecord))
	copy(newStaticRecord, z.staticRecord)
	return newStaticRecord
}

// FindStaticRecord is fins name server
func (z *Zone) FindStaticRecord(n string, t string, c string) ([]*StaticRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newStaticRecord := make([]*StaticRecord, 0, len(z.staticRecord))
	for _, sr := range z.StaticRecord {
		if sr.Name == n && sr.Type == t && sr.Content == c {
			newStaticRecord := append(newStaticRecord, sr)
		}
	}
	return newStaticRecord
}

// AddStaticRecord is add name server
func (z *Zone) AddStaticRecord(staticRecord *StaticRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	z.StaticRecord = append(z.StaticRecord, staticRecord)
}

// DeleteStaticRecord is delete name server
func (z *Zone) DeleteStaticRecord(n string, t string, c string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newStaticRecord := make([]*StaticRecord, 0, len(z.staticRecord) - 1)
	for _, sr := range z.StaticRecord {
		if sr.Name == n && sr.Type == t && sr.Content == c {
			continue
		}
		newStaticRecord := append(newStaticRecord, sr)
	}
	z.staticRecord = newStaticRecord
}

// ReplaceStaticRecord is replace name server
func (z *Zone) ReplaceStaticRecord(n string, t string, c string, staticRecord *StaticRecord) (replaced bool) {
	replaced = false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newStaticRecord := make([]*StaticRecord, 0, len(z.staticRecord) - 1)
	for _, sr := range z.StaticRecord {
		if sr.Name == n && sr.Type == t && sr.Content == c {
			newStaticRecord := append(newStaticRecord, staticRecord)
			replaced = true
		} else {
			newStaticRecord := append(newStaticRecord, sr)
		}
	}
	z.staticRecord = newStaticRecord
	return replaced
}

// GetDynamicGroupName is get dynamic group name
func (z *Zone) GetDynamicGroupName() ([]string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	dynamicGroupName := make([]string, 0, len(z.dynamicGroup))
	for n, d := range z.dynamicGroup {
		dynamicGroupName = append(dynamicGroupName, n)
	}
	return dynamicGroupName
}

// AddDynamicGroup is get force down
func (z *Zone) AddDynamicGroup(dynamicGroupName string) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	_, ok := z.dynamicGroup[dynamicGroupName]
	if ok {
		return errors.Errorf("already exists dynamic group name")
	}
	newDynamicGroup = &DynamicGroup {
		dynamicRecord:  make([]*DynamicRecord),
		negativeRecord: make([]*NegativeRecord),
	}
	z.dynamicGroup[dynamicGroupName] = newDynamicGroup
	return nil
}

// DeleteDynamicGroup is delete dynamicGroup
func (z *Zone) DeleteDynamicGroup(dynamicGroupName string) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if len(dynamicGroup.nameServer) == 0 && len(dynamicGroup.staticRecord) == 0 && len(dynamicGroup.dynamicGroup) == 0 {
		delete(z.dynamicGroup, domain)
		return nil
	}
	return errors.Errorf("not empty dynamic group")
}

// Watcher is watcher
type Watcher struct {
	zone          map[string]*Zone  // ゾーン [mutable]
	NotifySubject string            // Notifyの題名テンプレート
	NotifyBody    string            // Notifyの本文テンプレート
}

// GetDomain is get domain
func (w *Watcher) GetDomain() ([]string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	domain := make([]string, 0, len(z.zone))
	for d, z := range z.zone {
		domain = append(domain, d)
	}
	return domain
}

// GetZone is get zone
func (w *Watcher) GetZone() (*Zone, error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	zone, ok := w.zone[domain]
	if !ok {
		return nil, errors.Errorf("not exist domain")
	}
	return zone, nil
}

// AddZone is get force down
func (w *Watcher) AddZone(domain string) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	_, ok := w.zone[domain]
	if ok {
		return errors.Errorf("already exist domain")
	}
	newZone := &Zone {
		nameServer:     make([]*StaticRecord, 0),
		staticRecord:   make([]*StaticRecord, 0),
		dynamicGroup:   make(map[string]*DynamicGroup),
	}
	w.zone[domain] = newZone
	return nil
}

// DeleteZone is delete zone
func (w *Watcher) DeleteZone(domain string) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if len(zone.nameServer) == 0 && len(zone.staticRecord) == 0 && len(zone.dynamicGroup) == 0 {
		delete(w.zone, domain)
		return nil
	}
	return errors.Errorf("not empty zone")
}

// Mail is Mail
type Mail struct {
	HostPort      string   // smtp接続先ホストとポート
	Username      string   // ユーザ名
	Password      string   // パスワード
	To            string   // 宛先メールアドレス 複数書く場合は,で区切る
	From          string   // 送り元メールアドレス
	AuthType      string   // 認証タイプ  cram-md5, plain
	UseStartTLS   bool     // startTLSの使用フラグ
	UseTLS        bool     // TLS接続の使用フラグ
	TLSSkipVerify bool     // TLSの検証をスキップする
}

// Notifier is Notifier
type Notifier struct {
	Mail []*Mail // メールリスト
}

// Listen is listen
type Listen struct {
	AddrPort string // リッスンするアドレスとポート
	UseTLS   string // TLSを使うかどうか
	Certfile string // 証明書ファイルパス
	Keyfile  string // プライベートキーファイルパス
}

// Server is server
type Server struct {
	Debug        bool      // デバッグモードにする
	Listen       []*Listen // リッスンリスト
	Username     string    // ユーザー名
	Password     string    // パスワード
}

// Context is Context
type Context struct {
	Watcher  *Watcher             // 監視設定
	Notifier *Notifier            // 通知設定
	Server   *Server              // サーバー設定
	Logger   *belog.ConfigLoggers // ログ設定
}

// Lock is lock context
func (c *Context) Lock() {
	mutableMutex.Lock()
}

// Unlock is lock context
func (c *Context) Unlock() {
        mutableMutex.Unlock()
}

// Dump is cump
func (c *Context) Dump() {
	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	mutableMutex.Lock()
	err := encoder.Encode(c)
        mutableMutex.Unlock()
	if err != nil {
	    belog.Error("%v", errors.Wrap(err, "can not dump context"))
	    return
	}
	belog.Debug("%v", buffer.String())
}

func init() {
	mutableMutex = new(sync.Mutex)
}
