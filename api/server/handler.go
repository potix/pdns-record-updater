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
		newWatchResult := &structure.WatchResultResponse {
			Zone : make(map[string]*structure.ZoneResultResponse),
		}
		for domain, zone := range s.watcherContext.Zone {
			newZoneResult := &structure.ZoneResultResponse {
					NameServer : make([]*structure.StaticRecordResultResponse, 0, len(zone.NameServer)),
					StaticRecord : make([]*structure.StaticRecordResultResponse, 0, len(zone.StaticRecord)),
					DynamicRecord : make([]*structure.DynamicRecordResultResponse, 0, len(zone.DynamicGroup) * 10),
			}
			newWatchResult.Zone[domain] = newZoneResult
			for _, record := range zone.NameServer {
				if record.Name == "" || record.Type == "" || record.TTL == 0 || record.Content == "" {
					continue
				}
				newRecordResult := &structure.StaticRecordResultResponse {
					Name:    record.Name,
					Type:    record.Type,
					TTL:     record.TTL,
					Content: record.Content,
				}
				newZoneResult.NameServer = append(newZoneResult.NameServer, newRecordResult)
			}
			for _, record := range zone.StaticRecord {
				if record.Name == "" || record.Type == "" || record.TTL == 0 || record.Content == "" {
					continue
				}
				newRecordResult := &structure.StaticRecordResultResponse {
					Name:    record.Name,
					Type:    record.Type,
					TTL:     record.TTL,
					Content: record.Content,
				}
				newZoneResult.StaticRecord = append(newZoneResult.StaticRecord, newRecordResult)
			}

			for _, dynamicGroup := range zone.DynamicGroup {
				var aliveRecordCount uint32
				for _, record := range dynamicGroup.DynamicRecord{
					if record.Name == "" || record.Type == "" || record.TTL == 0 || record.Content == "" {
						continue
					}
					newRecordResult := &structure.DynamicRecordResultResponse {
						Name:    record.Name,
						Type:    record.Type,
						TTL:     record.TTL,
						Content: record.Content,
						Alive:   record.GetAlive(),
					}
					if record.GetForceDown() {
						newRecordResult.Alive = false
					}
					if newRecordResult.Alive {
						aliveRecordCount++
					}
					newZoneResult.DynamicRecord = append(newZoneResult.DynamicRecord, newRecordResult)
				}
				var negativeRecordActive bool
				if aliveRecordCount == 0 {
					negativeRecordActive = true
				} else {
					negativeRecordActive = false
				}
				for _, record := range dynamicGroup.NegativeRecord {
					if record.Name == "" || record.Type == "" || record.TTL == 0 || record.Content == "" {
						continue
					}
					newRecordResult := &structure.DynamicRecordResultResponse {
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
		response, err := json.Marshal(newWatchResult)
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
		request := new()
		if err := context.BindJSON(&record); err != nil {
			context.Status(http.StatusBadRequest)
			return
		}




		context.Status(http.StatusNoContent)
	}
}
