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
        "crypto/sha256"
	"fmt"
	"strconv"
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
        context *contexter.Context
}

func (c *Client) addAuthHeader(apiClientContext *contexter.APIClient, request *http.Request, u *url.URL) {
        unixTime := time.Now().Unix()
        seedString := fmt.Sprintf("%v+%v+%v+%v", unixTime, apiClientContext.APIKey, request.Method, u.Path)
        authValue := "PDRU " + fmt.Sprintf("%x", sha256.Sum256([]byte(seedString)))
        request.Header.Set("Authorization", authValue)
        request.Header.Set("x-pdru-unixtime", strconv.FormatInt(unixTime, 10))
}

func (c *Client) get(apiClientContext *contexter.APIClient, reqInfo *reqInfo) ([]byte, error) {
        u, err := url.Parse(reqInfo.url)
	if err != nil {
		return  nil, errors.Errorf("can not parse url (%v)", reqInfo.url)
	}
	httpClient := helper.NewHTTPClient(u.Scheme, u.Host, apiClientContext.TLSSkipVerify, apiClientContext.Timeout)
	request, err := http.NewRequest("GET", reqInfo.url, nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not create request (%v)", reqInfo.url))
	}
	c.addAuthHeader(apiClientContext, request, u)
	res, err := httpClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not get url (%v)", reqInfo.url))
	}
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(" (%v)", reqInfo.url))
	}
	if res.StatusCode != 200 {
		return nil, errors.Wrap(err, fmt.Sprintf("unexpected status code (%v) (%v) (%v)", reqInfo.url, res.StatusCode, string(body)))
	}
	belog.Debug("http ok (%v)", reqInfo.url)
	return body, nil
}

func (c *Client) retryRequest(apiClientContext *contexter.APIClient, methodFunc func(apiClientContext *contexter.APIClient, reqInfo *reqInfo) (response []byte, err error), reqInfo *reqInfo) (response []byte, err error)  {
	var i uint32
	for i = 0; i <= apiClientContext.Retry; i++ {
		response, err = methodFunc(apiClientContext, reqInfo)
		if err != nil {
			belog.Error("retry request (%v)", err)
			if apiClientContext.RetryWait > 0 {
				time.Sleep(time.Duration(apiClientContext.RetryWait) * time.Second)
			}
			continue
		}
		return response, err
	}
	return nil, errors.Errorf("give up retry (%v)", reqInfo.url)
}

func (c *Client) doRequest(methodFunc func(apiClientContext *contexter.APIClient, reqInfo *reqInfo) (response []byte, err error), reqInfo *reqInfo) (response []byte, err error) {
	apiClientContext := c.context.GetAPIClient()
	startEnd := [...]*startEnd{
		&startEnd{ start: c.urlBaseIndex, end: len(apiClientContext.APIServerURLList) },
		&startEnd{ start: 0, end: c.urlBaseIndex },
	}
	for _, startEnd := range startEnd  {
		for i := startEnd.start; i < startEnd.end; i++ {
			reqInfo.urlBase = apiClientContext.APIServerURLList[i].String()
			reqInfo.url = reqInfo.urlBase + reqInfo.resource
			response, err = c.retryRequest(apiClientContext, methodFunc, reqInfo)
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

// GetWatchResult is get watcher result
func (c *Client) GetWatchResult() (watchResultResponse *structure.WatchResultResponse, err error) {
	reqInfo := &reqInfo {
		resource : "/v1/watch/result",
	}
	response, err := c.doRequest(c.get, reqInfo)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not get watcher result (%v)", reqInfo.resource))
	}
	watchResultResponse = new(structure.WatchResultResponse)
        err = json.Unmarshal(response, watchResultResponse)
        if err != nil {
                return nil, errors.Wrap(err, fmt.Sprintf("can not unmarshal response (%v)", reqInfo.resource))
        }
	return watchResultResponse, nil
}

// New is create client
func New(context *contexter.Context) (*Client) {
        return &Client {
                urlBaseIndex:  0,
		context: context,
        }
}
