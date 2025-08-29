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

type QueryReq struct {
	Title    string `json:"title"`
	Status   uint8  `json:"status"` // 0-全部 1-启用 2-禁用
	Page     uint64 `json:"page"`
	PageSize uint64 `json:"page_size"`
}

type QueryItem struct {
	structure.Ad
	ClientName string `json:"client_name"`
}

type QueryRsp struct {
	Page     uint32      `json:"page"`
	PageSize uint32      `json:"page_size"`
	Total    uint64      `json:"total"`
	List     []QueryItem `json:"list"`
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

	tx := mysql.Instance().Model(new(structure.Ad))

	if req.Title != "" {
		tx = tx.Where(fmt.Sprintf("title LIKE'%%%s%%'", req.Title))
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

	atx := mysql.Instance().Table(fmt.Sprintf("%s as a", new(structure.Ad).TableName())).Select("a.*,c.name as client_name").
		Joins(fmt.Sprintf("left join %s as b on b.ad_id = a.id", new(structure.AdClient).TableName())).
		Joins(fmt.Sprintf("left join %s as c on c.id = b.client_id", new(structure.Client).TableName()))

	if req.Title != "" {
		atx = atx.Where(fmt.Sprintf("a.title LIKE'%%%s%%'", req.Title))
	}
	if req.Status != 0 {
		atx = atx.Where("a.status = ?", req.Status)
	}

	list := make([]QueryItem, 0)
	if atx = atx.Offset((page - 1) * limit).Limit(limit).Order("a.id desc").Find(&list); atx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, atx.Error.Error())
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
