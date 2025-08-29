package keyword

import (
	"errors"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"jarvis/middleware/mq/nats"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type InsertReq struct {
	Word     string `json:"word"`
	ParentID uint64 `json:"parent_id"`
	Level    uint8  `json:"level"`
	Status   uint8  `json:"status"` // 状态(1-启用 2-禁用)
}

type InsertRsp struct{ structure.Keyword }

const PathInsert = "insert"

func Insert(ctx *gin.Context) {
	var req InsertReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Word == "" || req.Level < 1 || req.Level > 5 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	exist := int64(0)
	if tx := mysql.Instance().Model(new(structure.Keyword)).Where("word = ?", req.Word).Count(&exist); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query word [%s] error: [%s]-%s", req.Word, trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if exist > 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the word is exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再插入
	kw := &structure.Keyword{Word: req.Word, ParentID: req.ParentID, Level: req.Level, Status: req.Status}
	if tx := mysql.Instance().Model(kw).Create(kw); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, InsertRsp{Keyword: *kw})
	ctx.AbortWithStatusJSON(http.StatusOK, response)

	go func() {
		go func() {
			if err := nats.Instance().Publish("Search.Cache", []byte("1")); err != nil {
				logger.App().Errorf("publish Search.UpdateAD err: %s", err.Error())
			}
		}()
	}()
}
