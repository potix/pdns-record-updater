package server

import (
	"github.com/potix/belog"
	"github.com/gin-gonic/gin"
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
        default:
                context.Status(http.StatusMethodNotAllowed)
                return
        }
        context.Next()
}

type recordResult struct {
	Name    string
	Type    string
	Content string
	Alive   uint32
}

type zoneResult struct {
	Domain string
	Record []*recordResult
}

type result struct {
	Zone []*zoneResult
}

func (s *Server) watchResult(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		context.Status(http.StatusNoContent)
        case http.MethodGet:
		newResult := &result {
			Zone : make([]*zoneResult, 0, len(s.watcherContext.Zone)),
		}
		for _, zone := range s.watcherContext.Zone {
			newZoneResult := &zoneResult {
				Domain: zone.Domain,
				Record : make([]*recordResult, 0, len(zone.Record) + len(zone.NegativeRecord)),
			}
			newResult.Zone = append(newResult.Zone, newZoneResult)
			var aliveRecordCount uint32
			for _, record := range zone.Record {
				newRecordResult := &recordResult {
					Name:    record.Name,
					Type:    record.Type,
					Content: record.Content,
					Alive:   record.GetAlive(),
				}
				newZoneResult.Record = append(newZoneResult.Record, newRecordResult)
				if newRecordResult.Alive == 1 {
					aliveRecordCount++
				}
			}
			var negativeRecordActive uint32
			if aliveRecordCount == 0 {
				negativeRecordActive = 1
			} else {
				negativeRecordActive = 0
			}
			for _, record := range zone.NegativeRecord {
				newRecordResult := &recordResult {
					Name:    record.Name,
					Type:    record.Type,
					Content: record.Content,
					Alive:   negativeRecordActive,
				}
				newZoneResult.Record = append(newZoneResult.Record, newRecordResult)
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
	}
}
