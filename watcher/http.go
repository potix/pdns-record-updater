package watcher

import (
        "github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"net/http"
	"io/ioutil"
	"time"
)

type httpWatcher struct {
        useRegex   bool
        url        string
        retry      uint32
        retryWait  uint32
        timeout    uint32
	status     []string
        regex      *pcre.Regex
        resSize    uint32
}

func (h *httpWatcher) isAlive() (uint32) {
	for i := 0; i < h.retry; i++ {
		httpClient := &http.Client{
			Timeout: time.Duration(h.timeout) * time.Second,
		}
		res, err := httpClient.Get(url)
		if err != nil {
			// GET出来なかった
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
				return 0
			}
		}
		if h.useRegex {
			if h.resSize == 0 {
				t.resSize = 2048
			}
			rb := make([]byte, h.resSize)
			_, err := res.Body.Read(rb)
			if err != nil {
				// bodyが読めなかった
				if retryWait > 0 {
					time.Sleep(time.Duration(retryWait))
				}
				continue
			}
                        loc = h.regex.FindIndex(rb, 0)
                        if loc == nil {
                                // 正規表現に一致しなかった
                                return 0
                        }

		}
		return 1
	}
	// retryの最大に達した
	return 0
}


func httpWatcherNew(target *configurator.Target) (*protoWatcherIf) {
        return &httpWatcher {
                useRegex:   false,
                url:        target.Dest,
                retry:      target.Retry,
                retryWait:  target.RetryWait,
                timeout:    target.Timeout,
                status:     target.HTTPStatus,
        }
}

func httpRegexWatcherNew(target *configurator.Target) (*protoWatcherIf) {
        return &httpWatcher {
                useRegex:   true,
                url:        target.Dest,
                retry:      target.Retry,
                retryWait:  target.RetryWait,
                timeout:    target.Timeout,
                status:     target.HTTPStatus,
                regex:      GetRegexFromCache(target.Regex, 0),
                resSize:    target.ResSize,
        }
}
