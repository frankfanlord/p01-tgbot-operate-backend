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

type LoginQueryReq struct {
	Username string `json:"username"`
	Page     uint64 `json:"page"`
	PageSize uint64 `json:"page_size"`
}

type LoginQueryRsp struct {
	Page     uint32               `json:"page"`
	PageSize uint32               `json:"page_size"`
	Total    uint64               `json:"total"`
	List     []structure.LoginLog `json:"list"`
}

const PathLoginQuery = "loginQuery"

func LoginQuery(ctx *gin.Context) {
	var req LoginQueryReq
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

	tx := mysql.Instance().Model(new(structure.LoginLog))

	if req.Username != "" {
		tx = tx.Where(fmt.Sprintf("user LIKE'%%%s%%'", req.Username))
	}

	total := int64(0)
	if tmp := tx.Count(&total); tmp.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query total error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	list := make([]structure.LoginLog, 0)
	if tx = tx.Offset((page - 1) * limit).Limit(limit).Order("id desc").Find(&list); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, LoginQueryRsp{
		Page:     uint32(page),
		PageSize: uint32(limit),
		Total:    uint64(total),
		List:     list[:],
	})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
