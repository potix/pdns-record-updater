package server

import (
	"github.com/pkg/errors"
	"github.com/potix/belog"
	"github.com/gin-gonic/gin"
	"github.com/potix/pdns-record-updater/contexter"
	"github.com/potix/pdns-record-updater/api/structure"
	"encoding/json"
	"crypto/sha256"
	"net/http"
	"time"
	"strings"
	"strconv"
	"fmt"
)

func (s *Server) authHandler(context *gin.Context) {
	// command example
	// TIME=$(date +%s) && curl -v -H "x-pdru-unixtime: ${TIME}" -H "Authroiztion: PDRU $(echo -n ${TIME}+API_KEY+GET+/v1/watch/result/  | sha256sum | awk '{print $1}')" http://127.0.0.1:28001/v1/watch/result
	authValue := context.Request.Header.Get("Authorization")
	unixTimeString := context.Request.Header.Get("x-pdru-unixtime")
	if authValue == "" {
		context.AbortWithStatus(http.StatusForbidden)
		return
	}
	if unixTimeString == "" {
		context.AbortWithStatus(http.StatusForbidden)
		return
	}
	unixTime, err := strconv.ParseInt(unixTimeString, 10, 64)
	if err != nil {
		context.AbortWithStatus(http.StatusForbidden)
		return
	}
	nowUnixTime := time.Now().Unix()
	diffTime := nowUnixTime - unixTime
	if diffTime < -600 || 600 < diffTime {
		context.AbortWithStatus(http.StatusForbidden)
		return
	}
	seedString := fmt.Sprintf("%v+%v+%v+%v", unixTimeString, s.context.GetAPIServer().APIKey, context.Request.Method, context.Request.URL.Path)
	if authValue != "PDRU " + fmt.Sprintf("%x", sha256.Sum256([]byte(seedString))) {
		context.AbortWithStatus(http.StatusForbidden)
		return
	}
        context.Next()
}

func (s *Server) commonHandler(context *gin.Context) {
        context.Header("Content-Type", gin.MIMEJSON)
        context.Next()
}

func (s *Server) jsonResponse(context *gin.Context, object interface{}) {
	jsonResponse, err := json.Marshal(object)
        if err != nil {
                belog.Error("can not marshal")
		context.String(http.StatusInternalServerError, "{\"reason\":\"can not marshal\"}")
        } else {
		belog.Debug("json response: %v", string(jsonResponse))
		context.Data(http.StatusOK, gin.MIMEJSON, jsonResponse)
	}
}

