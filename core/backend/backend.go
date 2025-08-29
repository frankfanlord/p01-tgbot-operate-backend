package backend

import (
	"context"
	"errors"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/backend/ad_manage"
	"operate-backend/core/backend/cog_manage"
	"operate-backend/core/backend/component_ik"
	"operate-backend/core/backend/dashboard"
	"operate-backend/core/backend/log_manage"
	"operate-backend/core/backend/login"
	"operate-backend/core/backend/permission_manage"
	"operate-backend/core/backend/tg_spider_account"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var (
	_server *http.Server
)

func Init(addr, prefix string) error {
	gin.DefaultWriter = logger.GinWriter(logrus.Fields{"component": "operate-backend"})
	gin.SetMode(gin.DebugMode)

	engine := gin.New()
	engine.Use(gin.Recovery(), gin.Logger(), CORSMiddleware())

	grouper := engine.Group(prefix)

	// 注册ik分词路由
	if err := component_ik.Init(grouper); err != nil {
		return err
	}

	// 注册TG爬虫账号路由
	if err := tg_spider_account.Init(grouper); err != nil {
		return err
	}

	// 注册后台管理路由
	if err := permission_manage.Init(grouper); err != nil {
		return err
	}

	// 注册登录相关路由
	if err := login.Init(grouper); err != nil {
		return err
	}

	// 注册仪表盘相关路由
	if err := dashboard.Init(grouper); err != nil {
		return err
	}

	// 注册频道管理相关路由
	if err := cog_manage.Init(grouper); err != nil {
		return err
	}

	// 注册广告管理相关路由
	if err := ad_manage.Init(grouper); err != nil {
		return err
	}

	// 注册日志相关路由
	if err := log_manage.Init(grouper); err != nil {
		return err
	}

	_server = &http.Server{Addr: addr, Handler: engine}

	return nil
}

func Start(channel chan<- any) {
	if _server == nil {
		channel <- errors.New("server is nil")
		return
	}

	// TG爬虫启动消息中间件通讯
	tg_spider_account.Start()

	if err := cog_manage.Start(); err != nil {
		channel <- err
	}

	channel <- nil

	if err := _server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.App().Errorf("failed to start server: %s", err.Error())
	}
}

func Shutdown() error {
	// TG爬虫关闭消息中间件活动
	if err := tg_spider_account.Shutdown(); err != nil {
		return err
	}

	if err := cog_manage.Shutdown(); err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*time.Duration(5))
	defer cancel()
	return _server.Shutdown(ctx)
}

func LoadCache() error {
	// TG爬虫加载缓存
	if err := tg_spider_account.LoadCache(); err != nil {
		return err
	}

	return nil
}

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
