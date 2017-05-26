package client

import (
	"github.com/pkg/errors"
        "github.com/potix/belog"
	"github.com/potix/pdns-record-updater/api/structure"
	"net/http"
	"net/url"
	"io/ioutil"
	"fmt"
)

type request struct {
	urlBase string
	resource string
}

type startEnd struct {
	start int
	end   int
}

// Client is client
type Client struct {
	urlBaseIndex  int
        urlBase       []string
	retry         uint32
	retryWait     uint32
        timeout       uint32
	tlsSkipVerify bool
	username      string
	password      string
}

func (c *Client) get(request *request) ([]byte, error) {
        u, err := url.Parse(request.url)
	if err != nil {
		return  nul, errors.Errorf("can not parse url (%v)", request.url)
	}
	httpClient := NewHTTPClient(u.Scheme, u.Host, c.tlsSkipVerify)
	request, err := http.NewRequest("GET", request.url, nil)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not create request (%v)", request.url))
	}
	request.SetBasicAuth(c.username, c.password)
	res, err := httpClient.Do(request)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not get url (%v)", request.url))
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		return nil, errors.Wrap(err, fmt.Sprintf("unexpected status code (%v) (%v)", request.url, res.StatusCode))
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf(" (%v)", request.url))
	}
	belog.Debug("http ok (%v)", request.url)
	return body, nil
}

func (c *Client) retry(methodFunc func(request *request), request *request) (response []byte, err error)  {
	for i = 0; i <  c.retry; i++ {
		response, err = methodFunc(request)
		if err != nil {
			belog.Error("retry request (%v)", err)
			if c.retryWait > 0 {
				time.Sleep(time.Duration(c.retryWait) * time.Second)
			}
			continue
		}
		return response, err
	}
	return nil, erros.Errorf("give up retry (%v)", request.url)
}

func (c *Client) doRequest(methodFunc func(request *request), request *request) (response []byte, err error) {
	startEnd := [...]*startEnd{
		&startEnd{ start: c.urlBaseIndex, end: len(c.urlBase) },
		&startEnd{ start: 0, end: c.urlBaseIndex },
	}
Loop:
	for _, startEnd := range startEnd  {
		for i := startEnd.start; i < startEnd.end; i++ {
			request.urlBase = c.urlBase[i]
			request.url = request.urlBase + request.resource
			response, err = c.retry(methodFunc, request)
			if err != nil {
				belog.Error("switch url base (%v)", err)
				continue
			}
			c.urlBaseIndex = i
			return resonse, err
		}
	}
	return nil, erros.Errorf("give up request (%v)", request.resource)
}

// GetWatcherResult is get watcher result
func (c *Client) GetWatcherResult() (result *structure.Result, err error) {
	request := &request {
		resource : "/v1/watcher/result",
	}
	response, err := c.doRequest(c.get, request)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("can not get watcher result (%v)", resource))
	}
	result = make(structure.Result)
        err = json.Unmarshal(response, result)
        if err != nil {
                return nil, errors.Wrap(err, fmt.Sprintf("can not unmarshal response (%v)", resource))
        }

	return result, nil
}

// New is create client
func New(client *contexter.Client) (*Client) {
        return &Client {
                urlBaseIndex:  0,
                urlBase:       client.URLBase,
                retry:         client.Retry,
                retryWait:     client.RetryWait,
                timeout:       client.Timeout,
		tlsSkipVerify: client.TLSSkipVerify,
		username:      client.Username,
		password:      client.Password,
        }
}