func (s *Server) contextToWatchResultResponse() (*structure.WatchResultResponse) {
	newWatchResultResponse := &structure.WatchResultResponse {
		ZoneMap : make(map[string]*structure.ZoneWatchResultResponse),
	}
	domainList := s.contexter.Context.Watcher.GetDomainList()
	for _, domain := range domainList {
		zone, err := s.contexter.Context.Watcher.GetZone(domain)
		if err != nil {
			belog.Notice("%v", err)
			continue
		}
		primaryNameServer, email := zone.GetPrimaryNameServerAndEmail()
		newZoneWatchResultResponse := &structure.ZoneWatchResultResponse {
				PrimaryNameServer: primaryNameServer,
				Email: email,
				NameServerList : make([]*structure.NameServerRecordWatchResultResponse, 0, len(zone.NameServerList)),
				StaticRecordList : make([]*structure.StaticRecordWatchResultResponse, 0, len(zone.StaticRecordList)),
				DynamicRecordList : make([]*structure.DynamicRecordWatchResultResponse, 0, 10 * len(zone.DynamicGroupMap)),
		}
		newWatchResultResponse.ZoneMap[domain] = newZoneWatchResultResponse
		for _, record := range zone.GetNameServerList() {
			newRecordWatchResultResponse := &structure.NameServerRecordWatchResultResponse {
				Name:    record.Name,
				Type:    strings.ToUpper(record.Type),
				TTL:     record.TTL,
				Content: record.Content,
			}
			newZoneWatchResultResponse.NameServerList = append(newZoneWatchResultResponse.NameServerList, newRecordWatchResultResponse)
		}
		for _, record := range zone.GetStaticRecordList() {
			newRecordWatchResultResponse := &structure.StaticRecordWatchResultResponse {
				Name:    record.Name,
				Type:    strings.ToUpper(record.Type),
				TTL:     record.TTL,
				Content: record.Content,
			}
			newZoneWatchResultResponse.StaticRecordList = append(newZoneWatchResultResponse.StaticRecordList, newRecordWatchResultResponse)
		}
		dynamicGroupNameList := zone.GetDynamicGroupNameList()
		for _, dynamicGroupName := range dynamicGroupNameList {
			dynamicGroup, err := zone.GetDynamicGroup(dynamicGroupName)
			if err != nil {
				belog.Notice("%v", err)
				continue
			}
			var aliveRecordCount uint32
			for _, record := range dynamicGroup.GetDynamicRecordList() {
				newRecordWatchResultResponse := &structure.DynamicRecordWatchResultResponse {
					Name:    record.Name,
					Type:    strings.ToUpper(record.Type),
					TTL:     record.TTL,
					Content: record.Content,
					Alive:   record.GetAlive(),
				}
				if record.GetForceDown() {
					newRecordWatchResultResponse.Alive = false
				}
				if newRecordWatchResultResponse.Alive {
					aliveRecordCount++
				}
				newZoneWatchResultResponse.DynamicRecordList = append(newZoneWatchResultResponse.DynamicRecordList, newRecordWatchResultResponse)
			}
			negativeRecordAlive := false
			if aliveRecordCount == 0 {
				negativeRecordAlive = true
			}
			for _, record := range dynamicGroup.GetNegativeRecordList() {
				newRecordWatchResultResponse := &structure.DynamicRecordWatchResultResponse {
					Name:    record.Name,
					Type:    strings.ToUpper(record.Type),
					TTL:     record.TTL,
					Content: record.Content,
					Alive:   negativeRecordAlive,
				}
				newZoneWatchResultResponse.DynamicRecordList = append(newZoneWatchResultResponse.DynamicRecordList, newRecordWatchResultResponse)
			}
		}
	}
	return newWatchResultResponse
}

func (s *Server) watchResult(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		context.Status(http.StatusOK)
		return
        case http.MethodGet:
		watchResultResponse := s.contextToWatchResultResponse()
		s.jsonResponse(context, watchResultResponse)
		return
	}
}

func (s *Server) config(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		fallthrough
        case http.MethodGet:
		dump, err := s.contexter.GetContext("json")
		if err != nil {
			context.String(http.StatusInternalServerError, "{\"reason\":\"%v\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			belog.Debug("json response: %v", string(dump))
			context.Data(http.StatusOK, gin.MIMEJSON, dump)
		}
		return
        case http.MethodPost:
		var configRequest structure.ConfigRequest
		if err := context.BindJSON(&configRequest); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if !configRequest.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		switch strings.ToUpper(configRequest.Action) {
		case "SAVE":
			err := s.contexter.SaveConfig()
			if err != nil {
				context.String(http.StatusInternalServerError, "{\"reason\":\"%v\"}", err)
				return
			}
			context.Status(http.StatusOK)
		case "LOAD":
			err := s.contexter.LoadConfig()
			if err != nil {
				context.String(http.StatusInternalServerError, "{\"reason\":\"%v\"}", err)
				return
			}
			context.Status(http.StatusOK)
		default:
			context.String(http.StatusBadRequest, "{\"reason\":\"unexpected action\"}")
			return
		}
	case http.MethodPut:
		newContext := new(contexter.Context)
		if err := context.BindJSON(newContext); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		err := s.contexter.PutContext(newContext)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
	}
}

func (s Server) getTarget(context *gin.Context) (*contexter.Target, error) {
	targetName := context.Param("tgname")
	if targetName == "" {
		return nil, errors.Errorf("lack of target name")
	}
	target, err := s.contexter.Context.Watcher.GetTarget(targetName)
	if err != nil {
		return nil, err
	}
	return target, nil
}

