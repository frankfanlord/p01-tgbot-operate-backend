package client

import (
	"errors"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeleteReq struct {
	ID int64 `json:"id"`
}
type DeleteRsp struct{}

const PathDelete = "delete"

func Delete(ctx *gin.Context) {
	var req DeleteReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Delete error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	exist := int64(0)
	if tx := mysql.Instance().Model(new(structure.Client)).Where("id = ? and status = 2", req.ID).Count(&exist); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Delete error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if exist == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the record of id is not exist or need turn off", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再修改
	if tx := mysql.Instance().Model(new(structure.Client)).Where("id = ?", req.ID).Delete(new(structure.Client)); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Delete error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if tx := mysql.Instance().Model(new(structure.AdClient)).Where("client_id = ?", req.ID).Delete(new(structure.AdClient)); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Delete error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, DeleteRsp{})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
