package server

import (
	"github.com/pkg/errors"
	"github.com/potix/belog"
	"github.com/gin-gonic/gin"
	"github.com/potix/pdns-record-updater/contexter"
	"github.com/potix/pdns-record-updater/api/structure"
	"encoding/json"
	"net/http"
)

func (s *Server) commonHandler(context *gin.Context) {
        context.Header("Content-Type", gin.MIMEJSON)
        context.Status(http.StatusMethodNotAllowed)
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
		Zone : make(map[string]*structure.ZoneWatchResultResponse),
	}
	domain := s.contexter.Context.Watcher.GetDomain()
	for _, d := range domain {
		zone, err := s.contexter.Context.Watcher.GetZone(d)
		if err != nil {
			belog.Notice("%v", err)
			continue
		}
		newZoneWatchResultResponse := &structure.ZoneWatchResultResponse {
				NameServer : make([]*structure.NameServerRecordWatchResultResponse, 0),
				StaticRecord : make([]*structure.StaticRecordWatchResultResponse, 0),
				DynamicRecord : make([]*structure.DynamicRecordWatchResultResponse, 0),
		}
		newWatchResultResponse.Zone[d] = newZoneWatchResultResponse
		for _, record := range zone.GetNameServer() {
			if !record.Validate() {
				continue
			}
			newRecordWatchResultResponse := &structure.NameServerRecordWatchResultResponse {
				Name:    record.Name,
				Type:    record.Type,
				TTL:     record.TTL,
				Content: record.Content,
				Email:   record.Email,
			}
			newZoneWatchResultResponse.NameServer = append(newZoneWatchResultResponse.NameServer, newRecordWatchResultResponse)
		}
		for _, record := range zone.GetStaticRecord() {
			if !record.Validate() {
				continue
			}
			newRecordWatchResultResponse := &structure.StaticRecordWatchResultResponse {
				Name:    record.Name,
				Type:    record.Type,
				TTL:     record.TTL,
				Content: record.Content,
			}
			newZoneWatchResultResponse.StaticRecord = append(newZoneWatchResultResponse.StaticRecord, newRecordWatchResultResponse)
		}
		dynamicGroupName := zone.GetDynamicGroupName()
		for _, dgname := range dynamicGroupName {
			dynamicGroup, err := zone.GetDynamicGroup(dgname)
			if err != nil {
				belog.Notice("%v", err)
				continue
			}
			var aliveRecordCount uint32
			for _, record := range dynamicGroup.GetDynamicRecord() {
				if !record.Validate() {
					continue
				}
				newRecordWatchResultResponse := &structure.DynamicRecordWatchResultResponse {
					Name:    record.Name,
					Type:    record.Type,
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
				newZoneWatchResultResponse.DynamicRecord = append(newZoneWatchResultResponse.DynamicRecord, newRecordWatchResultResponse)
			}
			var negativeRecordAlive bool
			if aliveRecordCount == 0 {
				negativeRecordAlive = true
			} else {
				negativeRecordAlive = false
			}
			for _, record := range dynamicGroup.GetNegativeRecord() {
				if !record.Validate() {
					continue
				}
				newRecordWatchResultResponse := &structure.DynamicRecordWatchResultResponse {
					Name:    record.Name,
					Type:    record.Type,
					TTL:     record.TTL,
					Content: record.Content,
					Alive:   negativeRecordAlive,
				}
				newZoneWatchResultResponse.DynamicRecord = append(newZoneWatchResultResponse.DynamicRecord, newRecordWatchResultResponse)
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
		domain := s.contexter.Context.Watcher.GetDomain()
		s.jsonResponse(context, domain)
		return
        case http.MethodPost:
		var zoneRequest structure.ZoneRequest
		if err := context.BindJSON(&zoneRequest); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		if zoneRequest.Domain == "" {
			context.String(http.StatusBadRequest, "{\"reason\":\"no domain\"}")
			return
		}
		if err := s.contexter.Context.Watcher.AddZone(zoneRequest.Domain); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"%v\"}", err)
			return
		}
		context.Status(http.StatusCreated)
		return
	}
}

func (s *Server) zoneDomain(context *gin.Context) {
        switch context.Request.Method {
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
			nameServer := zone.GetNameServer()
			s.jsonResponse(context, nameServer)
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
		if !nameServer.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
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
		if !nameServer.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
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
			staticRecord := zone.GetStaticRecord()
			s.jsonResponse(context, staticRecord)
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
		if !staticRecord.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
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
		if !staticRecord.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
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
			dynamicGroupName := zone.GetDynamicGroupName()
			s.jsonResponse(context, dynamicGroupName)
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
			dynamicRecord := dynamicGroup.GetDynamicRecord()
			s.jsonResponse(context, dynamicRecord)
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
		if !dynamicRecord.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
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
		if !dynamicRecord.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
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
		dynamicRecord := dynamicGroup.FindDynamicRecord(n, t, c)
		if len(dynamicRecord) == 0  {
			context.String(http.StatusNotFound, "{\"reason\":\"not found\"}", err)
			return
		}
		if len(dynamicRecord) > 0  {
			context.String(http.StatusNotFound, "{\"reason\":\"match too many record\"}", err)
			return
		}
		var zoneDynamicGroupDynamicRecordForceDownRequest structure.ZoneDynamicGroupDynamicRecordForceDownRequest
		if err := context.BindJSON(&zoneDynamicGroupDynamicRecordForceDownRequest); err != nil {
			context.String(http.StatusBadRequest, "{\"reason\":\"can not unmarshal\"}")
			return
		}
		dynamicRecord[0].SetForceDown(zoneDynamicGroupDynamicRecordForceDownRequest.ForceDown)
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
			negativeRecord := dynamicGroup.GetNegativeRecord()
			s.jsonResponse(context, negativeRecord)
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
		if !negativeRecord.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
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
		if !negativeRecord.Validate() {
			context.String(http.StatusBadRequest, "{\"reason\":\"lack of parameter\"}")
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

