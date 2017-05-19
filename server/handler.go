package server

import (
	"github.com/potix/belog"
	"github.com/gin-gonic/gin"
	"encoding/json"
	"sync/atomic"
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
	RecordResult []*recordResult  `json:"Record"`
}

type watcherResult struct {
	ZoneResult   map[string]*zoneResult  `json:"Zone"`
}

func (s *Server) watcherResult(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
		context.Status(http.StatusNoContent)
        case http.MethodGet:
		newWatcherResult := &watcherResult {
			ZoneResult : make(map[string]*zoneResult),
		}
		for zoneName, zone := range s.watcherConfig.Zone {
			var aliveRecordCount uint32
			newRecordResult := make([]*recordResult, 0, len(zone.Record))
			for _, record := range zone.Record {
				r := &recordResult {
					Name:    record.Name,
					Type:    record.Type,
					Content: record.Content,
					Alive:   atomic.LoadUint32(&record.Alive),
				}
				if r.Alive == 1 {
					aliveRecordCount++
				}
				newRecordResult = append(newRecordResult, r)
			}
			if aliveRecordCount == 0 {
				for _, record := range zone.NegativeRecord {
					r := &recordResult {
						Name:    record.Name,
						Type:    record.Type,
						Content: record.Content,
						Alive:   1,
					}
					newRecordResult = append(newRecordResult, r)
                                }
                        }
			newWatcherResult.ZoneResult[zoneName] = &zoneResult {
				RecordResult : newRecordResult,
			}

		}
		response, err := json.Marshal(newWatcherResult)
	        if err != nil {
	                belog.Error("can not marshal object with json")
			context.String(http.StatusInternalServerError, "")
	        } else {
			belog.Debug("response: %v", response)
		}
		context.Data(http.StatusOK, gin.MIMEJSON, response)
	}
}
