package watcher


import (
        "github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"github.com/potix/pdns-record-updater/contexter"
	"github.com/potix/pdns-record-updater/cacher"
	"crypto/tls"
        "net"
        "time"
	"fmt"
)

type tcpWatcher struct {
	useRegexp     bool
	ipPort        string
	retry         uint32
	retryWait     uint32
	timeout       uint32
	regexp        *pcre.Regexp
	regexpStr     string
        resSize       uint32
	useTLS        bool
	tlsSkipVerify bool
}

type connIf interface {
	Close() (error)
	Read(b []byte) (n int, err error)
}

func (t *tcpWatcher) connectTCP() (uint32, bool, error) {
	dialer := &net.Dialer{
		Timeout:   time.Duration(t.timeout) * time.Second,
		DualStack: true,
		Deadline:  time.Now().Add(time.Duration(t.timeout) * time.Second),
	}

	var conn connIf
	var err error
	if t.useTLS {
		belog.Debug("tls (%v)", t.ipPort)
		tlsConfig := &tls.Config{ InsecureSkipVerify: t.tlsSkipVerify }
		conn, err = tls.DialWithDialer(dialer, "tcp", t.ipPort, tlsConfig)
	} else {
		belog.Debug("tcp (%v)", t.ipPort)
		conn, err = dialer.Dial("tcp", t.ipPort)
	}
	if err != nil {
		return 0, true, errors.Wrap(err, fmt.Sprintf("can not connect (%v)", t.ipPort))
	}
	defer conn.Close()

	if (t.useRegexp) {
		if t.resSize == 0 {
			t.resSize = 1024
		}
		rb := make([]byte, t.resSize)
		_, err := conn.Read(rb)
		if err != nil {
			return 0, true, errors.Wrap(err, fmt.Sprintf("can not read response (%v)", t.ipPort))
		}
		loc := t.regexp.FindIndex(rb, 0)
		if loc == nil {
			belog.Debug("not match regexp (%v) (%v)", t.regexpStr, rb)
			return 0, false, nil
		}
	}
	belog.Debug("tcp ok (%v)", t.ipPort)
	return 1, false, nil
}

func (t *tcpWatcher) isAlive() (uint32) {
	var i uint32
        for i = 0; i < t.retry; i++ {
                alive, retryable, err := t.connectTCP()
                if err != nil {
                        belog.Error("%v", err)
                }
                if !retryable {
                        return alive
                }
                if t.retryWait > 0 {
                        time.Sleep(time.Duration(t.retryWait))
                }
        }
        belog.Error("retry count is exceeded limit (%v)", t.ipPort)
        return 0
}

func tcpWatcherNew(target *contexter.Target) (protoWatcherIf, error) {
        return &tcpWatcher {
		useRegexp:     false,
                ipPort:        target.Dest,
                retry:         target.Retry,
                retryWait:     target.RetryWait,
                timeout:       target.Timeout,
		useTLS:        target.TCPTLS,
		tlsSkipVerify: target.TLSSkipVerify,
        }, nil
}

func tcpRegexpWatcherNew(target *contexter.Target) (protoWatcherIf, error) {
	regexp, err := cacher.GetRegexpFromCache(target.Regexp, 0)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not get compiled regexp (%v)", target.Regexp))
	}
        return &tcpWatcher {
		useRegexp: true,
                ipPort:        target.Dest,
                retry:         target.Retry,
                retryWait:     target.RetryWait,
                timeout:       target.Timeout,
		regexp:        regexp,
		regexpStr:     target.Regexp,
		resSize:       target.ResSize,
		useTLS:        target.TCPTLS,
		tlsSkipVerify: target.TLSSkipVerify,
        }, nil
}
