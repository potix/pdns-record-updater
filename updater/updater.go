package updater

import (
        "github.com/potix/pdns-record-updater/contexter"
        "github.com/potix/pdns-record-updater/api/client"
)

// Updater is updater
type Updater struct {
	client *client.Client
	updaterContext *contexter.Updater
}


//# Combined replacement of multiple RRsets
//curl -X PATCH --data '{"rrsets": [
//  {"name": "test1.example.org.",
//   "type": "A",
//   "ttl": 86400,
//   "changetype": "REPLACE",
//   "records": [ {"content": "192.0.2.5", "disabled": false} ]
//  },
//  {"name": "test2.example.org.",
//   "type": "AAAA",
//   "ttl": 86400,
//   "changetype": "REPLACE",
//   "records": [ {"content": "2001:db8::6", "disabled": false} ]
//  }
//  ] }' -H 'X-API-Key: changeme' http://127.0.0.1:8081/api/v1/servers/localhost/zones/example.org. | jq .



func (u *Updater) getApi(resource string) (error) {
	url := fmt.Sprintf("%v/%v", u.updaterContext.PdnsServer, resource)
        u, err := url.Parse(url)
        if err != nil {
                return errors.Errorf("can not parse url (%v)", url)
        }
        httpClient := helper.NewHTTPClient(u.Scheme, u.Host, false, 30)
        request, err := http.NewRequest("GET", url, nil)
        if err != nil {
                return errors.Wrap(err, fmt.Sprintf("can not create request (%v)", url))
        }
	request.Header.Set("Accept", "*/*")
	request.Header.Set("X-API-Key", u.updaterContext.PdnsAPIKey)
        res, err := httpClient.Do(request)
        if err != nil {
                return errors.Wrap(err, fmt.Sprintf("can not request (%v)", url))
        }
        defer res.Body.Close()
        if res.StatusCode != 200 && res.StatusCode != 204 {
                return errors.Wrap(err, fmt.Sprintf("unexpected status code (%v) (%v)", url, res.StatusCode))
        }
	if res.StatusCode == 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("can not read body (%v)", url))
		}
		belog.Debug("body: %v", body)
	}
        belog.Debug("http ok (%v)", url)
        return nil
}

func (u *Updater) postApi(resource string, request interface{}) (error) {
	url := fmt.Sprintf("%v/%v", u.updaterContext.PdnsServer, resource)
        u, err := url.Parse(url)
        if err != nil {
                return errors.Errorf("can not parse url (%v)", url)
        }
        httpClient := helper.NewHTTPClient(u.Scheme, u.Host, false, 30)
        jsonRequest, err := json.Marshal(request)
        if err != nil {
                return errors.Wrap(err, fmt.Sprintf("can not marsnale request (%v)", url))
	}
        request, err := http.NewRequest("POST", url, jsonRequest)
        if err != nil {
                return errors.Wrap(err, fmt.Sprintf("can not create request (%v)", url))
        }
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", u.updaterContext.PdnsAPIKey)
        res, err := httpClient.Do(jsonRequest)
        if err != nil {
                return errors.Wrap(err, fmt.Sprintf("can not request (%v)", url))
        }
        defer res.Body.Close()
        if res.StatusCode != 200 && res.StatusCode != 201 && res.StatusCode != 204 {
                return errors.Wrap(err, fmt.Sprintf("unexpected status code (%v) (%v)", url, res.StatusCode))
        }
	if res.StatusCode == 200 || res.StatusCode == 201 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("can not read body (%v)", url))
		}
		belog.Debug("body: %v", body)
	}
        belog.Debug("http ok (%v)", url)
        return nil
}

func (u *Updater) zoneUpdate(domain string, watchResultResponse *structure.WatchResultResponse) (error) {
	rrsetRequest, err := zoneWatcherResultResponseToRrsetRequest(watchResultResponse)
	// curl -v -H 'X-API-Key: api-key' http://127.0.0.1:18080/api/v1/servers/localhost/zones/domian.jp
}

func (u *Updater) zoneCreate(domain string, watchResultResponse *structure.WatchResultResponse) (error) {
	zoneRequest, err := zoneWatcherResultResponseToZoneRequest(watchResultResponse)
	// curl -v -H 'X-API-Key: api-key' http://127.0.0.1:18080/api/v1/servers/localhost/zones
}

func (u *Updater) getZone(domain string) (bool, error) {
	resource := strings.Sprintf("api/v1/servers/localhost/zones/%v", domain)
	u.callApi(resource, "GET", nil)
	// curl -v -H 'X-API-Key: api-key' http://127.0.0.1:18080/api/v1/servers/localhost/zones/example.com

}

func (u *Updater) updateLoop() () {
	for atomic.LoadUint32(&u.running) == 1 {
		var watchResultResponse *structure.WatchResultResponse
		for {
			if watchResultResponse, err = u.client.GetWatchResult(); err != nil {
				belog.Error("can not get watcher result (%v)", err)
				continue;
			}
			time.Sleep(time.Second)
			break
		}
		for domain, zoneWatchResultResponse := range watchResultResponse.Zone {
			// GETしてみる
			exist, err :=  u.getZone(domain)
			if exist {
				if err = u.updateZone(domain, zoneWatchResultResponse); err != nil {
					belog.Error("can not call api (%v)", err)
				}
			} else {
				if err = u.createZone(domain, zoneWatchResultResponse); err != nil {
					belog.Error("can not call api (%v)", err)
				}
			}
		}
	}
}

// Start is start
func (u Updater) Start() {
	atomic.StoreUint32(&u.running, 1)
        go u.updateLoop()
}

// Stop is stop
func (u Updater) Stop() {
	atomic.StoreUint32(&u.running, 0)
}

// New is create updater
func New(updaterContext *contexter.Updater, client *client.Client) (*Updater) {
        return &Updater {
                client:    client,
                updaterContext: updaterContext,
        }
}
