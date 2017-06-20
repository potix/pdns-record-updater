package updater

import (
	"github.com/pkg/errors"
	"github.com/potix/belog"
        "github.com/potix/pdns-record-updater/contexter"
        "github.com/potix/pdns-record-updater/api/client"
	"github.com/potix/pdns-record-updater/api/structure"
	"github.com/potix/pdns-record-updater/helper"
        "sync/atomic"
	"encoding/json"
	"net/http"
	"net/url"
	"io/ioutil"
	"bytes"
	"strings"
	"time"
	"fmt"
	"sort"
)

// Updater is updater
type Updater struct {
	client         *client.Client
	updaterContext *contexter.Updater
	running        uint32
}

type recordData struct {
	Content  string `json:"content"`
	Disabled bool   `json:"disabled"`
}

type commentData struct {
	Content    string `json:"content"`
	Account    string `json:"account"`
	ModifiedAt int    `json:"modified_at"`
}

type rrsetData struct {
	Name        string         `json:"name"`
	Type        string         `json:"type"`
	TTL         int32          `json:"ttl"`
	ChangeType  string         `json:"changetype"`
	CommentList []*commentData `json:"comments"`
	RecordList  []*recordData  `json:"records"`
}

type rrsetRequest struct {
	Rrsets []*rrsetData `json:"rrsets"`
}

type zoneRequest struct {
	Name           string       `json:"name"`
	Kind           string       `json:"kind"`
	NameServerList []string     `json:"nameservers"`
	RrsetList      []*rrsetData `json:"rrsets"`
}

func (u *Updater) zoneWatcherResultResponseToZoneRequest(domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) (*zoneRequest, error) {
	zoneRequest := new(zoneRequest)
	zoneRequest.Name = helper.DotDomain(domain)
	zoneRequest.Kind = "NATIVE"
	zoneRequest.NameServerList = make([]string, 0) // NSレコードはrrsetsに含めるからここは空にする
	if len(zoneWatchResultResponse.NameServerList) == 0 {
		return nil, errors.Errorf("can not create soa, because no nameserver")
	}
	zoneRequest.RrsetList = u.zoneWatcherResultResponseToRrset(domain, zoneWatchResultResponse)
	return zoneRequest, nil
}

