package role

import (
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/backend/login"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
)

type QueryTotalReq struct{}

type QueryTotalItem struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

type QueryTotalRsp struct {
	List []QueryTotalItem `json:"list"`
}

const PathQueryTotal = "queryTotal"

func QueryTotal(ctx *gin.Context) {
	pid, admin := 0, false
	v, e := ctx.Get(login.CTXUserKey)
	if e {
		user := v.(*structure.OperateUser)
		pid = int(user.ID)
		admin = user.UserType == 2
	}

	children, gsErr := getAllSubUserIDs(pid)
	if gsErr != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query all sub user ids error: [%s]-%s", trace, gsErr.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	tx := mysql.Instance().Model(new(structure.OperateRole)).Select("id,name")
	if !admin && children != nil {
		tx = tx.Where("creator in ?", children)
	}

	list := make([]QueryTotalItem, 0)
	if err := tx.Order("id").Find(&list).Error; err != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, QueryTotalRsp{List: list[:]})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
