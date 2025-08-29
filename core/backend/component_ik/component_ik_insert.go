package component_ik

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
	Word   string `json:"word"`
	Type   uint8  `json:"type"`   // 1-分词 2-停词
	Status uint8  `json:"status"` // 1-启用 2-禁用
}

type InsertRsp struct {
	structure.Participle
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

	if req.Word == "" || req.Type < 1 || req.Type > 2 || req.Status < 1 || req.Status > 2 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	exist := int64(0)
	if tx := mysql.Instance().Model(new(structure.Participle)).Where("word = ?", req.Word).Count(&exist); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if exist > 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the word is exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再插入
	participle := &structure.Participle{Word: req.Word, Type: req.Type, Status: req.Status}
	if tx := mysql.Instance().Model(participle).Create(participle); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, participle)
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
