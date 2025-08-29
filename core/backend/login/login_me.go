package login

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

type MeReq struct{}

type MeRsp struct {
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Remark   string `json:"remark"`
	RoleName string `json:"role_name"`
	RoleCode string `json:"role_code"`
}

const PathMe = "me"

func Me(ctx *gin.Context) {
	// var req MeReq
	// if err := ctx.ShouldBindJSON(&req); err != nil {
	// 	trace, response := define.Response(define.CodeParamError, nil)
	// 	logger.App().Errorf("Update error: [%s]-%s", trace, err.Error())
	// 	ctx.AbortWithStatusJSON(http.StatusOK, response)
	// 	return
	// }

	token := ctx.GetHeader(CTXHeaderTokenKey)

	// 1.先查询有没有
	old := new(structure.OperateUser)
	if tx := mysql.Instance().Model(new(structure.OperateUser)).Where("token = ?", token).First(old); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if old.ID == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the token of user is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	rsp := &MeRsp{
		Nickname: old.Nickname,
		Username: old.Username,
		Remark:   old.Remark,
	}
	if err := mysql.Instance().Table(fmt.Sprintf("%s as a", new(structure.OperateUser).TableName())).Select("c.code as role_code,c.name as role_name").
		Joins(fmt.Sprintf("left join %s as b on b.user_id = a.id", new(structure.OperateUserRole).TableName())).
		Joins(fmt.Sprintf("left join %s as c on c.id = b.role_id", new(structure.OperateRole).TableName())).
		Where("a.token = ?", token).First(rsp).Error; err != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, rsp)
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
