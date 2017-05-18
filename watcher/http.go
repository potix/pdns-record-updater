package watcher

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"github.com/potix/pdns-record-updater/configurator"
	"github.com/potix/pdns-record-updater/cacher"
	"net/http"
	"time"
	"fmt"
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

func (h *httpWatcher) getHTTP() (uint32, bool, error) {
	httpClient := &http.Client{
		Timeout: time.Duration(h.timeout) * time.Second,
	}
	res, err := httpClient.Get(h.url)
	if err != nil {
		return 0, true, errors.Wrap(err, fmt.Sprintf("can not get url (%v) (%v)", h.url, err))
	}
	defer res.Body.Close()
	if h.status != nil && len(h.status) > 0 {
		match := false
		for _, status := range h.status {
			if status == res.Status {
				match = true
				break
			}
		}
		if !match {
			belog.Debug("not match status (%v)", h.status)
			return 0, false, nil
		}
	}
	if h.useRegexp {
		if h.resSize == 0 {
			h.resSize = 2048
		}
		rb := make([]byte, h.resSize)
		_, err := res.Body.Read(rb)
		if err != nil {
			return 0, true, errors.Wrap(err, fmt.Sprintf("can not read body (%v) (%v)", h.url, err))
		}
		loc := h.regexp.FindIndex(rb, 0)
		if loc == nil {
			belog.Debug("not match regexp (%v) (%v)", h.regexpStr, rb)
			return 0, false, nil
		}
	}
	return 1, false, nil
}

func (h *httpWatcher) isAlive() (uint32) {
	var i uint32
	for i = 0; i < h.retry; i++ {
		alive, retryable, err := h.getHTTP()
		if err != nil {
			belog.Error("%v", err)
		}
		if !retryable {
			return alive
		}
		if h.retryWait > 0 {
			time.Sleep(time.Duration(h.retryWait))
		}
	}
	belog.Error("retry count is exceeded limit", h.url)
	return 0
}

func httpWatcherNew(target *configurator.Target) (protoWatcherIf, error) {
        return &httpWatcher {
                useRegexp:  false,
                url:        target.Dest,
                retry:      target.Retry,
                retryWait:  target.RetryWait,
                timeout:    target.Timeout,
                status:     target.HTTPStatus,
        }, nil
}

func httpRegexpWatcherNew(target *configurator.Target) (protoWatcherIf, error) {
	regexp, err := cacher.GetRegexpFromCache(target.Regexp, 0)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not get compiled regexp (%v)", target.Regexp))
	}
        return &httpWatcher {
                useRegexp:  true,
                url:        target.Dest,
                retry:      target.Retry,
                retryWait:  target.RetryWait,
                timeout:    target.Timeout,
                status:     target.HTTPStatus,
                regexpStr:  target.Regexp,
                regexp:     regexp,
                resSize:    target.ResSize,
        }, nil
}
