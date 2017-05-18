package watcher

import (
        "github.com/potix/belog"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"github.com/potix/pdns-record-updater/configurator"
	"net/http"
	"time"
)

type httpWatcher struct {
        useRegexp   bool
        url        string
        retry      uint32
        retryWait  uint32
        timeout    uint32
	status     []string
        regexpStr   string
        regexp      *pcre.Regexp
        resSize    uint32
}

func (h *httpWatcher) isAlive() (uint32) {
	for i := 0; i < h.retry; i++ {
		httpClient := &http.Client{
			Timeout: time.Duration(h.timeout) * time.Second,
		}
		res, err := httpClient.Get(h.url)
		if err != nil {
			belog.Notice("can not get url (%v) (%v)", url, err)
			if retryWait > 0 {
				time.Sleep(time.Duration(retryWait))
			}
			continue
		}
		defer res.Body.Close()
		if h.status && len(h.status) > 0 {
			match := false
			for _, status := range h.status {
				if status == res.Status {
					match = true
					break
				}
			}
			if !match {
				belog.Debug("not match status (%v)", h.status)
				return 0
			}
		}
		if h.useRegexp {
			if h.resSize == 0 {
				t.resSize = 2048
			}
			rb := make([]byte, h.resSize)
			_, err := res.Body.Read(rb)
			if err != nil {
				belog.Notice("can not read body (%v) (%v)", url, err)
				if retryWait > 0 {
					time.Sleep(time.Duration(retryWait))
				}
				continue
			}
                        loc = h.regexp.FindIndex(rb, 0)
                        if loc == nil {
				belog.Debug("not match regexp (%v) (%v)", h.regexpStr, rb)
                                return 0
                        }

		}
		return 1
	}
	belog.Notice("retry count is exceeded limit", h.url)
	return 0
}

func httpWatcherNew(target *configurator.Target) (*protoWatcherIf) {
        return &httpWatcher {
                useRegexp:   false,
                url:        target.Dest,
                retry:      target.Retry,
                retryWait:  target.RetryWait,
                timeout:    target.Timeout,
                status:     target.HTTPStatus,
        }
}

func httpRegexpWatcherNew(target *configurator.Target) (*protoWatcherIf) {
        return &httpWatcher {
                useRegexp:   true,
                url:        target.Dest,
                retry:      target.Retry,
                retryWait:  target.RetryWait,
                timeout:    target.Timeout,
                status:     target.HTTPStatus,
                regexpStr:   target.Regexp,
                regexp:      GetRegexpFromCache(target.Regexp, 0),
                resSize:    target.ResSize,
        }
}
