package contexter

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/BurntSushi/toml"
	"github.com/potix/pdns-record-updater/configurator"
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

// NameServerRecord is static record
type NameServerRecord struct {
	Name        string  // SOAプライマリ,DNSレコード名
	Type        string  // DNSレコードタイプ
	TTL         uint32  // DNSレコードTTL
	Content     string  // DNSレコード内容
	Email       string  // SOAレコードEmail
}

// Validate is validate static record
func (n *NameServerRecord) Validate() (bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if n.Name == "" || n.Type == "" || n.TTL == 0 || n.Content == "" || n.Email == "" {
		return false
	}
	return true
}

// StaticRecord is static record
type StaticRecord struct {
	Name        string  // DNSレコード名
	Type        string  // DNSレコードタイプ
	TTL         uint32  // DNSレコードTTL
	Content     string  // DNSレコード内容
}

// Validate is validate static record
func (s *StaticRecord) Validate() (bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if s.Name == "" || s.Type == "" || s.TTL == 0 || s.Content == "" {
		return false
	}
	return true
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

// Validate is validate dynamic record
func (d *DynamicRecord) Validate() (bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if d.Name == "" || d.Type == "" || d.TTL == 0 || d.Content == "" ||
            d.WatchInterval == 0 || d.EvalRule == "" || d.Target == nil {
		return false
	}
	for _, target := range d.Target {
		if target.Name == "" || target.Protocol == "" || target.Dest == "" {
			return false
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
	Name        string  // DNSレコード名
	Type        string  // DNSレコードタイプ
	TTL         uint32  // DNSレコードTTL
	Content     string  // DNSレコード内容
}

// Validate is validate negative record
func (n *NegativeRecord) Validate() (bool) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	if n.Name == "" || n.Type == "" || n.TTL == 0 || n.Content == "" {
		return false
	}
	return true
}

// DynamicGroup is dynamicGroup
type DynamicGroup struct {
	dynamicRecord  []*DynamicRecord  // 動的レコード                                     [mutable]
	negativeRecord []*NegativeRecord // 動的レコードが全て死んだ場合に有効になるレコード [mutable]
}

// GetDynamicRecord is get name server
func (d *DynamicGroup) GetDynamicRecord() ([]*DynamicRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newDynamicRecord := make([]*DynamicRecord, 0, len(d.dynamicRecord))
	copy(newDynamicRecord, d.dynamicRecord)
	return newDynamicRecord
}

// FindDynamicRecord is fins name server
func (d *DynamicGroup) FindDynamicRecord(n string, t string, c string) ([]*DynamicRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newDynamicRecord := make([]*DynamicRecord, 0, len(d.dynamicRecord))
	for _, dr := range d.dynamicRecord {
		if dr.Name == n && dr.Type == t && dr.Content == c {
			newDynamicRecord = append(newDynamicRecord, dr)
		}
	}
	return newDynamicRecord
}

// AddDynamicRecord is add name server
func (d *DynamicGroup) AddDynamicRecord(dynamicRecord *DynamicRecord) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	for _, dr := range d.dynamicRecord {
		if dr.Name == dynamicRecord.Name && dr.Type == dynamicRecord.Type && dr.Content == dynamicRecord.Content {
			return errors.Errorf("already exists")
		}
	}
	d.dynamicRecord = append(d.dynamicRecord, dynamicRecord)
	return nil
}

// DeleteDynamicRecord is delete name server
func (d *DynamicGroup) DeleteDynamicRecord(n string, t string, c string) (error) {
	deleted := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newDynamicRecord := make([]*DynamicRecord, 0, len(d.dynamicRecord) - 1)
	for _, dr := range d.dynamicRecord {
		if dr.Name == n && dr.Type == t && dr.Content == c {
			deleted = true
			continue
		}
		newDynamicRecord = append(newDynamicRecord, dr)
	}
	if !deleted {
		return errors.Errorf("not exists")
	}
	d.dynamicRecord = newDynamicRecord
	return nil
}

// ReplaceDynamicRecord is replace name server
func (d *DynamicGroup) ReplaceDynamicRecord(n string, t string, c string, dynamicRecord *DynamicRecord) (error) {
	replaced := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newDynamicRecord := make([]*DynamicRecord, 0, len(d.dynamicRecord) - 1)
	for _, dr := range d.dynamicRecord {
		if dr.Name == n && dr.Type == t && dr.Content == c {
			newDynamicRecord = append(newDynamicRecord, dynamicRecord)
			replaced = true
		} else {
			newDynamicRecord = append(newDynamicRecord, dr)
		}
	}
	if !replaced {
		return errors.Errorf("not exists")
	}
	d.dynamicRecord = newDynamicRecord
	return nil
}

// GetNegativeRecord is get name server
func (d *DynamicGroup) GetNegativeRecord() ([]*NegativeRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNegativeRecord := make([]*NegativeRecord, 0, len(d.negativeRecord))
	copy(newNegativeRecord, d.negativeRecord)
	return newNegativeRecord
}

// FindNegativeRecord is fins name server
func (d *DynamicGroup) FindNegativeRecord(n string, t string, c string) ([]*NegativeRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNegativeRecord := make([]*NegativeRecord, 0, len(d.negativeRecord))
	for _, nr := range d.negativeRecord {
		if nr.Name == n && nr.Type == t && nr.Content == c {
			newNegativeRecord = append(newNegativeRecord, nr)
		}
	}
	return newNegativeRecord
}

// AddNegativeRecord is add name server
func (d *DynamicGroup) AddNegativeRecord(negativeRecord *NegativeRecord) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	for _, nr := range d.negativeRecord {
		if nr.Name == negativeRecord.Name && nr.Type == negativeRecord.Type && nr.Content == negativeRecord.Content {
			errors.Errorf("already exists");
		}
	}
	d.negativeRecord = append(d.negativeRecord, negativeRecord)
	return nil
}

