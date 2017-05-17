package watcher


import (
        "github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
        "net"
        "time"
)

type tcpWatcher struct {
	useRegex  bool
	ipPort    string
	retry     uint32
	retryWait uint32
	timeout   uint32
	regex     *pcre.Regex
        resSize   uint32
}

func (t *tcpWatcher) isAlive() (uint32) {
	for i := 0; i < t.retry; i++ {
		dialer := &net.Dialer{
			Timeout:   time.Duration(t.timeout) * time.Second,
			DualStack: true,
			Deadline:  time.Now().Add(time.Duration(t.timeout) * time.Second),
		}
		conn, err := dialer.Dial("tcp", t.ipPort)
		if err != nil {
			// コネクションが張れなかった
			if t.retryWait > 0 {
				time.Sleep(time.Duration(t.retryWait))
			}
			continue
		}
		defer conn.Close()
		if (t.useRegex) {
			if t.resSize == 0 {
				t.resSize = 1024
			}
			rb := make([]byte, t.resSize)
			_, err := conn.Read(rb)
			if err != nil {
				// レスポンスが読めなかった
				continue;
			}
			loc := t.regex.FindIndex(rb, 0)
			if loc == nil {
				// 読めたけど、正規表現に一致しなかった
				return 0
			}
		}
		return 1
	}
	// リトライの最大に達した
	return 0
}

func tcpWatcherNew(target *configurator.Target) (*protoWatcherIf) {
        return &tcpWatcher {
		useRegex:  false,
                ipPort:    target.Dest,
                retry:     target.Retry,
                retryWait: target.RetryWait,
                timeout:   target.Timeout,
        }
}

func tcpRegexWatcherNew(target *configurator.Target) (*protoWatcherIf) {
        return &tcpWatcher {
		useRegex:  true,
                ipPort:  target.Dest,
                retry:     target.Retry,
                retryWait: target.RetryWait,
                timeout:   target.Timeout,
		regex:     GetRegexFromCache(target.Regex, 0),
		resSize:   target.ResSize,
        }
}
