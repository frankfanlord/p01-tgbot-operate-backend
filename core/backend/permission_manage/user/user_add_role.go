package user

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

type AddRoleReq struct {
	UserID uint `json:"user_id"`
	RoleID uint `json:"role_id"`
}

type AddRoleRsp struct{ structure.OperateUserRole }

const PathAddRole = "addRole"

func AddRole(ctx *gin.Context) {
	var req AddRoleReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("ShouldBindJSON error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.RoleID == 0 || req.UserID == 0 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	roleExists := int64(0)
	if err := mysql.Instance().Model(new(structure.OperateRole)).Where("id = ? and status = 1", req.RoleID).Count(&roleExists).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query role [%d] error: [%s]-%s", trace, req.RoleID, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if roleExists == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the role is not exist", nil)
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
	if relationExist > 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the user already had a role", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	relation := &structure.OperateUserRole{
		UserID: req.UserID,
		RoleID: req.RoleID,
	}
	if err := mysql.Instance().Model(relation).Create(relation).Error; err != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, AddRoleRsp{OperateUserRole: *relation})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