// DeleteNegativeRecord is delete name server
func (d *DynamicGroup) DeleteNegativeRecord(n string, t string, c string) (error) {
	deleted := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNegativeRecord := make([]*NegativeRecord, 0, len(d.negativeRecord) - 1)
	for _, nr := range d.negativeRecord {
		if nr.Name == n && nr.Type == t && nr.Content == c {
			deleted = true
			continue
		}
		newNegativeRecord = append(newNegativeRecord, nr)
	}
	if !deleted {
		errors.Errorf("not exists");
	}
	d.negativeRecord = newNegativeRecord
	return nil
}

// ReplaceNegativeRecord is replace name server
func (d *DynamicGroup) ReplaceNegativeRecord(n string, t string, c string, negativeRecord *NegativeRecord) (error) {
	replaced := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNegativeRecord := make([]*NegativeRecord, 0, len(d.negativeRecord) - 1)
	for _, nr := range d.negativeRecord {
		if nr.Name == n && nr.Type == t && nr.Content == c {
			newNegativeRecord = append(newNegativeRecord, negativeRecord)
			replaced = true
		} else {
			newNegativeRecord = append(newNegativeRecord, nr)
		}
	}
	if !replaced {
		errors.Errorf("not exists");
	}
	d.negativeRecord = newNegativeRecord
	return nil
}

// Zone is zone
type Zone struct {
	nameServer     []*NameServerRecord           // ネームサーバーレコードリスト   [mutable]
	staticRecord   []*StaticRecord           // 固定レコードリスト             [mutable]
	dynamicGroup   map[string]*DynamicGroup  // 動的なレコードグループのリスト [mutable]
}

// GetNameServer is get name server
func (z *Zone) GetNameServer() ([]*NameServerRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNameServer := make([]*NameServerRecord, 0, len(z.nameServer))
	copy(newNameServer, z.nameServer)
	return newNameServer
}

