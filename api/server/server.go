package server

import (
        "github.com/pkg/errors"
        "github.com/braintree/manners"
        "github.com/gin-gonic/gin"
	"github.com/potix/pdns-record-updater/contexter"
	"github.com/potix/pdns-record-updater/configurator"
        "net/http"
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

// Server is Server
type Server struct {
	gracefulServers []*GracefulServer
	serverContext   *contexter.Server
	watcherContext  *contexter.Watcher
}

func (s *Server) addGetHandler(group *gin.RouterGroup, resource string, handler gin.HandlerFunc) {
        group.HEAD(resource, handler)
        group.GET(resource, handler)
}

func (s *Server) addPutHandler(group *gin.RouterGroup, resource string, handler gin.HandlerFunc) {
        group.PUT(resource, handler)
}

func (s *Server) startServer(gracefulServer *GracefulServer) {
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
        if s.serverContext.Listen == nil || len(s.serverContext.Listen) == 0 {
                errors.Errorf("not found linten port")
        }
	engine := gin.Default()
	if s.serverContext.username != "" && s.serverContext.password != "" {
		authHandler := gin.BasicAuth(gin.Accounts{s.serverContext.username, s.serverContext.password})
		newGroup := engine.Group("/v1", authHandler, s.commonHandler)
	} else {
		newGroup := engine.Group("/v1", s.commonHandler)
	}
	s.addGetHandler(newGroup, "/watch/result", s.watchResult)
	s.addGetHandler(newGroup, "/record", s.record)
	if s.serverContext.StaticPath != "" {
		newGroup.Static(newGroup, "/static", s.serverContext.StaticPath)
	}

	// create server
        for _, listen := range s.serverContext.Listen {
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
			certfile: listen.Certfile,
			keyfile: listen.Keyfile,
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
func (s *Server) Stop() {
        for _, gracefulServer := range s.gracefulServers {
                gracefulServer.server.BlockingClose()
        }
}

// New is create Server
func New(context *contexter.Context, configurator *configurator.Configurator) (s *Server) {
	s = &Server{
		configurator: configurator,
		serverContext: context.Server,
		watcherContext: context.Watcher,
        }
	if !context.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	return s
}