func (u *Updater) zoneWatcherResultResponseToRrset(domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) ([]*rrsetData) {
	rrsets := make([]*rrsetData, 0, 1 + len(zoneWatchResultResponse.NameServerList) + len(zoneWatchResultResponse.StaticRecordList) + len(zoneWatchResultResponse.DynamicRecordList))
	// soa
	soa := &rrsetData {
		Name:     helper.DotDomain(domain),
		Type:        "SOA",
		TTL:         3600,
		ChangeType:  "REPLACE",
		CommentList: make([]*commentData, 0),
		RecordList:  make([]*recordData, 0, 1),
	}
        soaMinimumTtls := i.updaterContent.SoaMinimumTtls
        if soaMinimumTtls == 0 {
                soaMinimumTtls = 60
        }
	record := &recordData {
		Content : fmt.Sprintf("%v %v 1 10800 3600 604800 %v", helper.DotHostname(zoneWatchResultResponse.PrimaryNameServer, domain), helper.DotEmail(zoneWatchResultResponse.Email), soaMinimumTtls),
		Disabled : false,
	}
	soa.RecordList = append(soa.RecordList, record)
	rrsets = append(rrsets, soa)
	// sort
	sort.Sort(zoneWatchResultResponse.NameServerList)
	sort.Sort(zoneWatchResultResponse.StaticRecordList)
	sort.Sort(zoneWatchResultResponse.DynamicRecordList)

	var latestRrset *rrsetData
	// ns record 
	latestRrset = nil
	for _, nameServer := range zoneWatchResultResponse.NameServerList {
		if nameServer.Type != "A" && nameServer.Type != "AAAA" {
			continue
		}
		name := helper.DotDomain(domain)
		if latestRrset == nil || latestRrset.Name != name {
			latestRrset = &rrsetData {
				Name:        name,
				Type:        "NS",
				TTL:         nameServer.TTL,
				ChangeType:  "REPLACE",
				CommentList: make([]*commentData, 0),
				RecordList:  make([]*recordData, 0, len(zoneWatchResultResponse.NameServerList)),
			}
			rrsets = append(rrsets, latestRrset)
		}
		content := helper.FixupRrsetContent(nameServer.Name, domain, "NS", true)
		record := &recordData {
			Content : content,
			Disabled : false,
		}
		latestRrset.RecordList = append(latestRrset.RecordList, record)
	}
	// name server
	latestRrset = nil
	for _, nameServer := range zoneWatchResultResponse.NameServerList {
		name := helper.FixupRrsetName(nameServer.Name, domain, nameServer.Type, true)
		rrsetType := strings.ToUpper(nameServer.Type)
		if latestRrset == nil || latestRrset.Name != name || latestRrset.Type != rrsetType {
			latestRrset = &rrsetData {
				Name:        name,
				Type:        rrsetType,
				TTL:         nameServer.TTL,
				ChangeType:  "REPLACE",
				CommentList: make([]*commentData, 0),
				RecordList:  make([]*recordData, 0, len(zoneWatchResultResponse.NameServerList)),
			}
			rrsets = append(rrsets, latestRrset)
		}
                content := helper.FixupRrsetContent(nameServer.Content, domain, nameServer.Type, true)
		record := &recordData {
			Content : content,
			Disabled : false,
		}
		latestRrset.RecordList = append(latestRrset.RecordList, record)
	}
	// static record
	latestRrset = nil
	for _, staticRecord := range zoneWatchResultResponse.StaticRecordList {
		name := helper.FixupRrsetName(staticRecord.Name, domain, staticRecord.Type, true)
		rrsetType := strings.ToUpper(staticRecord.Type)
		if latestRrset == nil || latestRrset.Name != name || latestRrset.Type != rrsetType {
			latestRrset = &rrsetData {
				Name:        name,
				Type:        rrsetType,
				TTL:         staticRecord.TTL,
				ChangeType:  "REPLACE",
				CommentList: make([]*commentData, 0),
				RecordList:  make([]*recordData, 0, len(zoneWatchResultResponse.StaticRecordList)),
			}
			rrsets = append(rrsets, latestRrset)
		}
                content := helper.FixupRrsetContent(staticRecord.Content, domain, staticRecord.Type, true)
		record := &recordData {
			Content : content,
			Disabled : false,
		}
		latestRrset.RecordList = append(latestRrset.RecordList, record)
	}
	// dynamic record
	latestRrset = nil
	for _, dynamicRecord := range zoneWatchResultResponse.DynamicRecordList {
		name := helper.FixupRrsetName(dynamicRecord.Name, domain, dynamicRecord.Type, true)
		rrsetType := strings.ToUpper(dynamicRecord.Type)
		if latestRrset == nil || latestRrset.Name != name || latestRrset.Type != rrsetType {
			latestRrset = &rrsetData {
				Name:        name,
				Type:        rrsetType,
				TTL:         dynamicRecord.TTL,
				ChangeType:  "REPLACE",
				CommentList: make([]*commentData, 0),
				RecordList:  make([]*recordData, 0, len(zoneWatchResultResponse.DynamicRecordList)),
			}
			rrsets = append(rrsets, latestRrset)
		}
                content := helper.FixupRrsetContent(dynamicRecord.Content, domain, dynamicRecord.Type, true)
		record := &recordData {
			Content : content,
			Disabled : !dynamicRecord.Alive,
		}
		latestRrset.RecordList = append(latestRrset.RecordList, record)
	}
	return rrsets
}

func (u *Updater) get(resource string) (int, error) {
        parsedURL, err := url.Parse(resource)
        if err != nil {
                return 0, errors.Errorf("can not parse url (%v)", resource)
        }
        httpClient := helper.NewHTTPClient(parsedURL.Scheme, parsedURL.Host, false, 30)
        request, err := http.NewRequest("GET", resource, nil)
        if err != nil {
                return 0, errors.Wrap(err, fmt.Sprintf("can not create request (%v)", resource))
        }
	request.Header.Set("Accept", "*/*")
	request.Header.Set("X-API-Key", u.updaterContext.PdnsAPIKey)
        res, err := httpClient.Do(request)
        if err != nil {
                return 0, errors.Wrap(err, fmt.Sprintf("can not request (%v)", resource))
        }
        defer res.Body.Close()
        if res.StatusCode != 200 && res.StatusCode != 204 {
                return res.StatusCode, errors.Errorf("unexpected status code (%v) (%v)", resource, res.StatusCode)
        }
	if res.StatusCode == 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return res.StatusCode, errors.Wrap(err, fmt.Sprintf("can not read body (%v)", resource))
		}
		belog.Debug("body: %v", string(body))
	}
        belog.Debug("http ok (%v)", resource)
        return res.StatusCode, nil
}

