package client

import (
	"fmt"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
)

type TotalReq struct {
	Keyword string `json:"keyword"`
}

type TotalItem struct {
	ID   uint   `json:"id"`
	Code string `json:"code"`
	Name string `json:"name"`
}

type TotalRsp struct {
	List []TotalItem `json:"list"`
}

const PathTotal = "total"

func Total(ctx *gin.Context) {
	var req TotalReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	tx := mysql.Instance().Model(new(structure.Client))

	if req.Keyword != "" {
		tx = tx.Where(fmt.Sprintf("name LIKE'%%%s%%'", req.Keyword))
	}

	list := make([]TotalItem, 0)
	if tx = tx.Select("id,code,name").Order("id desc").Find(&list); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, TotalRsp{List: list[:]})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
