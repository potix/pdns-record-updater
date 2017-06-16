package updater

import (
	"github.com/pkg/errors"
	"github.com/potix/belog"
        "github.com/potix/pdns-record-updater/contexter"
        "github.com/potix/pdns-record-updater/api/client"
	"github.com/potix/pdns-record-updater/api/structure"
	"github.com/potix/pdns-record-updater/helper"
	"time"
	"fmt"
)

// Updater is updater
type Updater struct {
	client *client.Client
	updaterContext *contexter.Updater
}

type recordRequest struct {
	Content  string
	Disabled string
}

type commentRequest struct {
	Content    string
	Account    string
	ModifiedAt int `json:"modified_at"`
}

type rrsetRequest struct {
	Name     string
	Type     string
	TTL      string
	Comments []commentRequest
	Records  []recordRequest
}

type zoneRequest struct {
	Name        string
	Kind        string
	Nameservers []string
	Rrsets      []rrsetRequest
}

func (u *Updater) zoneWatcherResultResponseToZoneRequest(domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) (*zoneRequest, error) {
	zoneRequest := new(zoneRequest)
	zoneRequest.Name = helper.DotDomain(domain)
	zoneRequest.Kind = "NATIVE"
	zoneRequest.Nameservers = make([]string, 0, len(zoneWatchResultResponse.NameServer))
	for _, nameServer := range zoneWatchResultResponse.NameServer {
		if nameServer.Type != "A" || nameServer.Type != "AAA" {
			continue
		}
		zoneRequest.Nameservers = append(zoneRequest.Nameservers, helper.DotHostname(nameServer.Name, domain))
	}
	if len(zoneWatchResultResponse.Nameserver) == 0 {
		return nil, errors.Errorf("can not create soa, because no nameserver")
	}
	zoneRequest.Rrsets, err = u.zoneWatcherResultResponseToRrsetRequest(domain, zoneWatchResultResponse)
	if err != nil {
		return nil, err
	}
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
	soa := &rrsetRequest{
		Name:     helper.DotDomain(domain),
		Type:     "SOA",
		TTL:      3600,
		Comments: make(commentRequest, 0),
		Records: make(recordRequest, 0, 1),
	}
	record = &recordRequest {
		Content : fmt.Sprintf("%v %v 1 10800 3600 604800 60", helper.DotHostname(primary.Name, domain), helper.DotEmail(primary.Email)),
		Disabled : false,
	}
	soa.Records = append(soa.Records, record)
	zoneRequest.Rrsets = append(zoneRequest.Rrsets, soa)
	return zoneRequest, nil
}

func (u *Updater) zoneWatcherResultResponseToRrsetRequest(domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) ([]*rrsetRequest, error) {
	rrsets := make([]*rrsetRequest, 0, 1 + len(zoneWatchResultResponse.Nameserver) + len(zoneWatchResultResponse.StaticRecord) + len(zoneWatchResultResponse.DynamicRecord))
	// name server
	for _, nameServer := range zoneWatchResultResponse.NameServer {
                name := helper.FixupRrsetName(nameServer.Name, domain, nameServer.Type, true)
                content := helper.FixupRrsetContent(nameServer.Content, domain, nameServer.Type, true)
		rrset := &rrsetRequest{
			Name:     name,
			Type:     nameServer.Type,
			TTL:      nameServer.TTTL,
			Comments: make(commentRequest, 0),
			Records: make(recordRequest, 0, 1),
		}
		record = &recordRequest {
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
		rrset := &rrsetRequest{
			Name:     name,
			Type:     staticRecord.Type,
			TTL:      staticRecord.TTTL,
			Comments: make(commentRequest, 0),
			Records: make(recordRequest, 0, 1),
		}
		record = &recordRequest {
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
		rrset := &rrsetRequest{
			Name:     name,
			Type:     dynamicRecord.Type,
			TTL:      dynamicRecord.TTTL,
			Comments: make(commentRequest, 0),
			Records: make(recordRequest, 0, 1),
		}
		record = &recordRequest {
			Content : content,
			Disabled : !dynamicRecord.Alive,
		}
		rrset.Records = append(rrset.Records, record)
		rrsets = append(rrsets, rrset)
	}
}

func (u *Updater) get(resource string) (int, error) {
	url := fmt.Sprintf("%v/%v", u.updaterContext.PdnsServer, resource)
        u, err := url.Parse(url)
        if err != nil {
                return 0, errors.Errorf("can not parse url (%v)", url)
        }
        httpClient := helper.NewHTTPClient(u.Scheme, u.Host, false, 30)
        request, err := http.NewRequest("GET", url, nil)
        if err != nil {
                return 0, errors.Wrap(err, fmt.Sprintf("can not create request (%v)", url))
        }
	request.Header.Set("Accept", "*/*")
	request.Header.Set("X-API-Key", u.updaterContext.PdnsAPIKey)
        res, err := httpClient.Do(request)
        if err != nil {
                return 0, errors.Wrap(err, fmt.Sprintf("can not request (%v)", url))
        }
        defer res.Body.Close()
        if res.StatusCode != 200 && res.StatusCode != 204 {
                return res.StatusCode, errors.Wrap(err, fmt.Sprintf("unexpected status code (%v) (%v)", url, res.StatusCode))
        }
	if res.StatusCode == 200 {
		body, err := ioutil.ReadAll(res.Body)
		if err != nil {
			return res.StatusCode, errors.Wrap(err, fmt.Sprintf("can not read body (%v)", url))
		}
		belog.Debug("body: %v", body)
	}
        belog.Debug("http ok (%v)", url)
        return res.StatusCode, nil
}

func (u *Updater) postPutPatch(resource string, method string, request interface{}) (error) {
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
        request, err := http.NewRequest(strings.ToUpper(method), url, jsonRequest)
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

func (u *Updater) zoneUpdate(domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) (error) {
	rrsetRequest, err := zoneWatcherResultResponseToRrsetRequest(zoneWatchResultResponse)
	resource := fmt.Sprintf("api/v1/servers/localhost/zones/%v", domain)
	return u.postPutPatch(resource, "PATCH", rrserRequest)
}

func (u *Updater) zoneCreate(domain string, zoneWatchResultResponse *structure.ZoneWatchResultResponse) (error) {
	zoneRequest, err := zoneWatcherResultResponseToZoneRequest(domain, zoneWatchResultResponse)
	return u.postPutPatch("api/v1/servers/localhost/zones", "POST", zoneRequest)
}

func (u *Updater) getZone(domain string) (bool, error) {
	resource := fmt.Sprintf("api/v1/servers/localhost/zones/%v", domain)
	statusCode, err := u.get(resource)
	if err != nil {
		if res.Stauscode == 0 {
			return false, errors.Warp(err, fmt.Sprintf("can not get api (%v)", resource))
		} else if res.StatusCode != 200 && res.StatusCode != 204 {
			belog.Debug("%v", errors.Warp(err, fmt.Sprintf("can not get api (%v)", resource)))
			return false, nil
		} else {
			belog.Debug("%v", errors.Warp(err, fmt.Sprintf("can not get api (%v)", resource)))
			return true, nil
		}
	}
	return true, nil
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
			exist, err :=  u.getZone(domain)
			if err != nil {
				belog.Error("can not get zone (%v)", err)
				continue
			}
			if exist {
				err := u.updateZone(domain, zoneWatchResultResponse)
				if err != nil {
					belog.Error("can not call api (%v)", err)
				}
			} else {
				err := u.createZone(domain, zoneWatchResultResponse)
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
