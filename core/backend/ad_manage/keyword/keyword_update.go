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

type UpdateReq struct {
	ID       int64  `json:"id"`
	Word     string `json:"word"`
	ParentID uint64 `json:"parent_id"`
	Level    uint8  `json:"level"`
	Status   uint8  `json:"status"` // 状态(1-启用 2-禁用)
}

type UpdateRsp struct{ structure.Keyword }

const PathUpdate = "update"

func Update(ctx *gin.Context) {
	var req UpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.ID == 0 || req.Level > 5 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	old := new(structure.Keyword)
	if tx := mysql.Instance().Model(old).Where("id = ?", req.ID).First(old); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if old.ID == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the record of id is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// if want to change word
	if req.Word != "" {
		exist := int64(0)
		if err := mysql.Instance().Model(new(structure.Keyword)).Where("word = ? and id != ?", req.Word, req.ID).Count(&exist).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("query word [%s] error: [%s]-%s", req.Word, trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
		if exist > 0 {
			_, response := define.ResponseMsg(define.CodeParamError, "the word is exist", nil)
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	updates := map[string]any{}

	if req.Word != "" {
		updates["word"] = req.Word
	}
	if req.ParentID != 0 {
		updates["parent_id"] = req.ParentID
	}
	if req.Level != 0 {
		updates["level"] = req.Level
	}
	if req.Status != 0 {
		updates["status"] = req.Status
	}

	if len(updates) == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "nothing to update", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再修改
	if tx := mysql.Instance().Model(new(structure.Keyword)).Where("id = ?", req.ID).Updates(updates); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 3.查出来
	the := &structure.Keyword{}
	if tx := mysql.Instance().Model(the).Where("id = ?", req.ID).First(the); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, UpdateRsp{Keyword: *the})
	ctx.AbortWithStatusJSON(http.StatusOK, response)

	go func() {
		go func() {
			if err := nats.Instance().Publish("Search.Cache", []byte("1")); err != nil {
				logger.App().Errorf("publish Search.UpdateAD err: %s", err.Error())
			}
		}()
	}()
}
