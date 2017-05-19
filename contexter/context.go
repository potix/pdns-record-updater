package contenxter

import (
        "github.com/potix/belog"
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
	Alive       uint32   // 生存フラグ
}

// Record is config of record
type Record struct {
	Name                 string    // DNSレコード名
	Type                 string    // DNSレコードタイプ
	Content              string    // DNSレコード内容
	Target               []*Target // ターゲットリスト
	EvalRule             string    // 生存を判定する際のターゲットの評価ルール example: "(%(a) && (%(b) || !%(c))) || ((%(d) && %(e)) || !%(f))"  (a,b,c,d,e,f is target name)
	WatchInterval        uint32    // 監視する間隔
	CurrentIntervalCount uint32    // 現在の時間
	progress             uint32    // 監視中を示すフラグ
	alive                uint32    // 生存フラグ
	NotifyTrigger        []string  // notifierを送信するトリガー changed, latestDown, latestUp
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

// Config is config
type Config struct {
	Watcher  *Watcher             // 監視設定
	Notifier *Notifier            // 通知設定
	Server   *Server              // サーバー設定
	Logger   *belog.ConfigLoggers // ログ設定
}
