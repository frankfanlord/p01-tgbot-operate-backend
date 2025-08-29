package cog_manage

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
	Title    string `json:"title"`
	Type     uint8  `json:"type"`   // 0-全部 1-频道 2-群组
	Status   uint8  `json:"status"` // 0-全部 1-新增 2-确认中 3-已确认
	Page     uint64 `json:"page"`
	PageSize uint64 `json:"page_size"`
}

type QueryRsp struct {
	Page     uint32          `json:"page"`
	PageSize uint32          `json:"page_size"`
	Total    uint64          `json:"total"`
	List     []structure.COG `json:"list"`
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

	tx := mysql.Instance().Model(new(structure.COG))

	if req.Title != "" {
		tx = tx.Where(fmt.Sprintf("title LIKE'%%%s%%'", req.Title))
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

	list := make([]structure.COG, 0)
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
