package ad

import (
	"operate-backend/core/backend/login"
	"operate-backend/core/backend/middlewares"

	"github.com/gin-gonic/gin"
)

const ComponentName = "ad"

// Init 初始化
func Init(grouper *gin.RouterGroup) error {
	handler := grouper.Group(ComponentName).Use(login.SessionVerify, middlewares.HiJack)

	handler.POST(PathInsert, Insert)
	handler.POST(PathUpdate, Update)
	handler.POST(PathQuery, Query)
	handler.POST(PathDelete, Delete)
	handler.POST(PathTotal, Total)
	// handler.POST(PathDeleteRole, DeleteRole)

	return nil
}

// Start 启动
func Start() error { return nil }

// Shutdown 关闭
func Shutdown() error { return nil }

// LoadCache 加载缓存
func LoadCache() error { return nil }
