package permission_manage

import (
	"operate-backend/core/backend/permission_manage/menu"
	"operate-backend/core/backend/permission_manage/role"
	"operate-backend/core/backend/permission_manage/user"

	"github.com/gin-gonic/gin"
)

const ComponentName = "admin"

// Init 初始化
func Init(grouper *gin.RouterGroup) error {
	handler := grouper.Group(ComponentName)

	if err := user.Init(handler); err != nil {
		return err
	}

	if err := menu.Init(handler); err != nil {
		return err
	}

	if err := role.Init(handler); err != nil {
		return err
	}

	return nil
}

// Start 启动
func Start() error { return nil }

// Shutdown 关闭
func Shutdown() error { return nil }

// LoadCache 加载缓存
func LoadCache() error { return nil }
