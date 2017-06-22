package manager

import (
        "github.com/pkg/errors"
        "github.com/potix/belog"
        "github.com/gin-gonic/gin"
        "github.com/potix/pdns-record-updater/api/structure"
        "encoding/json"
        "net/http"
        "strings"
)

func (m *Manager) index(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                context.Status(http.StatusOK)
        case http.MethodGet:
		context.Header("Content-Type", gin.MIMEHTML)
		context.String(http.StatusOK, index)
	default:
		context.Status(http.StatusMethodNotAllowed)
	}
}

func (m *Manager) asset(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                context.Status(http.StatusOK)
        case http.MethodGet:
	default:
		context.Status(http.StatusMethodNotAllowed)
	}
}

func (m *Manager) login(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodPost:
	default:
		context.Status(http.StatusMethodNotAllowed)
	}
}

func (m *Manager) config(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                context.Status(http.StatusOK)
        case http.MethodGet:
        case http.MethodPost:
	default:
		context.Status(http.StatusMethodNotAllowed)
	}
}