func (s *Server) target(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		context.Status(http.StatusOK)
		return
        case http.MethodGet:
		targetNameList := s.contexter.Context.Watcher.GetTargetNameList()
		s.jsonResponse(context, targetNameList)
		return
        case http.MethodPost:
		var targetRequest structure.TargetRequest
		if err := context.BindJSON(&targetRequest); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if !targetRequest.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		if targetRequest.TargetName == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"no domain\"}")
			return
		}
		newTarget := &contexter.Target {
			Protocol:           targetRequest.Protocol,
			Dest:               targetRequest.Dest,
			TCPTLS:             targetRequest.TCPTLS,
			HTTPMethod:         targetRequest.HTTPMethod,
			HTTPStatusList:     targetRequest.HTTPStatusList,
			Regexp:             targetRequest.Regexp,
			ResSize:            targetRequest.ResSize,
			Retry:              targetRequest.Retry,
			RetryWait:          targetRequest.RetryWait,
			Timeout:            targetRequest.Timeout,
			TLSSkipVerify:      targetRequest.TLSSkipVerify,
		}
		if err := s.contexter.Context.Watcher.AddTarget(targetRequest.TargetName, newTarget); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusCreated)
		return
	}
}

func (s *Server) targetTargetName(context *gin.Context) {
        switch context.Request.Method {
	case http.MethodHead:
		fallthrough
	case http.MethodGet:
		target, err := s.getTarget(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			s.jsonResponse(context, target)
		}
		return
        case http.MethodPut:
		target, err := s.getTarget(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		newTarget :=  new(contexter.Target)
		if err := context.BindJSON(target); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		target.Update(newTarget)
		context.Status(http.StatusOK)
		return
        case http.MethodDelete:
		targetName := context.Param("tgname")
		if targetName == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"no domain\"}")
			return
		}
		if err := s.contexter.Context.Watcher.DeleteTarget(targetName); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
		return
	}
}















func (s Server) getZone(context *gin.Context) (*contexter.Zone, error) {
	domain := context.Param("domain")
	if domain == "" {
		return nil, errors.Errorf("lack of domain")
	}
	zone, err := s.contexter.Context.Watcher.GetZone(domain)
	if err != nil {
		return nil, err
	}
	return zone, nil
}

func (s Server) getDynamicGroup(context *gin.Context) (*contexter.DynamicGroup, error) {
	zone, err := s.getZone(context)
	if err != nil {
		return nil, err
	}
	dgname := context.Param("dgname")
	if dgname == "" {
		return nil, errors.Errorf("lack of dynamic group name")
	}
	dynamicGroup, err := zone.GetDynamicGroup(dgname)
	if err != nil {
		return nil, err
	}
	return dynamicGroup, nil
}

func (s *Server) zone(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		context.Status(http.StatusOK)
		return
        case http.MethodGet:
		domainList := s.contexter.Context.Watcher.GetDomainList()
		s.jsonResponse(context, domainList)
		return
        case http.MethodPost:
		var zoneRequest structure.ZoneRequest
		if err := context.BindJSON(&zoneRequest); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if !zoneRequest.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		if zoneRequest.Domain == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"no domain\"}")
			return
		}
		newZone := &contexter.Zone {
			Email:              zoneRequest.Email,
			PrimaryNameServer:  zoneRequest.PrimaryNameServer,
			NameServerList:     make([]*contexter.NameServerRecord, 0),
			StaticRecordList:   make([]*contexter.StaticRecord, 0),
			DynamicGroupMap:    make(map[string]*contexter.DynamicGroup),
		}
		if err := s.contexter.Context.Watcher.AddZone(zoneRequest.Domain, newZone); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusCreated)
		return
	}
}

