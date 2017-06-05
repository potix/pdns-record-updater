package client

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/potix/pdns-record-updater/contexter"
	"github.com/potix/pdns-record-updater/api/structure"
	"github.com/potix/pdns-record-updater/helper"
	"encoding/json"
	"net/http"
	"net/url"
	"io/ioutil"
	"fmt"
	"time"
)

type reqInfo struct {
	urlBase string
	url string
	resource string
}

type startEnd struct {
	start int
	end   int
}

// Client is client
type Client struct {
	urlBaseIndex  int
        clientContext *contexter.Client
}

func (c *Client) get(reqInfo *reqInfo) ([]byte, error) {
        u, err := url.Parse(reqInfo.url)
	if err != nil {
		return  nil, errors.Errorf("can not parse url (%v)", reqInfo.url)
	}
	httpClient := helper.NewHTTPClient(u.Scheme, u.Host, c.clientContext.TLSSkipVerify, c.clientContext.Timeout)
	request, err := http.NewRequest("GET", reqInfo.url, nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not create request (%v)", reqInfo.url))
	}
	request.SetBasicAuth(c.clientContext.Username, c.clientContext.Password)
	res, err := httpClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not get url (%v)", reqInfo.url))
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.Wrap(err, fmt.Sprintf("unexpected status code (%v) (%v)", reqInfo.url, res.StatusCode))
	}
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(" (%v)", reqInfo.url))
	}
	belog.Debug("http ok (%v)", reqInfo.url)
	return body, nil
}

func (c *Client) retryRequest(methodFunc func(reqInfo *reqInfo) (response []byte, err error), reqInfo *reqInfo) (response []byte, err error)  {
	var i uint32
	for i = 0; i <= c.clientContext.Retry; i++ {
		response, err = methodFunc(reqInfo)
		if err != nil {
			belog.Error("retry request (%v)", err)
			if c.clientContext.RetryWait > 0 {
				time.Sleep(time.Duration(c.clientContext.RetryWait) * time.Second)
			}
			continue
		}
		return response, err
	}
	return nil, errors.Errorf("give up retry (%v)", reqInfo.url)
}

func (c *Client) doRequest(methodFunc func(reqInfo *reqInfo) (response []byte, err error), reqInfo *reqInfo) (response []byte, err error) {
	startEnd := [...]*startEnd{
		&startEnd{ start: c.urlBaseIndex, end: len(c.clientContext.URL) },
		&startEnd{ start: 0, end: c.urlBaseIndex },
	}
	for _, startEnd := range startEnd  {
		for i := startEnd.start; i < startEnd.end; i++ {
			reqInfo.urlBase = c.clientContext.URL[i]
			reqInfo.url = reqInfo.urlBase + reqInfo.resource
			response, err = c.retryRequest(methodFunc, reqInfo)
			if err != nil {
				belog.Error("switch url base (%v)", err)
				continue
			}
			c.urlBaseIndex = i
			return response, err
		}
	}
	return nil, errors.Errorf("give up request (%v)", reqInfo.resource)
}

// GetWatcherResult is get watcher result
func (c *Client) GetWatcherResult() (result *structure.WatchResultResponse, err error) {
	reqInfo := &reqInfo {
		resource : "/v1/watcher/result",
	}
	response, err := c.doRequest(c.get, reqInfo)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not get watcher result (%v)", reqInfo.resource))
	}
	result = new(structure.WatchResultResponse)
        err = json.Unmarshal(response, result)
        if err != nil {
                return nil, errors.Wrap(err, fmt.Sprintf("can not unmarshal response (%v)", reqInfo.resource))
        }

	return result, nil
}

// New is create client
func New(context *contexter.Context) (*Client) {
        return &Client {
                urlBaseIndex:  0,
		clientContext: context.Client,
        }
}
