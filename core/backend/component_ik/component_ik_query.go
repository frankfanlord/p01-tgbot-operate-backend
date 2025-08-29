package component_ik

import (
	"fmt"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
)

type QueryReq struct {
	Word     string `json:"word"`      // 模糊搜索
	Type     uint8  `json:"type"`      // 0-全部 1-分词 2-停词
	Status   uint8  `json:"status"`    // 0-全部 1-启用 2-停用
	Page     uint32 `json:"page"`      // 页码(from 1)
	PageSize uint32 `json:"page_size"` // 每页
}

type QueryRsp struct {
	Page     uint32                 `json:"page"`
	PageSize uint32                 `json:"page_size"`
	Total    uint64                 `json:"total"`
	List     []structure.Participle `json:"list"`
}

const PathQuery = "query"

func Query(ctx *gin.Context) {
	var req QueryReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	page := int(req.Page)
	if req.Page == 0 {
		page = 1
	}
	limit := int(req.PageSize)
	if req.PageSize == 0 {
		limit = 20
	}

	tx := mysql.Instance().Model(new(structure.Participle))

	if req.Word != "" {
		tx = tx.Where(fmt.Sprintf("word LIKE'%%%s%%'", req.Word))
	}
	if req.Type != 0 {
		tx = tx.Where("type = ?", req.Type)
	}
	if req.Status != 0 {
		tx = tx.Where("status = ?", req.Status)
	}

	total := int64(0)
	if tmp := tx.Count(&total); tmp.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query total error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	list := make([]structure.Participle, 0)
	if tx = tx.Offset((page - 1) * limit).Limit(limit).Order("id desc").Find(&list); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, QueryRsp{
		Page:     uint32(page),
		PageSize: uint32(limit),
		Total:    uint64(total),
		List:     list[:],
	})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
