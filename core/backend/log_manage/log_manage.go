package log_manage

import (
	"operate-backend/core/backend/login"

	"github.com/gin-gonic/gin"
)

const ComponentName = "log_manage"

// Init 初始化
func Init(grouper *gin.RouterGroup) error {
	handler := grouper.Group(ComponentName)

	handler.POST(PathLoginQuery, login.SessionVerify, LoginQuery)
	handler.POST(PathOperateQuery, login.SessionVerify, OperateQuery)

	return nil
}

// Start 启动
func Start() error { return nil }

// Shutdown 关闭
func Shutdown() error { return nil }

// LoadCache 加载缓存
func LoadCache() error { return nil }
