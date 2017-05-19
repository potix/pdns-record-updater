package server

import (
        "github.com/pkg/errors"
        "github.com/braintree/manners"
        "github.com/gin-gonic/gin"
	"github.com/potix/pdns-record-updater/contexter"
        "net/http"
        "time"
        "fmt"
)

// GracefulServer is GracefulServer
type GracefulServer struct {
	server    *manners.GracefulServer
	startChan chan error
}

// Server is Server
type Server struct {
	gracefulServers []*GracefulServer
	serverContext   *contexter.Server
	watcherContext  *contexter.Watcher
}

func (s *Server) addHandlers(group *gin.RouterGroup, resource string, handler gin.HandlerFunc) {
        group.HEAD(resource, handler)
        group.GET(resource, handler)
}

func (s *Server) startServer(gracefulServer *GracefulServer) {
        err := gracefulServer.server.ListenAndServe()
        if err != nil {
                gracefulServer.startChan <- err
        }
}

// Start is Start
func (s *Server) Start() (err error) {
        if s.serverContext.Listen == nil || len(s.serverContext.Listen) == 0 {
                errors.Errorf("not found linten port")
        }
	engine := gin.Default()
	newGroup := engine.Group("/v1", s.commonHandler)
	s.addHandlers(newGroup, "/v1/watcher/result", s.watcherResult)

	// XXX TODO https
	// XXX TODO user password

	// create server
        for _, listen := range s.serverContext.Listen {
                server := manners.NewWithServer(&http.Server{
                        Addr:    listen.AddrPort,
                        Handler: engine,
                })
		newGracefulServer := &GracefulServer{
			server: server,
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
func New(context *contexter.Context) (s *Server) {
	s = &Server{
		serverContext: context.Server,
		watcherContext: context.Watcher,
        }
	if !context.Server.Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	return s
}
