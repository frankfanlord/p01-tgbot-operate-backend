package keyword

import (
	"errors"
	"fmt"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type QueryAdReq struct {
	KeywordID uint `json:"keyword_id"`
}

type QueryAdItem struct {
	ID     uint   `json:"id"`
	Title  string `json:"title"`
	Status uint8  `json:"status"` // 状态(1-启用 2-禁用)
}

type QueryAdRsp struct {
	List []QueryAdItem `json:"list"`
}

const PathQueryAd = "queryAd"

func QueryAd(ctx *gin.Context) {
	var req QueryAdReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("ShouldBindJSON error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.KeywordID == 0 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// make sure keyword is exists
	keywordExists := int64(0)
	if err := mysql.Instance().Model(new(structure.Keyword)).Where("id = ?", req.KeywordID).Count(&keywordExists).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query keyword [%d] error: [%s]-%s", req.KeywordID, trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if keywordExists == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the keyword is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	atx := mysql.Instance().
		Table(fmt.Sprintf("%s as a", new(structure.KeywordAd).TableName())).
		Select("b.id as id,b.title as title,b.status as status").
		Joins(fmt.Sprintf("left join %s as b on b.id = a.ad_id", new(structure.Ad).TableName())).
		Where("a.keyword_id = ?", req.KeywordID)

	list := make([]QueryAdItem, 0)
	if atx = atx.Order("a.id desc").Find(&list); atx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, atx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, QueryAdRsp{List: list[:]})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