// FindNameServer is fins name server
func (z *Zone) FindNameServer(n string, t string, c string) ([]*NameServerRecord) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNameServer := make([]*NameServerRecord, 0, len(z.nameServer))
	for _, ns := range z.nameServer {
		if ns.Name == n && ns.Type == t && ns.Content == c {
			newNameServer = append(newNameServer, ns)
		}
	}
	return newNameServer
}

// AddNameServer is add name server
func (z *Zone) AddNameServer(nameServer *NameServerRecord) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	for _, ns := range z.nameServer {
		if ns.Name == nameServer.Name && ns.Type == nameServer.Type && ns.Content == nameServer.Content {
			return errors.Errorf("already exists")
		}
	}
	z.nameServer = append(z.nameServer, nameServer)
	return nil
}

// DeleteNameServer is delete name server
func (z *Zone) DeleteNameServer(n string, t string, c string) (error) {
	deleted := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNameServer := make([]*NameServerRecord, 0, len(z.nameServer) - 1)
	for _, ns := range z.nameServer {
		if ns.Name == n && ns.Type == t && ns.Content == c {
			deleted = true
			continue
		}
		newNameServer = append(newNameServer, ns)
	}
	if !deleted {
		return errors.Errorf("not exists")
	}
	z.nameServer = newNameServer
	return nil
}

// ReplaceNameServer is replace name server
func (z *Zone) ReplaceNameServer(n string, t string, c string, nameServer *NameServerRecord) (error) {
	replaced := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newNameServer := make([]*NameServerRecord, 0, len(z.nameServer) - 1)
	for _, ns := range z.nameServer {
		if ns.Name == n && ns.Type == t && ns.Content == c {
			newNameServer = append(newNameServer, nameServer)
			replaced = true
		} else {
			newNameServer = append(newNameServer, ns)
		}
	}
	if !replaced {
		return errors.Errorf("not exists")
	}
	z.nameServer = newNameServer
	return nil
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
	for _, sr := range z.staticRecord {
		if sr.Name == n && sr.Type == t && sr.Content == c {
			newStaticRecord = append(newStaticRecord, sr)
		}
	}
	return newStaticRecord
}

// AddStaticRecord is add name server
func (z *Zone) AddStaticRecord(staticRecord *StaticRecord) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	for _, sr := range z.staticRecord {
		if sr.Name == staticRecord.Name && sr.Type == staticRecord.Type && sr.Content == staticRecord.Content {
			return errors.Errorf("already exists")
		}
	}
	z.staticRecord = append(z.staticRecord, staticRecord)
	return nil
}

// DeleteStaticRecord is delete name server
func (z *Zone) DeleteStaticRecord(n string, t string, c string) (error) {
	deleted := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newStaticRecord := make([]*StaticRecord, 0, len(z.staticRecord) - 1)
	for _, sr := range z.staticRecord {
		if sr.Name == n && sr.Type == t && sr.Content == c {
			deleted = true
			continue
		}
		newStaticRecord = append(newStaticRecord, sr)
	}
	if !deleted {
		return errors.Errorf("not exists")
	}
	z.staticRecord = newStaticRecord
	return nil
}

// ReplaceStaticRecord is replace name server
func (z *Zone) ReplaceStaticRecord(n string, t string, c string, staticRecord *StaticRecord) (error) {
	replaced := false
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	newStaticRecord := make([]*StaticRecord, 0, len(z.staticRecord) - 1)
	for _, sr := range z.staticRecord {
		if sr.Name == n && sr.Type == t && sr.Content == c {
			newStaticRecord = append(newStaticRecord, staticRecord)
			replaced = true
		} else {
			newStaticRecord = append(newStaticRecord, sr)
		}
	}
	if !replaced {
		return errors.Errorf("not exists")
	}
	z.staticRecord = newStaticRecord
	return nil
}

// GetDynamicGroupName is get dynamic group name
func (z *Zone) GetDynamicGroupName() ([]string) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	dynamicGroupName := make([]string, 0, len(z.dynamicGroup))
	for n := range z.dynamicGroup {
		dynamicGroupName = append(dynamicGroupName, n)
	}
	return dynamicGroupName
}

