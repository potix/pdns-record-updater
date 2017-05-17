package server

import (
        "github.com/pkg/errors"
        "github.com/potix/belog"
        "github.com/braintree/manners"
        "github.com/gin-gonic/gin"
        "net/http"
        "time"
)

// GracefulServer is GracefulServer
type GracefulServer struct {
	server *manners.GracefulServer
	startChan chan error
}

// Server is Server
type Server struct {
	gracefulServers []*GracefulServer
	serverConfig     *configurator.Server
}

func (s *Server) addHandlers(group *gin.RouterGroup, resource string, handler gin.HandlerFunc, flag addHandlersFlag) {
        group.HEAD(resource, handler)
        group.GET(resource, handler)
}

func (s *Server) startServer(gracefulServer *Gracefulserver) {
        err := gracefulServer.server.ListenAndServe()
        if err != nil {
                gracefulServer.startChan <- err
        }
}

// Start is Start
func (s *Server) Start() (err error) {
        if s.serverConfig.Listen == nil || len(s.serverConfig.Listen) == 0 {
                errors.Errorf("not found linten port")
        }
	engine := gin.Default()
	newGroup := engine.Group("/v1", c.commonHandler)
	s.addHandlers(newGroup, "/v1/watcher/result", s.watcherResult)

	// create server
        for _, listen := range s.serverConfig.Listen {
                server := manners.NewWithServer(&http.Server{
                        Addr:    listen.AddrPort,
                        Handler: engine,
                })
		newGracefulServer := &GracefulServer{
			server: server,
			startChan: make(chan err),
		}
                s.gracefulServers = append(s.gracefulServers, newGracefulServer)
        }

	// start server
        for _, gracefulServer := range s.gracefulServers {
                go c.startServer(gracefulServer)
                select {
                case err = <-gracefulServer.startChan:
			return err
		case <-time.After(time.Second):
			// ok
                }
        }
        return nil
}

// Stop is Stop
func (s *Server) Stop() {
        for _, gracefulServer := range s.gracefulServers {
                gracefulServer.BlockingClose()
        }
}

// New is create Server
func New(config *configurator.config) (s *Server) {
        return &Server{
		serverConfig:  config.Server,
        }
}
