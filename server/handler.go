package server

import (
	"github.com/pkg/errors"
	"github.com/potix/belog"
	"github.com/gin-gonic/gin"
	"sync/atomic"
)

type recordResult struct {
	Name    string
	Type    string
	Content string
	Alive   uint32
}

type wathcerResult struct {
	RecordResult []*recordResult
}

func (s *Server) watcherResult(context *gin.Context) {
	switch context.Request.Method {
	case http.MethodHead:
		context.String(http.StatusNoContent, nil)
	case http.MethodGet:
		newWathcerResult := new(watcherResult)
		for _, record := range s.config.Wacther.Record {
			newRecordResult := &recordResult {
				Name:    record.Name,
				Type:    record.Type,
				Content: record.Content,
				Alive:   atomic.LoadUint32(&record.Alive),
			}
			newWatcherResult.recordResult = append(newWatcherResult.recordResult, newRecordResult)
		}
		context.JSON(newWatcherResult, newWatcherResult)
	}
}
