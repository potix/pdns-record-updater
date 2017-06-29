package manager

import (
        "github.com/pkg/errors"
        "github.com/braintree/manners"
        "github.com/gin-gonic/gin"
	"github.com/gin-gonic/contrib/sessions"
        "github.com/potix/pdns-record-updater/contexter"
        "github.com/potix/pdns-record-updater/api/client"
        "net/http"
        "path/filepath"
        "time"
        "fmt"
)

// GracefulServer is GracefulServer
type GracefulServer struct {
        server    *manners.GracefulServer
        useTLS    bool
        certfile  string
        keyfile   string
        startChan chan error
}

// Manager is Manager
type Manager struct {
        gracefulServers  []*GracefulServer
        context          *contexter.Context
        client           *client.Client
}

func (m *Manager) addEngineGetHandler(engine *gin.Engine, resource string, handler gin.HandlerFunc) {
        engine.HEAD(resource, handler)
        engine.GET(resource, handler)
}

func (m *Manager) addEnginePostHandler(engine *gin.Engine, resource string, handler gin.HandlerFunc) {
        engine.POST(resource, handler)
}

func (m *Manager) addGroupGetHandler(group *gin.RouterGroup, resource string, handler gin.HandlerFunc) {
	        group.HEAD(resource, handler)
		group.GET(resource, handler)
}

func (m *Manager) addGroupPostHandler(group *gin.RouterGroup, resource string, handler gin.HandlerFunc) {
		group.POST(resource, handler)
}

func (m *Manager) startServer(gracefulServer *GracefulServer) {
        if gracefulServer.useTLS {
                err := gracefulServer.server.ListenAndServeTLS(gracefulServer.certfile, gracefulServer.keyfile)
                if err != nil {
                        gracefulServer.startChan <- err
                }
        } else {
                err := gracefulServer.server.ListenAndServe()
                if err != nil {
                        gracefulServer.startChan <- err
                }
        }
}

// Start is Start
func (m *Manager) Start() (err error) {
	managerContext := m.context.GetManager()
        if managerContext.ListenList == nil || len(managerContext.ListenList) == 0 {
                errors.Errorf("not found linten port")
        }
	engine := gin.Default()
	store := sessions.NewCookieStore([]byte("secret"))
	engine.Use(sessions.Sessions("pdns-record-updater-session", store))

	// setup resource
	m.addEngineGetHandler(engine, "/", m.index)                           // index
	m.addEngineGetHandler(engine, "/index.html", m.index)                 // index
	m.addEngineGetHandler(engine, "/bower_components/*wildcard", m.asset) // asset
	m.addEnginePostHandler(engine, "/login", m.login)                     // login
	m.addEngineGetHandler(engine, "/logout", m.logout)                    // logout
	newGroup := engine.Group("/mngmnt", m.checkSession)
	m.addGroupGetHandler(newGroup, "/", m.mngmntIndex)                    // management index
	m.addGroupGetHandler(newGroup, "/index", m.mngmntIndex)               // management index
	m.addGroupGetHandler(newGroup, "/s/*wildcard", m.mngmntStatic)        // management static
	m.addGroupGetHandler(newGroup, "/a/config", m.mngmntConfig)           // get config on memory
	m.addGroupPostHandler(newGroup, "/a/config", m.mngmntConfig)          // update config on memory / save config to disk / load config from disk
	if managerContext.LetsEncryptPath != "" {
		engine.Static("/.well-known", filepath.Join(managerContext.LetsEncryptPath, ".well-known"))
	}

	// create server
        for _, listen := range managerContext.ListenList {
                server := manners.NewWithServer(&http.Server{
                        Addr:    listen.AddrPort,
                        Handler: engine,
			ReadTimeout:    30 * time.Second,
			WriteTimeout:   30 * time.Second,
			MaxHeaderBytes: 1 << 20,
                })
		newGracefulServer := &GracefulServer{
			server: server,
			useTLS: listen.UseTLS,
			certfile: listen.CertFile,
			keyfile: listen.KeyFile,
			startChan: make(chan error),
		}
                m.gracefulServers = append(m.gracefulServers, newGracefulServer)
        }

	// start server
        for _, gracefulServer := range m.gracefulServers {
                go m.startServer(gracefulServer)
                select {
                case err = <-gracefulServer.startChan:
			return errors.Wrap(err, fmt.Sprintf("can not start server (%s)", gracefulServer.server.Addr))
		case <-time.After(time.Second):
			// ok
                }
        }
        return nil
}

// Stop is Stop
func (m *Manager) Stop() {
        for _, gracefulServer := range m.gracefulServers {
                gracefulServer.server.BlockingClose()
        }
}

// New is create manager
func New(context *contexter.Context, client *client.Client) (*Manager) {
	m := &Manager{
                context: context,
                client:  client,
        }
        if !context.GetManager().Debug {
                gin.SetMode(gin.ReleaseMode)
        }
        return m
}
