package contexter

import (
        "github.com/potix/belog"
	"sync/atomic"
)

// Target is config of target
type Target struct {
	Name        string   // target名
	ProtoType   string   // プロトコルタイプ icmp, udp, udpRegexp, tcp, tcpRegexp, http, httpRegexp
	Dest        string   // 宛先
	HTTPStatus  []string // OKとみなすHTTPステータスコード
	Regexp      string   // OKとみなす正規表現 
	ResSize     uint32   // 受信する最大レスポンスサイズ
	Retry       uint32   // リトライ回数
	RetryWait   uint32   // 次のリトライまでの待ち時間
	Timeout     uint32   // タイムアウトしたとみなす時間
	alive       uint32   // 生存フラグ
}

//SetAlive is set alive
func (t *Target) SetAlive(alive uint32) {
	atomic.StoreUint32(&t.alive, alive)
}

//GetAlive is get alive
func (t *Target) GetAlive() (uint32) {
	return atomic.LoadUint32(&t.alive)
}

// Record is config of record
type Record struct {
	Name                 string    // DNSレコード名
	Type                 string    // DNSレコードタイプ
	Content              string    // DNSレコード内容
	Target               []*Target // ターゲットリスト
	EvalRule             string    // 生存を判定する際のターゲットの評価ルール example: "(%(a) && (%(b) || !%(c))) || ((%(d) && %(e)) || !%(f))"  (a,b,c,d,e,f is target name)
	WatchInterval        uint32    // 監視する間隔
	currentIntervalCount uint32    // 現在の時間
	progress             uint32    // 監視中を示すフラグ
	Alive                uint32    // 生存フラグ
	NotifyTrigger        []string  // notifierを送信するトリガー changed, latestDown, latestUp
}

//SwapAlive is swap alive
func (r *Record) SwapAlive(alive uint32) (oldAlive uint32) {
	return atomic.SwapUint32(&r.Alive, alive);
}

//GetAlive is get alive
func (r *Record) GetAlive() (uint32) {
	return atomic.LoadUint32(&r.Alive)
}

//SetProgress is set progress
func (r *Record) SetProgress(progress uint32) {
	atomic.StoreUint32(&r.progress, progress);
}

//CompareAndSwapProgress is set progress
func (r *Record) CompareAndSwapProgress(oldProgress uint32, newProgress uint32) (bool) {
	return atomic.CompareAndSwapUint32(&r.progress, oldProgress, newProgress);
}

//GetCurrentIntervalCount is get currentIntervalCount
func (r *Record) GetCurrentIntervalCount() (uint32) {
	return r.currentIntervalCount
}

//IncrementCurrentIntervalCount is increment currentIntervalCount
func (r *Record) IncrementCurrentIntervalCount() {
	r.currentIntervalCount++
}

// NegativeRecord is negative record
type NegativeRecord struct {
	Name                 string    // DNSレコード名
	Type                 string    // DNSレコードタイプ
	Content              string    // DNSレコード内容
}

// Zone is zone
type Zone struct {
	Record         []*Record         // レコードリスト
	NegativeRecord []*NegativeRecord // レコードが全て死んだ場合に有効になるレコード
}

// Watcher is watcher
type Watcher struct {
	Zone map[string]*Zone
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
}

// Server is server
type Server struct {
	Listen []*Listen // リッスンリスト
}

// Context is Context
type Context struct {
	Watcher  *Watcher             // 監視設定
	Notifier *Notifier            // 通知設定
	Server   *Server              // サーバー設定
	Logger   *belog.ConfigLoggers // ログ設定
}