func (s *Server) zoneDomain(context *gin.Context) {
        switch context.Request.Method {
	case http.MethodHead:
		fallthrough
	case http.MethodGet:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			primaryNameServer, email := zone.GetPrimaryNameServerAndEmail()
			zoneDomainResponse := &structure.ZoneDomainResponse {
				PrimaryNameServer : primaryNameServer,
				Email : email,
			}
			s.jsonResponse(context, zoneDomainResponse)
		}
		return
        case http.MethodPut:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		var zoneDomainRequest structure.ZoneDomainRequest
		if err := context.BindJSON(&zoneDomainRequest); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if !zoneDomainRequest.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		zone.SetPrimaryNameServerAndEmail(zoneDomainRequest.PrimaryNameServer, zoneDomainRequest.Email)
		context.Status(http.StatusOK)
		return
        case http.MethodDelete:
		domain := context.Param("domain")
		if domain == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"no domain\"}")
			return
		}
		if err := s.contexter.Context.Watcher.DeleteZone(domain); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
		return
	}
}

func (s *Server) zoneNameServer(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		fallthrough
        case http.MethodGet:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			nameServerList := zone.GetNameServerList()
			s.jsonResponse(context, nameServerList)
		}
		return
        case http.MethodPost:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		nameServer := new(contexter.NameServerRecord)
		if err := context.BindJSON(nameServer); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if err := zone.AddNameServer(nameServer); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusCreated)
		return
	}
}

func (s *Server) zoneNameServerNTC(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		fallthrough
        case http.MethodGet:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		nameServer := zone.FindNameServer(n, t, c)
		if len(nameServer) == 0 {
			context.String(http.StatusNotFound, "{\"reason\":\"not found\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			s.jsonResponse(context, nameServer)
		}
		return
        case http.MethodPost:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		nameServer := new(contexter.NameServerRecord)
		if err := context.BindJSON(nameServer); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if err := zone.ReplaceNameServer(n, t, c, nameServer); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
		return
        case http.MethodDelete:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		if err := zone.DeleteNameServer(n, t, c); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
		return
	}
}

func (s *Server) zoneStaticRecord(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		fallthrough
        case http.MethodGet:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			staticRecordList := zone.GetStaticRecordList()
			s.jsonResponse(context, staticRecordList)
		}
		return
        case http.MethodPost:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		staticRecord := new(contexter.StaticRecord)
		if err := context.BindJSON(staticRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if err := zone.AddStaticRecord(staticRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusCreated)
		return
	}
}

func (s *Server) zoneStaticRecordNTC(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		fallthrough
        case http.MethodGet:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		staticRecord := zone.FindStaticRecord(n, t, c)
		if len(staticRecord) == 0 {
			context.String(http.StatusNotFound, "{\"reason\":\"not found\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			s.jsonResponse(context, staticRecord)
		}
		return
        case http.MethodPost:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		staticRecord := new(contexter.StaticRecord)
		if err := context.BindJSON(staticRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if err := zone.ReplaceStaticRecord(n, t, c, staticRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
		return
        case http.MethodDelete:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		if err := zone.DeleteStaticRecord(n, t, c); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
		return
	}
}

func (s *Server) zoneDynamicGroup(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		fallthrough
        case http.MethodGet:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			dynamicGroupNameList := zone.GetDynamicGroupNameList()
			s.jsonResponse(context, dynamicGroupNameList)
		}
		return
        case http.MethodPost:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		var zoneDynamicGroupRequest structure.ZoneDynamicGroupRequest
		if err := context.BindJSON(&zoneDynamicGroupRequest); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if !zoneDynamicGroupRequest.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		if zoneDynamicGroupRequest.DynamicGroupName == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"no dynamic group name\"}")
			return
		}

		if err := zone.AddDynamicGroup(zoneDynamicGroupRequest.DynamicGroupName); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusCreated)
		return
	}
}

