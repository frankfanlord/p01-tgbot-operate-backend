package login

import (
	"operate-backend/core/backend/middlewares"

	"github.com/gin-gonic/gin"
)

// Init 初始化
func Init(grouper *gin.RouterGroup) error {

	grouper.POST(PathLogin, middlewares.HiJack, Login)
	grouper.POST(PathBindTFA, middlewares.HiJack, BindTFA)
	grouper.POST(PathUpdateMe, SessionVerify, middlewares.HiJack, UpdateMe)
	grouper.POST(PathMe, SessionVerify, middlewares.HiJack, Me)
	grouper.POST(PathMeMenus, SessionVerify, middlewares.HiJack, MeMenus)

	return nil
}

// Start 启动
func Start() error { return nil }

// Shutdown 关闭
func Shutdown() error { return nil }

// LoadCache 加载缓存
func LoadCache() error { return nil }
