package watcher

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/glenn-brown/golang-pkg-pcre/src/pkg/pcre"
	"github.com/potix/pdns-record-updater/contexter"
	"github.com/potix/pdns-record-updater/cacher"
	"net/http"
	"net/url"
	"crypto/tls"
	"time"
	"fmt"
	"strings"
)

type httpWatcher struct {
        useRegexp     bool
        url           string
        method        string
        retry         uint32
        retryWait     uint32
        timeout       uint32
	status        []string
        regexpStr     string
        regexp        *pcre.Regexp
        resSize       uint32
	tlsSkipVerify bool
}

func (h *httpWatcher) getHTTP() (uint32, bool, error) {
        u, err := url.Parse(h.url)
	if err != nil {
		return 0, false, errors.Errorf("can not parse url (%v)", h.url)
	}
	method := strings.ToUpper(h.method)
	if method == "" {
		method = "GET"
	}
	if method != "GET" && method != "HEAD" {
		return 0, false, errors.Errorf("unsupported method (%v)", method)
	}
	transport := &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}
	if u.Scheme == "https" {
		transport.TLSClientConfig = &tls.Config{ServerName: u.Host, InsecureSkipVerify: h.tlsSkipVerify}
	}
	httpClient := &http.Client{
		Transport: transport,
		Timeout: time.Duration(h.timeout) * time.Second,
	}
	request, err := http.NewRequest(method, h.url, nil)
	if err != nil {
		return 0, false, errors.Wrap(err, fmt.Sprintf("can not create request (%v)", h.url))
	}
	res, err := httpClient.Do(request)
	if err != nil {
		return 0, true, errors.Wrap(err, fmt.Sprintf("can not get url (%v)", h.url))
	}
	defer res.Body.Close()
	if h.status != nil && len(h.status) > 0 {
		match := false
		for _, status := range h.status {
			if len(status) <= len(res.Status) && status == res.Status[0:len(status)] {
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
			return 0, true, errors.Wrap(err, fmt.Sprintf("can not read body (%v)", h.url))
		}
		loc := h.regexp.FindIndex(rb, 0)
		if loc == nil {
			belog.Debug("not match regexp (%v) (%v)", h.regexpStr, rb)
			return 0, false, nil
		}
	}
	belog.Debug("http ok (%v)", h.url)
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
	belog.Error("retry count is exceeded limit (%v)", h.url)
	return 0
}

func httpWatcherNew(target *contexter.Target) (protoWatcherIf, error) {
        return &httpWatcher {
                useRegexp:     false,
                url:           target.Dest,
                retry:         target.Retry,
                retryWait:     target.RetryWait,
                timeout:       target.Timeout,
                status:        target.HTTPStatus,
		tlsSkipVerify: target.TLSSkipVerify,
        }, nil
}

func httpRegexpWatcherNew(target *contexter.Target) (protoWatcherIf, error) {
	regexp, err := cacher.GetRegexpFromCache(target.Regexp, 0)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not get compiled regexp (%v)", target.Regexp))
	}
        return &httpWatcher {
                useRegexp:     true,
                url:           target.Dest,
                method:        target.HTTPMethod,
                retry:         target.Retry,
                retryWait:     target.RetryWait,
                timeout:       target.Timeout,
                status:        target.HTTPStatus,
                regexpStr:     target.Regexp,
                regexp:        regexp,
                resSize:       target.ResSize,
		tlsSkipVerify: target.TLSSkipVerify,
        }, nil
}
