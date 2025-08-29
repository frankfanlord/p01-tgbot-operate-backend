package cog_manage

import (
	"operate-backend/core/backend/login"
	"operate-backend/core/backend/middlewares"

	"github.com/gin-gonic/gin"
)

const ComponentName = "cog"

// Init 初始化
func Init(grouper *gin.RouterGroup) error {
	handler := grouper.Group(ComponentName).Use(login.SessionVerify, middlewares.HiJack)

	handler.POST(PathInsert, Insert)
	handler.POST(PathUpdate, Update)
	handler.POST(PathQuery, Query)
	handler.POST(PathDelete, Delete)

	return nil
}

// Start 启动
func Start() error {

	go doCogDistribution()

	return nil
}

// Shutdown 关闭
func Shutdown() error {

	close(_close)

	return nil
}

// LoadCache 加载缓存
func LoadCache() error { return nil }