func (u *Updater) postPutPatch(resource string, method string, data interface{}) (error) {
        parsedURL, err := url.Parse(resource)
        if err != nil {
                return errors.Errorf("can not parse url (%v)", resource)
        }
        httpClient := helper.NewHTTPClient(parsedURL.Scheme, parsedURL.Host, false, 30)
        jsonData, err := json.Marshal(data)
        if err != nil {
                return errors.Wrap(err, fmt.Sprintf("can not marsnale request data (%v)", resource))
	}
	belog.Debug("request data: %v", string(jsonData))
        request, err := http.NewRequest(strings.ToUpper(method), resource, bytes.NewBuffer(jsonData))
        if err != nil {
                return errors.Wrap(err, fmt.Sprintf("can not create request (%v)", resource))
        }
	request.Header.Set("Accept", "*/*")
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-API-Key", u.updaterContext.PdnsAPIKey)
        res, err := httpClient.Do(request)
        if err != nil {
                return errors.Wrap(err, fmt.Sprintf("can not request (%v)", resource))
        }
        defer res.Body.Close()
        if res.StatusCode != 200 && res.StatusCode != 201 && res.StatusCode != 204 {
                return errors.Errorf("unexpected status code (%v) (%v)", resource, res.StatusCode)
        }
	if res.StatusCode == 200 || res.StatusCode == 201 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return errors.Wrap(err, fmt.Sprintf("can not read body (%v)", resource))
		}
		belog.Debug("body: %v", string(body))
	}
        belog.Debug("http ok (%v)", resource)
        return nil
}

func (u *Updater) updateZone(domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) (error) {
	rrsetRequest := &rrsetRequest {
		Rrsets : u.zoneWatcherResultResponseToRrset(domain, zoneWatchResultResponse),
	}
	resource := fmt.Sprintf("%v/api/v1/servers/localhost/zones/%v", u.updaterContext.PdnsServer, domain)
	return u.postPutPatch(resource, "PATCH", rrsetRequest)
}

func (u *Updater) createZone(domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) (error) {
	zoneRequest, err := u.zoneWatcherResultResponseToZoneRequest(domain, zoneWatchResultResponse)
	if err != nil {
		return err
	}
	resource := fmt.Sprintf("%v/api/v1/servers/localhost/zones", u.updaterContext.PdnsServer)
	return u.postPutPatch(resource, "POST", zoneRequest)
}

func (u *Updater) getZone(domain string) (bool, error) {
	resource := fmt.Sprintf("%v/api/v1/servers/localhost/zones/%v", u.updaterContext.PdnsServer, helper.NoDotDomain(domain))
	statusCode, err := u.get(resource)
	if err != nil {
		if statusCode == 0 || statusCode == 401 {
			return false, errors.Wrap(err, fmt.Sprintf("can not get zone (%v)", resource))
		} else if statusCode != 200 && statusCode != 204  {
			belog.Debug("%v", errors.Wrap(err, fmt.Sprintf("can not get zone (%v)", resource)))
			return false, nil
		} else {
			belog.Debug("%v", errors.Wrap(err, fmt.Sprintf("can not get zone (%v)", resource)))
			return true, nil
		}
	}
	return true, nil
}

func (u *Updater) updateLoop() () {
	for atomic.LoadUint32(&u.running) == 1 {
		var watchResultResponse *structure.WatchResultResponse
		var err error
		for {
			if watchResultResponse, err = u.client.GetWatchResult(); err != nil {
				belog.Error("can not get watcher result (%v)", err)
				continue;
			}
			break
		}
		for domain, zoneWatchResultResponse := range watchResultResponse.ZoneMap {
			exist, err :=  u.getZone(domain)
			if err != nil {
				belog.Error("can not get zone (%v)", err)
				continue
			}
			if exist {
				err = u.updateZone(domain, zoneWatchResultResponse)
				if err != nil {
					belog.Error("can not call api (%v)", err)
				}
			} else {
				err = u.createZone(domain, zoneWatchResultResponse)
				if err != nil {
					belog.Error("can not call api (%v)", err)
				}
			}
		}
		time.Sleep(time.Duration(u.updaterContext.UpdateInterval) * time.Second)
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
