package server

import (
	"github.com/potix/belog"
	"github.com/gin-gonic/gin"
	"github.com/potix/pdns-record-updater/api/structure"
	"encoding/json"
	"net/http"
)

func (s *Server) commonHandler(context *gin.Context) {
        context.Header("Content-Type", gin.MIMEJSON)
        switch context.Request.Method {
        case http.MethodHead:
		context.Status(http.StatusInternalServerError)
        case http.MethodGet:
		context.Status(http.StatusInternalServerError)
        case http.MethodPut:
		context.Status(http.StatusInternalServerError)
        default:
                context.Status(http.StatusMethodNotAllowed)
                return
        }
        context.Next()
}

func (s *Server) watchResult(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		context.Status(http.StatusNoContent)
        case http.MethodGet:
		newResult := &structure.Result {
			Zone : make(map[string]*structure.ZoneResult),
		}
		for zoneName, zone := range s.watcherContext.Zone {
			newZoneResult = &structure.ZoneResult {
					Nameserver : make([]*structure.RecordResult, 0, len(zone.NameServer)),
					Record : make([]*structure.RecordResult, 0, len(zone.DynamicRecord) + len(zone.NegativeRecord)),
					DynamicRecord : make([]*structure.DynamicRecordResult, 0, len(zone.DynamicGroup) * 10),
			}
			newResult.Zone[zone.Domain] = newZoneResult
			for _, record := range zone.NameServer {
				if record.Name == "" || record.Type == "" || record.TTL == 0 || record.Content == "" {
					continue
				}
				newRecordResult := &structure.recordResult {
					Name:    record.Name,
					Type:    record.Type,
					TTL:     record.TTL,
					Content: record.Content,
				}
				newZoneResult.NameServer = append(newZoneResult.NameServer, newRecordResult)
			}
			for _, record := range zone.Record {
				if record.Name == "" || record.Type == "" || record.TTL == 0 || record.Content == "" {
					continue
				}
				newRecordResult := &structure.recordResult {
					Name:    record.Name,
					Type:    record.Type,
					TTL:     record.TTL,
					Content: record.Content,
				}
				newZoneResult.Record = append(newZoneResult.Record, newRecordResult)
			}

			for _, dynamicGroup := range zone.DynamicGroup {
				var aliveRecordCount uint32
				for _, record := range dynamicGroup.DynamicRecord{
					if record.Name == "" || record.Type == "" || record.TTL == 0 || record.Content == "" {
						continue
					}
					newRecordResult := &structure.RecordResult {
						Name:    record.Name,
						Type:    record.Type,
						TTL:     record.TTL,
						Content: record.Content,
						Alive:   record.GetAlive(),
					}
					if record.GetForceDown() == 1 {
						newRecordResult.Alive = 0
					}
					if newRecordResult.Alive == 1 {
						aliveRecordCount++
					}
					newZoneResult.DynamicRecord = append(newZoneResult.DynamicRecord, newRecordResult)
				}
				var negativeRecordActive uint32
				if aliveRecordCount == 0 {
					negativeRecordActive = 1
				} else {
					negativeRecordActive = 0
				}
				for _, record := range dynamicGroup.NegativeRecord {
					if record.Name == "" || record.Type == "" || record.TTL == 0 || record.Content == "" {
						continue
					}
					newRecordResult := &structure.RecordResult {
						Name:    record.Name,
						Type:    record.Type,
						TTL:     record.TTL,
						Content: record.Content,
						Alive:   negativeRecordActive,
					}
					newZoneResult.DynamicRecord = append(newZoneResult.DynamicRecord, newRecordResult)
				}
			}
		}
		response, err := json.Marshal(newResult)
	        if err != nil {
	                belog.Error("can not marshal object with json")
			context.String(http.StatusInternalServerError, "")
	        } else {
			belog.Debug("response: %v", string(response))
		}
		context.Data(http.StatusOK, gin.MIMEJSON, response)
        case http.MethodPut:
                context.Status(http.StatusMethodNotAllowed)
	}
}

func (s *Server) record(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		context.Status(http.StatusNoContent)
        case http.MethodGet:
        case http.MethodPut:
		if err := context.BindJSON(&record); err != nil {
			context.Status(http.StatusBadRequest)
			return
		}




		context.Status(http.StatusNoContent)
	}
}
