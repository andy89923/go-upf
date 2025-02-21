package sbi

import (
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/free5gc/go-upf/internal/logger"
	"github.com/free5gc/go-upf/pkg/factory"
	"github.com/free5gc/openapi/models"
	"github.com/free5gc/util/httpwrapper"
	logger_util "github.com/free5gc/util/logger"
)

type UPF interface {
	Config() *factory.Config
}

type Server struct {
	UPF

	httpServer *http.Server
	router     *gin.Engine
}

func NewServer(upf UPF, tlsKeyLogPath string) (*Server, error) {
	s := &Server{
		UPF:    upf,
		router: logger_util.NewGinWithLogrus(logger.SBILog),
	}
	s.ApplyService()

	cfg := upf.Config().GetSbiConfig()
	logger.SBILog.Infof("SBI Binding: %s:%d", cfg.BindingIp, cfg.Port)

	var err error
	if s.httpServer, err = httpwrapper.NewHttp2Server(cfg.BindingIp, tlsKeyLogPath, s.router); err != nil {
		return nil, err
	}
	s.httpServer.ReadHeaderTimeout = 3 * time.Second

	return s, nil
}

func (s *Server) ApplyService() {
	nwdafOamGroup := s.router.Group("/nwdaf-oam")
	nwdafOamRoutes := s.getNwdafOamRoutes()
	applyRoutes(nwdafOamGroup, nwdafOamRoutes)
}

func (s *Server) Start(ctx context.Context, wg *sync.WaitGroup) error {
	// TODO
	// register UPF to NRF

	wg.Add(1)
	go s.startServer(wg)

	return nil
}

func (s *Server) startServer(wg *sync.WaitGroup) {
	defer func() {
		if p := recover(); p != nil {
			logger.SBILog.Errorf("Recovered in Server.Start: %v", p)
		}
		wg.Done()
	}()

	var err error

	c := s.Config().GetSbiConfig()
	switch c.Scheme {
	case models.UriScheme_HTTP:
		err = s.httpServer.ListenAndServe()
	case models.UriScheme_HTTPS:
		err = s.httpServer.ListenAndServeTLS(c.Cert.Pem, c.Cert.Key)
	default:
		err = fmt.Errorf("Invalid SBI scheme: %s", c.Scheme)
	}
	if err != nil && err != http.ErrServerClosed {
		logger.SBILog.Errorf("HTTP server error: %v", err)
		return
	}
	logger.SBILog.Infof("HTTP server stopped")
}

func (s *Server) Stop() {
	// TODO
	// deregister UPF from NRF

	if s.httpServer != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := s.httpServer.Shutdown(ctx); err != nil {
			logger.SBILog.Errorf("HTTP server shutdown error: %v", err)
		}
	}
	logger.SBILog.Infof("UPF SBI Server terminated")
}
