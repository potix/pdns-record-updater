package manager

import (
        //"github.com/pkg/errors"
        //"github.com/potix/belog"
        "github.com/gin-gonic/gin"
        //"github.com/potix/pdns-record-updater/api/structure"
        //"encoding/json"
        "net/http"
	"path/filepath"
        //"strings"
)

func (m Manager) replyFromAsset(context *gin.Context, requestPath string) {
	context.Header("Content-Type", gin.MIMEHTML)
	data, err := Asset(requestPath)
	if err != nil {
		context.String(http.StatusNotFound, "404 not found " + requestPath)
		return
	}
	if context.Request.Method == http.MethodHead {
                context.Status(http.StatusOK)
		return
	}
	context.Data(http.StatusOK, gin.MIMEHTML, data)
	return
}

func (m *Manager) index(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                fallthrough
        case http.MethodGet:
		m.replyFromAsset(context, filepath.Join("asset", context.Request.URL.Path[1:]))
	}
}

func (m *Manager) asset(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                fallthrough
        case http.MethodGet:
		m.replyFromAsset(context, filepath.Join("asset", context.Request.URL.Path[1:]))
	}
}

func (m *Manager) login(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodPost:
	}
}

func (m *Manager) config(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                context.Status(http.StatusOK)
        case http.MethodGet:
        case http.MethodPost:
	}
}
