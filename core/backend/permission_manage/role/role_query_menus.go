package role

import (
	"fmt"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
)

type QueryMenusReq struct {
	RoleID uint `json:"role_id"`
}

type MenuItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type QueryMenusRsp struct {
	List []MenuItem `json:"list"`
}

const PathQueryMenus = "queryMenus"

func QueryMenus(ctx *gin.Context) {
	var req QueryMenusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.RoleID == 0 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	list := make([]MenuItem, 0)
	if err := mysql.Instance().Table(fmt.Sprintf("%s as c", new(structure.OperateRole).TableName())).Select("e.id as id, e.name as name").
		Joins(fmt.Sprintf("left join %s as d on d.role_id = c.id", new(structure.OperateRoleMenu).TableName())).
		Joins(fmt.Sprintf("left join %s as e on e.id = d.menu_id", new(structure.OperateMenu).TableName())).
		Where("c.id = ? AND c.id IS NOT NULL", req.RoleID).Order("e.sort,e.id").Find(&list).Error; err != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, list[:])
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
