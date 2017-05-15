package watcher

import (
        "github.com/pkg/errors"
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
	statusList []string
        regex      *pcre.Regex
        resSize    uint32
}

func (h *httpWatcher) check() (alive bool) {
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
		if h.statusList && len(h.statusList) > 0 {
			match := false
			for _, status := range h.statusList {
				if status == res.Status {
					match = true
					break
				}
			}
			if !match {
				return false
			}
		}
		if h.useRegex {
			rb := make([]byte, t.resSize)
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
                                return false
                        }

		}
		return true
	}
	// retryの最大に達した
	return false
}


func httpWatcherNew(target *configurator.Target) (*protoWatcherIf) {
        return &httpWatcher {
                useRegex:   false,
                url:        target.Dest,
                retry:      target.Retry,
                retryWait:  target.RetryWait,
                timeout:    target.Timeout,
                statusList: target.StatusList,
        }
}

func httpRegexWatcherNew(target *configurator.Target) (*protoWatcherIf) {
        return &httpWatcher {
                useRegex:   true,
                url:        target.Dest,
                retry:      target.Retry,
                retryWait:  target.RetryWait,
                timeout:    target.Timeout,
                statusList: target.StatusList,
                regex:      GetRegexFromCache(target.Regex, 0),
                resSize:    target.ResSize,
        }
}
