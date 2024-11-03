package httpserver

import (
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GinServer struct {
	engine *gin.Engine
	option *HttpServerOption
}

func (s *GinServer) Start() error {
	return s.engine.Run(s.option.Address)
}

type OptionFunc func(*HttpServerOption)

type HttpServerOption struct {
	Pprof       bool
	SystemToken string
	Address     string
}

func NewGinServer(opt *HttpServerOption) *GinServer {
	router := gin.Default()
	adminGroup := router.Group("/_system", func(c *gin.Context) {
		if c.Request.Header.Get("Authorization") != opt.SystemToken {
			c.AbortWithStatus(http.StatusForbidden)
			return
		}
		c.Next()
	})
	if opt.Pprof {
		pprof.RouteRegister(adminGroup, "pprof")
	}
	return &GinServer{
		option: opt,
		engine: router,
	}
}

// NewHttpServerOption 创建并返回一个 HttpServerOption
func NewHttpServerOption(opts ...OptionFunc) *HttpServerOption {
	serverOption := &HttpServerOption{
		Address: ":8080",
	}
	for _, opt := range opts {
		opt(serverOption)
	}
	return serverOption
}

// WithPprof 将 HttpServerOption 的 Pprof 设置为 true
func WithPprof() OptionFunc {
	return func(o *HttpServerOption) {
		o.Pprof = true
	}
}

// WithAddrss 设置 HttpServerOption 的 Address
func WithAddrss(addr string) OptionFunc {
	return func(o *HttpServerOption) {
		o.Address = addr
	}
}
