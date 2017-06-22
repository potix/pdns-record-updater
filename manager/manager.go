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
        "os"
        "path"
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
        managerContext   *contexter.Manager
        client           *client.Client
	execDir          string
}

func (m *Manager) addGetHandler(engine *gin.Engine, resource string, handler gin.HandlerFunc) {
        engine.HEAD(resource, handler)
        engine.GET(resource, handler)
}

func (m *Manager) addPostHandler(engine *gin.Engine, resource string, handler gin.HandlerFunc) {
        engine.POST(resource, handler)
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
func (s *Server) Start() (err error) {
        if m.managerContext.ListenList == nil || len(m.managerContext.ListenList) == 0 {
                errors.Errorf("not found linten port")
        }
	engine := gin.Default()
	store := sessions.NewCookieStore([]byte("secret"))
	engine.Use(sessions.Sessions("pdns-record-updater-session", store))

	// setup resource
	m.addGetHandler(engine, "/index.html", m.index)
	m.addGetHandler(engine, "/img", m.asset)
	m.addGetHandler(engine, "/js", m.asset)
	m.addGetHandler(engine, "/css", m.asset)
	m.addPostHandler(engine, "/login", m.login)
	m.addGetHandler(engine, "/config", m.config)
	m.addPostHandler(engine, "/config", m.config)
	m.addPostHandler(engine, "/zone", m.zone)
	m.addPostHandler(engine, "/record", m.record)
	if m.managerContext.LetsEncryptPath != "" {
		engine.Static("/.well-known", filepath.Join(m.managerContext.LetsEncryptPath, ".well-known"))
	}

	// create server
        for _, listen := range m.managerContext.ListenList {
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
                s.gracefulServers = append(s.gracefulServers, newGracefulServer)
        }

	// start server
        for _, gracefulServer := range s.gracefulServers {
                go s.startServer(gracefulServer)
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
func New(managerContext *contexter.Manager, client *client.Client) (*Server, error) {
	exec, err := os.Executable()
	if err != nil {
		errors.Wrap(err, "can not get executable path")
	}
	execDir := path.Dir(exec)
        s = &Manager{
                managerContext: managerContext,
                client: client,
		execDir: execDir,
        }
        if !managerContext.Debug {
                gin.SetMode(gin.ReleaseMode)
        }
        return s
}