func (s *Server) zoneDynamicGroupName(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodDelete:
		zone, err := s.getZone(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		dgname := context.Param("dgname")
		if  dgname == ""{
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		if err := zone.DeleteDynamicGroup(dgname); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
		return
	}
}

func (s *Server) zoneDynamicGroupDynamicRecord(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		fallthrough
        case http.MethodGet:
		dynamicGroup, err := s.getDynamicGroup(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			dynamicRecordList := dynamicGroup.GetDynamicRecordList()
			s.jsonResponse(context, dynamicRecordList)
		}
        case http.MethodPost:
		dynamicGroup, err := s.getDynamicGroup(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		dynamicRecord := new(contexter.DynamicRecord)
		if err := context.BindJSON(dynamicRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if err := dynamicGroup.AddDynamicRecord(dynamicRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusCreated)
	}
}

func (s *Server) zoneDynamicGroupDynamicRecordNTC(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		fallthrough
        case http.MethodGet:
		dynamicGroup, err := s.getDynamicGroup(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		dynamicRecord := dynamicGroup.FindDynamicRecord(n, t, c)
		if len(dynamicRecord) == 0 {
			context.String(http.StatusNotFound, "{\"reason\":\"not found\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			s.jsonResponse(context, dynamicRecord)
		}
		return
        case http.MethodPost:
		dynamicGroup, err := s.getDynamicGroup(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		dynamicRecord := new(contexter.DynamicRecord)
		if err := context.BindJSON(dynamicRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if err := dynamicGroup.ReplaceDynamicRecord(n, t, c, dynamicRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
		return
        case http.MethodDelete:
		dynamicGroup, err := s.getDynamicGroup(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		if err := dynamicGroup.DeleteDynamicRecord(n, t, c); err != nil{
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
		return
	}
}

func (s *Server) zoneDynamicGroupDynamicRecordNTCForceDown(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodPut:
		dynamicGroup, err := s.getDynamicGroup(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		dynamicRecordList := dynamicGroup.FindDynamicRecord(n, t, c)
		if len(dynamicRecordList) == 0  {
			context.String(http.StatusNotFound, "{\"reason\":\"not found\"}", err)
			return
		}
		var zoneDynamicGroupDynamicRecordForceDownRequest structure.ZoneDynamicGroupDynamicRecordForceDownRequest
		if err := context.BindJSON(&zoneDynamicGroupDynamicRecordForceDownRequest); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		for _, dynamicRecord :=  range dynamicRecordList {
			dynamicRecord.SetForceDown(zoneDynamicGroupDynamicRecordForceDownRequest.ForceDown)
		}
		context.Status(http.StatusOK)
		return
	}
}

func (s *Server) zoneDynamicGroupNegativeRecord(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		fallthrough
        case http.MethodGet:
		dynamicGroup, err := s.getDynamicGroup(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			negativeRecordList := dynamicGroup.GetNegativeRecordList()
			s.jsonResponse(context, negativeRecordList)
		}
        case http.MethodPost:
		dynamicGroup, err := s.getDynamicGroup(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		negativeRecord := new(contexter.NegativeRecord)
		if err := context.BindJSON(negativeRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if err := dynamicGroup.AddNegativeRecord(negativeRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusCreated)
	}
}

func (s *Server) zoneDynamicGroupNegativeRecordNTC(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		fallthrough
        case http.MethodGet:
		dynamicGroup, err := s.getDynamicGroup(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		negativeRecord := dynamicGroup.FindNegativeRecord(n, t, c)
		if len(negativeRecord) == 0 {
			context.String(http.StatusNotFound, "{\"reason\":\"not found\"}", err)
			return
		}
		if context.Request.Method == http.MethodHead {
			context.Status(http.StatusOK)
		} else {
			s.jsonResponse(context, negativeRecord)
		}
		return
        case http.MethodPost:
		dynamicGroup, err := s.getDynamicGroup(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		negativeRecord := new(contexter.NegativeRecord)
		if err := context.BindJSON(negativeRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if err := dynamicGroup.ReplaceNegativeRecord(n, t, c, negativeRecord); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
		return
        case http.MethodDelete:
		dynamicGroup, err := s.getDynamicGroup(context)
		if err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		n := context.Param("name")
		t := context.Param("type")
		c := context.Param("content")
		if n == "" || t == "" || c == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
			return
		}
		if err := dynamicGroup.DeleteNegativeRecord(n, t, c); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusOK)
		return
	}
}

