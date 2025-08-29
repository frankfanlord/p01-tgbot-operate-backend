package role

import (
	"errors"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/backend/login"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InsertReq struct {
	Name   string `json:"name"`
	Code   string `json:"code"`
	Remark string `json:"remark"`
	Status uint8  `json:"status"` // 1-启用 2-禁用
}

type InsertRsp struct{ structure.OperateRole }

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

	if req.Name == "" || req.Code == "" || req.Status < 1 || req.Status > 2 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	exist := int64(0)
	if tx := mysql.Instance().Model(new(structure.OperateRole)).Where("name = ?", req.Name).Count(&exist); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if exist > 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the name of role is exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	codeExist := int64(0)
	if tx := mysql.Instance().Model(new(structure.OperateRole)).Where("code = ?", req.Code).Count(&codeExist); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if codeExist > 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the code of role is exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再插入
	newRole := &structure.OperateRole{
		Name:    req.Name,
		Code:    req.Code,
		Remark:  req.Remark,
		Creator: uint(pid),
		Status:  req.Status,
	}
	if tx := mysql.Instance().Model(newRole).Create(newRole); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, InsertRsp{OperateRole: *newRole})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
