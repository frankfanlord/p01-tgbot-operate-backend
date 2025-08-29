package log_manage

import (
	"fmt"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
)

type OperateQueryReq struct {
	Username string `json:"username"`
	Behavior string `json:"behavior"`
	Page     uint64 `json:"page"`
	PageSize uint64 `json:"page_size"`
}

type OperateQueryRsp struct {
	Page     uint32                 `json:"page"`
	PageSize uint32                 `json:"page_size"`
	Total    uint64                 `json:"total"`
	List     []structure.OperateLog `json:"list"`
}

const PathOperateQuery = "operateQuery"

func OperateQuery(ctx *gin.Context) {
	var req OperateQueryReq
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

	tx := mysql.Instance().Model(new(structure.OperateLog))

	if req.Username != "" {
		tx = tx.Where(fmt.Sprintf("user LIKE'%%%s%%'", req.Username))
	}
	if req.Behavior != "" {
		tx = tx.Where(fmt.Sprintf("desc LIKE'%%%s%%'", req.Behavior))
	}

	total := int64(0)
	if tmp := tx.Count(&total); tmp.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query total error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	list := make([]structure.OperateLog, 0)
	if tx = tx.Offset((page - 1) * limit).Limit(limit).Order("id desc").Find(&list); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, OperateQueryRsp{
		Page:     uint32(page),
		PageSize: uint32(limit),
		Total:    uint64(total),
		List:     list[:],
	})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
