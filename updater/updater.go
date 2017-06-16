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
)

// Updater is updater
type Updater struct {
	client         *client.Client
	updaterContext *contexter.Updater
	running        uint32
}

type record struct {
	Content  string
	Disabled bool
}

type comment struct {
	Content    string
	Account    string
	ModifiedAt int `json:"modified_at"`
}

type rrset struct {
	Name     string
	Type     string
	TTL      int32
	Comments []*comment
	Records  []*record
}

type rrsetRequest struct {
	Rrsets      []*rrset
}

type zoneRequest struct {
	Name        string
	Kind        string
	Nameservers []string
	Rrsets      []*rrset
}

func (u *Updater) zoneWatcherResultResponseToZoneRequest(domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) (*zoneRequest, error) {
	zoneRequest := new(zoneRequest)
	zoneRequest.Name = helper.DotDomain(domain)
	zoneRequest.Kind = "NATIVE"
	zoneRequest.Nameservers = make([]string, 0, len(zoneWatchResultResponse.NameServer))
	for _, nameServer := range zoneWatchResultResponse.NameServer {
		t := strings.ToUpper(nameServer.Type)
		if t != "A" || t != "AAA" {
			continue
		}
		zoneRequest.Nameservers = append(zoneRequest.Nameservers, helper.DotHostname(nameServer.Name, domain))
	}
	if len(zoneWatchResultResponse.NameServer) == 0 {
		return nil, errors.Errorf("can not create soa, because no nameserver")
	}
	rrsets := u.zoneWatcherResultResponseToRrset(domain, zoneWatchResultResponse)
	// create soa
	var primary *structure.NameServerRecordWatchResultResponse
	for _, nameServer := range zoneWatchResultResponse.NameServer {
		if nameServer.Type != "A" && nameServer.Type != "AAA" {
			continue
		}
		primary = nameServer
		break
	}
	if primary != nil {
		return nil, errors.Errorf("can not create soa, because no primary nameserver")
	}
	soa := &rrset{
		Name:     helper.DotDomain(domain),
		Type:     "SOA",
		TTL:      3600,
		Comments: make([]*comment, 0),
		Records: make([]*record, 0, 1),
	}
	record := &record {
		Content : fmt.Sprintf("%v %v 1 10800 3600 604800 60", helper.DotHostname(primary.Name, domain), helper.DotEmail(primary.Email)),
		Disabled : false,
	}
	soa.Records = append(soa.Records, record)
	rrsets = append(rrsets, soa)
	zoneRequest.Rrsets =  rrsets
	return zoneRequest, nil
}

func (u *Updater) zoneWatcherResultResponseToRrset(domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) ([]*rrset) {
	rrsets := make([]*rrset, 0, 1 + len(zoneWatchResultResponse.NameServer) + len(zoneWatchResultResponse.StaticRecord) + len(zoneWatchResultResponse.DynamicRecord))
	// name server
	for _, nameServer := range zoneWatchResultResponse.NameServer {
                name := helper.FixupRrsetName(nameServer.Name, domain, nameServer.Type, true)
                content := helper.FixupRrsetContent(nameServer.Content, domain, nameServer.Type, true)
		rrset := &rrset{
			Name:     name,
			Type:     nameServer.Type,
			TTL:      nameServer.TTL,
			Comments: make([]*comment, 0),
			Records: make([]*record, 0, 1),
		}
		record := &record {
			Content : content,
			Disabled : false,
		}
		rrset.Records = append(rrset.Records, record)
		rrsets = append(rrsets, rrset)
	}
	// static record
	for _, staticRecord := range zoneWatchResultResponse.StaticRecord {
                name := helper.FixupRrsetName(staticRecord.Name, domain, staticRecord.Type, true)
                content := helper.FixupRrsetContent(staticRecord.Content, domain, staticRecord.Type, true)
		rrset := &rrset{
			Name:     name,
			Type:     staticRecord.Type,
			TTL:      staticRecord.TTL,
			Comments: make([]*comment, 0),
			Records: make([]*record, 0, 1),
		}
		record := &record {
			Content : content,
			Disabled : false,
		}
		rrset.Records = append(rrset.Records, record)
		rrsets = append(rrsets, rrset)
	}
	// dynamic record
	for _, dynamicRecord := range zoneWatchResultResponse.DynamicRecord {
                name := helper.FixupRrsetName(dynamicRecord.Name, domain, dynamicRecord.Type, true)
                content := helper.FixupRrsetContent(dynamicRecord.Content, domain, dynamicRecord.Type, true)
		rrset := &rrset{
			Name:     name,
			Type:     dynamicRecord.Type,
			TTL:      dynamicRecord.TTL,
			Comments: make([]*comment, 0),
			Records: make([]*record, 0, 1),
		}
		record := &record {
			Content : content,
			Disabled : !dynamicRecord.Alive,
		}
		rrset.Records = append(rrset.Records, record)
		rrsets = append(rrsets, rrset)
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
	resource := fmt.Sprintf("%v/api/v1/servers/localhost/zones/%v", u.updaterContext.PdnsServer, domain)
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
			time.Sleep(time.Second)
			break
		}
		for domain, zoneWatchResultResponse := range watchResultResponse.Zone {
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
