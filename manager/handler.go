package manager

import (
        //"github.com/pkg/errors"
        //"github.com/potix/belog"
        "github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/sessions"
        //"github.com/potix/pdns-record-updater/api/structure"
        //"encoding/json"
	"html/template"
        "net/http"
	"path/filepath"
	"bytes"
        //"strings"
)

type message struct {
	message string
}

func (m Manager) saveSession(context *gin.Context, username string) {
	session := sessions.Default(context)
	session.Set("username", username)
	session.Save()
}

func (m Manager) loadSession(context *gin.Context) (string) {
	session := sessions.Default(context)
	username, ok := session.Get("username").(string)
	if !ok {
		return ""
	}
	return username
}

func (m Manager) clearSession(context *gin.Context) {
	session := sessions.Default(context)
	session.Clear()
	session.Save()
}

func (m Manager) checkSession(context *gin.Context) {
        managerContext := m.context.GetManager()
	username := m.loadSession(context)
	if username == "" || managerContext.Username != username {
		m.clearSession(context)
		m.replyFromAsset(context, filepath.Join("asset", "index.html"), &message{ message : "" } )
		context.Abort()
                return
	}
	context.Next()
}

func (m Manager) replyFromAsset(context *gin.Context, requestPath string, tmplData interface{}) {
	context.Header("Content-Type", gin.MIMEHTML)
	binData, err := Asset(requestPath)
	if err != nil {
		context.String(http.StatusNotFound, "404 not found " + requestPath)
		return
	}
	if context.Request.Method == http.MethodHead {
                context.Status(http.StatusOK)
		return
	}
	if tmplData == nil {
		context.Data(http.StatusOK, gin.MIMEHTML, binData)
		return
	}
	tmpl, err := template.New("template").Parse(string(binData))
	if err != nil {
		context.String(http.StatusInternalServerError, "500 internal server error")
		return
	}
	var page bytes.Buffer
	err = tmpl.Execute(&page, tmplData)
	if err != nil {
		context.String(http.StatusInternalServerError, "500 internal server error")
		return
	}
	context.String(http.StatusOK, page.String())
	return
}

func (m *Manager) index(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                fallthrough
        case http.MethodGet:
		m.replyFromAsset(context, filepath.Join("asset", "index.html"), &message{ message : "" } )
	}
}

func (m *Manager) asset(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                fallthrough
        case http.MethodGet:
		m.replyFromAsset(context, filepath.Join("asset", context.Request.URL.Path[1:]), nil)
	}
}

func (m *Manager) login(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodPost:
		username := context.PostForm("username")
		password := context.PostForm("password")
		managerContext := m.context.GetManager()
		if managerContext.Username != username || managerContext.Password != password {
			m.replyFromAsset(context, filepath.Join("asset", "index.html"), &message{ message : "<red>login failed</red><br>" } )
			return
		}
		m.saveSession(context, username)
		m.replyFromAsset(context, filepath.Join("asset", "mngmnt", "index.html"), nil)
	}
}

func (m *Manager) mngmnt(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                fallthrough
        case http.MethodGet:
		m.replyFromAsset(context, filepath.Join("asset", "mngmnt", "index.html"), nil)
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
