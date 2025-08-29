package tg_spider_account

import (
	"operate-backend/core/backend/login"
	"operate-backend/core/backend/middlewares"

	"github.com/gin-gonic/gin"
)

const ComponentName = "tg_spider_account"

// Init 初始化
func Init(grouper *gin.RouterGroup) error {
	handler := grouper.Group(ComponentName).Use(login.SessionVerify, middlewares.HiJack)

	handler.POST(PathInsert, Insert)
	handler.POST(PathUpdate, Update)
	handler.POST(PathQuery, Query)
	handler.POST(PathDelete, Delete)

	// 初始化 NatsJetStream 连接
	if err := InitSubscribe(); err != nil {
		return err
	}

	return nil
}

// Start 启动
func Start() {

	go NatsLoop()
}

// Shutdown 关闭
func Shutdown() error {
	if _subscription != nil {
		if err := _subscription.Unsubscribe(); err != nil {
			return err
		}
	}

	close(_close)

	<-_done

	return nil
}

// LoadCache 加载缓存
func LoadCache() error { return nil }
