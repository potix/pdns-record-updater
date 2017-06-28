package manager

import (
        //"github.com/pkg/errors"
        "github.com/potix/belog"
        "github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/sessions"
        //"github.com/potix/pdns-record-updater/api/structure"
        //"encoding/json"
	"html/template"
        "net/http"
	"path"
	"path/filepath"
	"mime"
	"bytes"
        "strings"
)

type message struct {
	Message string
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
		context.Redirect(http.StatusSeeOther, "../")
		context.Abort()
                return
	}
	context.Next()
}

func (m Manager) replyFromAsset(context *gin.Context, requestPath string, tmplData interface{}) {
	ext := path.Ext(requestPath)
	mimeType := mime.TypeByExtension(ext)
	context.Header("Content-Type", mimeType)
	binData, err := Asset(requestPath)
	if err != nil {
		belog.Error("not found (%v)", requestPath)
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
		belog.Error("can not create new template (%s)", string(binData))
		context.String(http.StatusInternalServerError, "500 internal server error")
		return
	}
	var page bytes.Buffer
	err = tmpl.Execute(&page, tmplData)
	if err != nil {
		belog.Error("can not create page from template (%v), (%v)", tmplData, string(binData) )
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
		m.replyFromAsset(context, filepath.Join("asset", "template", "index.html"), nil )
	}
}

func (m *Manager) asset(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                fallthrough
        case http.MethodGet:
		elems := strings.Split(context.Request.URL.Path, "/")
		if len(elems) == 0 {
			context.String(http.StatusNotFound, "400 bad request")
		}
		newElems := make([]string, 0, len(elems))
		newElems = append(newElems, "asset")
		newElems = append(newElems, elems[1:]...)
		m.replyFromAsset(context, filepath.Join(newElems...), nil)
	}
}

func (m *Manager) login(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodPost:
		username := context.PostForm("username")
		password := context.PostForm("password")
		managerContext := m.context.GetManager()
		if managerContext.Username != username || managerContext.Password != password {
			m.replyFromAsset(context, filepath.Join("asset", "template", "login.html"), &message{ Message : "login failed" } )
			return
		}
		m.saveSession(context, username)
		context.Redirect(http.StatusSeeOther, "./mngmnt/")
	}
}

func (m *Manager) logout(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodGet:
		m.clearSession(context)
		context.Redirect(http.StatusSeeOther, "./")
	}
}

func (m *Manager) mngmnt(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                fallthrough
        case http.MethodGet:
		//m.replyFromAsset(context, filepath.Join("asset", "template", "mngmnt", "index.html"), nil)
		// for debug
		context.Header("Content-Type", "text/html")
		context.File("./manager/asset/template/mngmnt/index.html")
	}
}

func (m *Manager) mngmntConfig(context *gin.Context) {
        switch context.Request.Method {
        case http.MethodHead:
                context.Status(http.StatusOK)
        case http.MethodGet:
        case http.MethodPost:
	}
}
