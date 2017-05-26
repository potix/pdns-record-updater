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
	alive         bool     // 生存フラグ [mutable]
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
	Name                 string    // DNSレコード名
	Type                 string    // DNSレコードタイプ
	TTL                  uint32    // DNSレコードTTL
	Content              string    // DNSレコード内容
}

// DynamicRecord is config of record
type DynamicRecord struct {
	Name                 string    // DNSレコード名
	Type                 string    // DNSレコードタイプ
	TTL                  uint32    // DNSレコードTTL                   [mutable]
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

// SetTTL is set ttl
func (d *DynamicRecord) SetTTL(ttl uint32) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	d.TTL = ttl
}

// GetTTL is get ttl
func (d *DynamicRecord) GetTTL() (uint32) {
	mutableMutex.Lock()
	defer mutableMutex.Unlock()
	return d.TTL
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
	DynamicRecord  []*DynamicRecord  // 動的レコード
	NegativeRecord []*NegativeRecord // 動的レコードが全て死んだ場合に有効になるレコード
}

// Zone is zone
type Zone struct {
	NameServer     []*StaticRecord  // ネームサーバーレコードリスト
	StaticRecord   []*StaticRecord  // 固定レコードリスト
	DynamicGroup   []*DynamicGroup  // 動的なレコードグループのリスト
}

// Watcher is watcher
type Watcher struct {
	Zone          map[string]*Zone
	NotifySubject string   // Notifyの題名テンプレート
	NotifyBody    string   // Notifyの本文テンプレート
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
