package contexter

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/BurntSushi/toml"
	"sync/atomic"
	"bytes"
)

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
	alive         uint32   // 生存フラグ
}

//SetAlive is set alive
func (t *Target) SetAlive(alive uint32) {
	atomic.StoreUint32(&t.alive, alive)
}

//GetAlive is get alive
func (t *Target) GetAlive() (uint32) {
	return atomic.LoadUint32(&t.alive)
}

// Record is negative record
type Record struct {
	Name                 string    // DNSレコード名
	Type                 string    // DNSレコードタイプ
	TTL                  uint32    // DNSレコードTTL
	Content              string    // DNSレコード内容
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
	currentIntervalCount uint32    // 現在の時間
	progress             uint32    // 監視中を示すフラグ
	Alive                uint32    // 生存フラグ
	ForceDown            uint32     // 強制的にダウンしたとみなすフラグ
	NotifyTrigger        []string  // notifierを送信するトリガー changed, latestDown, latestUp
}

// SetForceDown is set force down
func (d *DynamicRecord) SetForceDown(forceDown bool) {
	atomic.StoreUint32(&d.ForceDown, forceDown)
}

// GetForceDown is get alive
func (d *DynamicRecord) GetForceDown() (uint32) {
	return atomic.LoadUint32(&d.Alive)
}

// SwapAlive is swap alive
func (d *DynamicRecord) SwapAlive(alive uint32) (oldAlive uint32) {
	return atomic.SwapUint32(&d.Alive, alive);
}

// GetAlive is get alive
func (d *DynamicRecord) GetAlive() (uint32) {
	return atomic.LoadUint32(&d.Alive)
}

// SetProgress is set progress
func (d *DynamicRecord) SetProgress(progress uint32) {
	atomic.StoreUint32(&d.progress, progress);
}

// CompareAndSwapProgress is set progress
func (r *Record) CompareAndSwapProgress(oldProgress uint32, newProgress uint32) (bool) {
	return atomic.CompareAndSwapUint32(&d.progress, oldProgress, newProgress);
}

// GetCurrentIntervalCount is get currentIntervalCount
func (d *DynamicRecord) GetCurrentIntervalCount() (uint32) {
	return d.currentIntervalCount
}

// IncrementCurrentIntervalCount is increment currentIntervalCount
func (d *DynamicRecord) IncrementCurrentIntervalCount() {
	d.currentIntervalCount++
}

// ClearCurrentIntervalCount is clear currentIntervalCount
func (d *DynamicRecord) ClearCurrentIntervalCount() {
	d.currentIntervalCount = 0
}

// DynamicGroup is dynamicGroup
type DynamicGroup struct {
	DynamicRecord  []*DynamicRecord  // 動的レコード
	NegativeRecord []*SimpleRecord   // 動的レコードが全て死んだ場合に有効になるレコード
}

// Zone is zone
type Zone struct {
	NameServer     []*Record        // ネームサーバーレコードリスト
	FixedRecord    []*Record        // 固定レコードリスト
	DynamicGroup   []*DynamicGroup  // 動的なレコードグループのリスト
}

// Watcher is watcher
type Watcher struct {
	Zone [string]*Zone
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
	Subject       string   // 題名テンプレート
	Body          string   // bodyテンプレート
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

// Dump is cump
func (c *Context) Dump() {
	var buffer bytes.Buffer
	encoder := toml.NewEncoder(&buffer)
	err := encoder.Encode(c)
	if err != nil {
	    belog.Error("%v", errors.Wrap(err, "can not dump context"))
	    return
	}
	belog.Debug("%v", buffer.String())
}
