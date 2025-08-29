package menu

import (
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
)

type QueryReq struct{}

type QueryRsp struct {
	List []*MenuItem `json:"list"`
}

const PathQuery = "query"

func Query(ctx *gin.Context) {
	// var req QueryReq
	// if err := ctx.ShouldBindJSON(&req); err != nil {
	// 	trace, response := define.Response(define.CodeParamError, nil)
	// 	logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
	// 	ctx.AbortWithStatusJSON(http.StatusOK, response)
	// 	return
	// }

	list := make([]structure.OperateMenu, 0)
	if err := mysql.Instance().Model(new(structure.OperateMenu)).Order("sort,id").Find(&list).Error; err != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query total error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	root := &MenuItem{Children: make([]*MenuItem, 0)}

	for _, item := range list {
		root.AddChild(&MenuItem{OperateMenu: item, Children: make([]*MenuItem, 0)})
	}

	_, response := define.Response(define.CodeSuccess, QueryRsp{
		List: root.Children[:],
	})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
