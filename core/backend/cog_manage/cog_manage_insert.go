package cog_manage

import (
	"errors"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InsertReq struct {
	Username string `json:"username"`
	Remark   string `json:"remark"`
	Category uint8  `json:"category"`
}

type InsertRsp struct{ structure.COG }

const PathInsert = "insert"

func Insert(ctx *gin.Context) {
	var req InsertReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Username == "" || req.Category < 1 || req.Category > 27 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if strings.HasPrefix(req.Username, "@") {
		req.Username = strings.TrimPrefix(req.Username, "@")
	}

	// 1.先查询有没有
	exist := int64(0)
	if tx := mysql.Instance().Model(new(structure.COG)).Where("username = ?", req.Username).Count(&exist); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query username [%s] error: [%s]-%s", req.Username, trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if exist > 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the username is exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再插入
	cog := &structure.COG{Username: req.Username, Remark: req.Remark, Category: req.Category}
	if tx := mysql.Instance().Model(cog).Create(cog); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, InsertRsp{COG: *cog})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
