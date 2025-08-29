package user

import (
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"operate-backend/core/backend/login"
	"operate-backend/core/backend/middlewares"
	"operate-backend/core/structure"

	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/duke-git/lancet/v2/random"
	"github.com/gin-gonic/gin"
)

const ComponentName = "user"

// Init 初始化
func Init(grouper *gin.RouterGroup) error {
	handler := grouper.Group(ComponentName).Use(login.SessionVerify, middlewares.HiJack)

	handler.POST(PathInsert, Insert)
	handler.POST(PathUpdate, Update)
	handler.POST(PathQuery, Query)
	handler.POST(PathDelete, Delete)
	handler.POST(PathAddRole, AddRole)
	handler.POST(PathDeleteRole, DeleteRole)

	// 初始化管理员账号
	admin := &structure.OperateUser{
		Username:  "admin",
		Nickname:  "SuperAdmin",
		Token:     random.RandString(32),
		UserType:  2,
		Password:  cryptor.Md5String("88888888"), // 密码加密
		Remark:    "super admin account",
		TFASalt:   random.RandString(20),
		TFAStatus: 2,
	}
	if err := mysql.Instance().Where("username = ?", admin.Username).FirstOrCreate(admin).Error; err != nil {
		return err
	}
	if admin.ID > 0 {
		logger.App().Infof("管理员账号已存在,ID: %d, Username: %s", admin.ID, admin.Username)
	}

	return nil
}

// Start 启动
func Start() error { return nil }

// Shutdown 关闭
func Shutdown() error { return nil }

// LoadCache 加载缓存
func LoadCache() error { return nil }
