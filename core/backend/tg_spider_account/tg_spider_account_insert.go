package tg_spider_account

import (
	"errors"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"
)

type InsertReq struct {
	Phone   string `json:"phone"`
	AppID   uint64 `json:"app_id"`
	AppHash string `json:"app_hash"`
	TFAPwd  string `json:"tfa_pwd"`
}

type InsertRsp struct {
	structure.TGSpiderAccount
}

const PathInsert = "insert"

func Insert(ctx *gin.Context) {
	var req InsertReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Phone == "" {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	exist := int64(0)
	if tx := mysql.Instance().Model(new(structure.TGSpiderAccount)).
		Where("phone = ? AND app_id = ? AND app_hash = ?", req.Phone, req.AppID, req.AppHash).
		Count(&exist); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if exist > 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the combination is exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再插入
	sAccount := &structure.TGSpiderAccount{Phone: req.Phone, AppID: req.AppID, AppHash: req.AppHash, TFAPwd: req.TFAPwd}
	if tx := mysql.Instance().Model(sAccount).Create(sAccount); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, sAccount)
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
