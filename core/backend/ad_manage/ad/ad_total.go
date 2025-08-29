package ad

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
	Title string `json:"title"`
}

type TotalItem struct {
	ID         uint   `json:"id"`
	Title      string `json:"title"`
	ClientName string `json:"client_name"`
}

type TotalRsp struct {
	List []TotalItem `json:"list"`
}

const PathTotal = "total"

func Total(ctx *gin.Context) {
	var req TotalReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Total error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	atx := mysql.Instance().Table(fmt.Sprintf("%s as a", new(structure.Ad).TableName())).Select("a.id as id,a.title as title,c.name as client_name").
		Joins(fmt.Sprintf("left join %s as b on b.ad_id = a.id", new(structure.AdClient).TableName())).
		Joins(fmt.Sprintf("left join %s as c on c.id = b.client_id", new(structure.Client).TableName())).
		Where("a.status = 1 and a.type = 1 and a.id is not null")

	if req.Title != "" {
		atx = atx.Where(fmt.Sprintf("a.title LIKE'%%%s%%'", req.Title))
	}

	list := make([]TotalItem, 0)
	if atx = atx.Order("a.id desc").Find(&list); atx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Total error: [%s]-%s", trace, atx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, TotalRsp{List: list[:]})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
