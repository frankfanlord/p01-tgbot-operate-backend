package user

import (
	"errors"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/backend/login"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/duke-git/lancet/v2/random"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InsertReq struct {
	Username  string `json:"username"`
	Nickname  string `json:"nickname"`
	UserType  uint8  `json:"user_type"` // 1-普通用户 2-管理员
	Password  string `json:"password"`  // 密码
	Remark    string `json:"remark"`
	TFAStatus uint8  `json:"tfa_status"` // 2FA状态(1-未启用 2-已启用)
	Status    uint8  `json:"status"`     // 1-启用 2-禁用
}

type InsertRsp struct{ structure.OperateUser }

const PathInsert = "insert"

func Insert(ctx *gin.Context) {
	pid := 0
	v, e := ctx.Get(login.CTXUserKey)
	if e {
		user := v.(*structure.OperateUser)
		pid = int(user.ID)
	}

	var req InsertReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Username == "" || req.Nickname == "" || req.Password == "" || req.UserType < 1 || req.UserType > 2 || req.Status < 1 || req.Status > 2 || req.TFAStatus < 1 || req.TFAStatus > 2 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	exist := int64(0)
	if tx := mysql.Instance().Model(new(structure.OperateUser)).Where("username = ?", req.Username).Count(&exist); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if exist > 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the username is exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再插入
	tfaSalt := ""
	if req.TFAStatus == 2 {
		tfaSalt = random.RandString(20)
		req.TFAStatus = 3
	}
	newUser := &structure.OperateUser{
		Username:  req.Username,
		Nickname:  req.Nickname,
		Token:     random.RandString(32),
		UserType:  req.UserType,
		Password:  cryptor.Md5String(req.Password), // 密码加密
		Remark:    req.Remark,
		TFASalt:   tfaSalt,
		TFAStatus: req.TFAStatus,
		ParentID:  uint(pid),
		Status:    req.Status,
	}
	if tx := mysql.Instance().Model(newUser).Create(newUser); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, InsertRsp{OperateUser: *newUser})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