// GetDynamicGroup is get dynamicGroup
func (z *Zone) GetDynamicGroup(dynamicGroupName string) (*DynamicGroup, error) {
        mutableMutex.Lock()
        defer mutableMutex.Unlock()
        dynamicGroup, ok := z.dynamicGroup[dynamicGroupName]
	if !ok {
		return nil, errors.Errorf("not exist synamic group")
	}
	return dynamicGroup, nil
}

// AddDynamicGroup is get force down
func (z *Zone) AddDynamicGroup(dynamicGroupName string) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	_, ok := z.dynamicGroup[dynamicGroupName]
	if ok {
		return errors.Errorf("already exists dynamic group name")
	}
	newDynamicGroup := &DynamicGroup {
		dynamicRecord:  make([]*DynamicRecord, 0),
		negativeRecord: make([]*NegativeRecord, 0),
	}
	z.dynamicGroup[dynamicGroupName] = newDynamicGroup
	return nil
}

// DeleteDynamicGroup is delete dynamicGroup
func (z *Zone) DeleteDynamicGroup(dynamicGroupName string) (error) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	dynamicGroup, ok := z.dynamicGroup[dynamicGroupName]
	if !ok {
		return errors.Errorf("not exist dynamic group name")
	}
	if len(dynamicGroup.dynamicRecord) == 0 && len(dynamicGroup.negativeRecord) == 0 {
		delete(z.dynamicGroup, dynamicGroupName)
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
	domain := make([]string, 0, len(w.zone))
	for d := range w.zone {
		domain = append(domain, d)
	}
	return domain
}

// GetZone is get zone
func (w *Watcher) GetZone(domain string) (*Zone, error) {
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
		nameServer:     make([]*NameServerRecord, 0),
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
	zone, ok := w.zone[domain]
	if !ok {
		return errors.Errorf("not exist domain")
	}
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
	UseTLS   bool   // TLSを使うかどうか
	Certfile string // 証明書ファイルパス
	Keyfile  string // プライベートキーファイルパス
}

// Server is server
type Server struct {
	Debug        bool      // デバッグモードにする
	Listen       []*Listen // リッスンリスト
	Username     string    // ユーザー名
	Password     string    // パスワード
	StaticPath   string    // Staticリソースのパス
}

// Client is server
type Client struct {
	URL           []string // url list
	Retry         uint32   // retry回数
	RetryWait     uint32   // retry時のwait時間
	Timeout       uint32   // タイムアウト
	TLSSkipVerify bool     // TLSのverifyをスキップルするかどうか
	Username      string   // ユーザー名
	Password      string   // パスワード
}

// Updater is updater
type Updater struct {
	PdnsServer string
        PdnsAPIKey string
}

// Initializer is initializer
type Initializer struct {
	PdnsSqlitePath string
}

// Context is context
type Context struct {
	Watcher  *Watcher             // 監視設定
	Notifier *Notifier            // 通知設定
	Server   *Server              // サーバー設定
	Client   *Client              // クライアント設定
	Logger   *belog.ConfigLoggers // ログ設定
}

// Contexter is contexter
type Contexter struct {
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
	err := c.configurator.Save(newContext)
	if err != nil {
		return err
	}
	c.Context = newContext
	return nil
}

// SaveConfig is save config
func (c *Contexter) SaveConfig() (error) {
	mutableMutex.Lock()
        defer mutableMutex.Unlock()
	return c.configurator.Load(c.Context)
}

// Dump is dump context
func (c *Contexter) Dump() {
	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	mutableMutex.Lock()
	err := encoder.Encode(c.Context)
        mutableMutex.Unlock()
	if err != nil {
	    belog.Error("%v", errors.Wrap(err, "can not dump context"))
	    return
	}
	belog.Debug("%v", buffer.String())
}

// New is create new contexter
func New(configurator *configurator.Configurator) (*Contexter) {
	return &Contexter {
		Context: nil,
		configurator: configurator,
	}
}

func init() {
	mutableMutex = new(sync.Mutex)
}


