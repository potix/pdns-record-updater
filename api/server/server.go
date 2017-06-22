package server

import (
        "github.com/pkg/errors"
        "github.com/braintree/manners"
        "github.com/gin-gonic/gin"
	"github.com/potix/pdns-record-updater/contexter"
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

// Server is Server
type Server struct {
	gracefulServers []*GracefulServer
	context         *contexter.Context
	contexter       *contexter.Contexter
}

func (s *Server) addGetHandler(group *gin.RouterGroup, resource string, handler gin.HandlerFunc) {
        group.HEAD(resource, handler)
        group.GET(resource, handler)
}

func (s *Server) addPostHandler(group *gin.RouterGroup, resource string, handler gin.HandlerFunc) {
        group.POST(resource, handler)
}

func (s *Server) addPutHandler(group *gin.RouterGroup, resource string, handler gin.HandlerFunc) {
        group.PUT(resource, handler)
}

func (s *Server) addDeleteHandler(group *gin.RouterGroup, resource string, handler gin.HandlerFunc) {
        group.DELETE(resource, handler)
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
	apiServerContext := s.context.GetAPIServer()
        if apiServerContext.ListenList == nil || len(apiServerContext.ListenList) == 0 {
                errors.Errorf("not found linten port")
        }
	engine := gin.Default()
	var newGroup *gin.RouterGroup

	// set up resource
	if apiServerContext.APIKey != "" {
		newGroup = engine.Group("/v1", s.authHandler, s.commonHandler)
	} else {
		newGroup = engine.Group("/v1", s.commonHandler)
	}
	s.addGetHandler(newGroup, "/watch/result", s.watchResult) // 監視結果取得
	s.addGetHandler(newGroup, "/config", s.config) // 設定取得
	s.addPostHandler(newGroup, "/config", s.config) // 設定読み込み、保存
	s.addPutHandler(newGroup, "/config", s.config) // replace config
	s.addGetHandler(newGroup, "/zone", s.zone)  // ゾーン一覧取得
	s.addPostHandler(newGroup, "/zone", s.zone)  // ゾーン作成
	s.addGetHandler(newGroup, "/zone/:domain", s.zoneDomain)  // ゾーン情報取得
	s.addPutHandler(newGroup, "/zone/:domain", s.zoneDomain)  // ゾーン情報変更
	s.addDeleteHandler(newGroup, "/zone/:domain", s.zoneDomain)  // ゾーン削除
	s.addGetHandler(newGroup, "/zone/:domain/nameserver", s.zoneNameServer)  // ネームサーバ一覧取得
	s.addPostHandler(newGroup, "/zone/:domain/nameserver", s.zoneNameServer) // ネームサーバ作成
	s.addGetHandler(newGroup, "/zone/:domain/nameserver/:name/:type/:Content", s.zoneNameServerNTC)        // ネームサーバ取得
	s.addPostHandler(newGroup, "/zone/:domain/nameserver/:name/:type/:Content", s.zoneNameServerNTC)       // ネームサーバ変更
	s.addDeleteHandler(newGroup, "/zone/:domain/nameserver/:name/:type/:Content", s.zoneNameServerNTC)     // ネームサーバ削除
	s.addGetHandler(newGroup, "/zone/:domain/staticrecord", s.zoneStaticRecord)  // 静的コード一覧取得
	s.addPostHandler(newGroup, "/zone/:domain/staticrecord", s.zoneStaticRecord) // 静的レコード作成 
	s.addGetHandler(newGroup, "/zone/:domain/staticrecord/:name/:type/:Content", s.zoneStaticRecordNTC)    // 静的レコード取得
	s.addPostHandler(newGroup, "/zone/:domain/staticrecord/:name/:type/:Content", s.zoneStaticRecordNTC)   // 静的レコード変更
	s.addDeleteHandler(newGroup, "/zone/:domain/staticrecord/:name/:type/:Content", s.zoneStaticRecordNTC) // 静的レコード削除
	s.addGetHandler(newGroup, "/zone/:domain/dynamicgroup", s.zoneDynamicGroup)  // 動的グループ一覧取得
	s.addPostHandler(newGroup, "/zone/:domain/dynamicgroup", s.zoneDynamicGroup) // 動的グループ作成
	s.addDeleteHandler(newGroup, "/zone/:domain/dynamicgroup/:dgname", s.zoneDynamicGroupName) // 動的グループ削除
	s.addGetHandler(newGroup, "/zone/:domain/dynamicgroup/:dgname/dynamicrecord", s.zoneDynamicGroupDynamicRecord)  // 動的レコードの一覧を取得
	s.addPostHandler(newGroup, "/zone/:domain/dynamicgroup/:dgname/dynamicrecord", s.zoneDynamicGroupDynamicRecord) // 動的レコードの作成 
	s.addGetHandler(newGroup, "/zone/:domain/dynamicgroup/:dgname/dynamicrecord/:name/:type/:Content", s.zoneDynamicGroupDynamicRecordNTC)                    // 動的レコードの取得
	s.addPostHandler(newGroup, "/zone/:domain/dynamicgroup/:dgname/dynamicrecord/:name/:type/:Content", s.zoneDynamicGroupDynamicRecordNTC)                   // 動的レコードの変更
	s.addPutHandler(newGroup, "/zone/:domain/dynamicgroup/:dgname/dynamicrecord/:name/:type/:Content/forcedown", s.zoneDynamicGroupDynamicRecordNTCForceDown) // 動的レコードの変更
	s.addDeleteHandler(newGroup, "/zone/:domain/dynamicgroup/:dgname/dynamicrecord/:name/:type/:Content", s.zoneDynamicGroupDynamicRecordNTC)                 // 動的レコードの削除
	s.addGetHandler(newGroup, "/zone/:domain/dynamicgroup/:dgname/negativerecord", s.zoneDynamicGroupNegativeRecord)  // ネガティブレコードの一覧取得
	s.addPostHandler(newGroup, "/zone/:domain/dynamicgroup/:dgname/negativerecord", s.zoneDynamicGroupNegativeRecord) // ネガティブレコードの作成
	s.addPostHandler(newGroup, "/zone/:domain/dynamicgroup/:dgname/negativerecord/:name/:type/:Content", s.zoneDynamicGroupNegativeRecordNTC)   // ネガティブレコードの変更
	s.addDeleteHandler(newGroup, "/zone/:domain/dynamicgroup/:dgname/negativerecord/:name/:type/:Content", s.zoneDynamicGroupNegativeRecordNTC) // ネガティブレコードの削除
	if apiServerContext.LetsEncryptPath != "" {
		engine.Static("/.well-known", filepath.Join(apiServerContext.LetsEncryptPath, ".well-known"))
	}

	// create server
        for _, listen := range apiServerContext.ListenList {
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
func (s *Server) Stop() {
        for _, gracefulServer := range s.gracefulServers {
                gracefulServer.server.BlockingClose()
        }
}

// New is create Server
func New(context *contexter.Context, contexter *contexter.Contexter) (*Server) {
	s := &Server{
		context: context,
		contexter: contexter,
        }
	if !context.GetAPIServer().Debug {
		gin.SetMode(gin.ReleaseMode)
	}
	return s
}
