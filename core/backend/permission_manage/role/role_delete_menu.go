package role

import (
	"errors"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type DeleteMenuReq struct {
	UserID uint `json:"user_id"`
}

type DeleteMenuRsp struct{}

const PathDeleteMenu = "deleteMenu"

func DeleteMenu(ctx *gin.Context) {
	var req DeleteMenuReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("ShouldBindJSON error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.UserID == 0 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	userExists := int64(0)
	if err := mysql.Instance().Model(new(structure.OperateUser)).Where("id = ?", req.UserID).Count(&userExists).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query user [%d] error: [%s]-%s", trace, req.UserID, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if userExists == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the user is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	relationExist := int64(0)
	if err := mysql.Instance().Model(new(structure.OperateUserRole)).Where("user_id = ?", req.UserID).Count(&relationExist).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query user [%d] relation error: [%s]-%s", trace, req.UserID, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if relationExist == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the user don't have any role", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再修改
	if tx := mysql.Instance().Model(new(structure.OperateUserRole)).Where("user_id = ?", req.UserID).Delete(new(structure.OperateUserRole)); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("delete error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, DeleteMenuRsp{})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
