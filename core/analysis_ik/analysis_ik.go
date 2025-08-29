package analysis_ik

import (
	"context"
	"errors"
	"jarvis/logger"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	_server *http.Server
)

func Init(addr, prefix string) error {
	gin.DefaultWriter = logger.GinWriter(logrus.Fields{"component": "analysis-ik"})
	gin.SetMode(gin.ReleaseMode)

	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger(), CORSMiddleware())

	handler := engine.Group(prefix)

	handler.HEAD("/ext_dict", ExtDict)
	handler.HEAD("/ext_stopWords", ExtStopWords)
	handler.GET("/ext_dict", ExtDict)
	handler.GET("/ext_stopWords", ExtStopWords)

	_server = &http.Server{Addr: addr, Handler: engine}

	return nil
}

func Start(channel chan<- any) {
	if _server == nil {
		channel <- errors.New("server is nil")
		return
	}

	channel <- nil

	if err := _server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.App().Errorf("failed to start server: %s", err.Error())
	}
}

func Shutdown() error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	defer cancel()
	return _server.Shutdown(ctx)
}

func LoadCache() error { return nil }

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type, Authorization, X-Requested-With, session, token")
		// c.Writer.Header().Set("Access-Control-Allow-Credentials", "true") // 支持携带 cookie 或自定义头部认证信息
		c.Writer.Header().Set("Access-Control-Max-Age", "86400") // 预检请求缓存时间（单位：秒）

		// 对于预检请求，直接返回 204
		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
