package role

import (
	"errors"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/duke-git/lancet/v2/slice"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateMenusReq struct {
	RoleID  uint   `json:"role_id"`
	MenuIDs []uint `json:"menu_ids"`
}

type UpdateMenusRsp struct {
	List []structure.OperateRoleMenu
}

const PathUpdateMenus = "updateMenus"

func UpdateMenus(ctx *gin.Context) {
	var req UpdateMenusReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("ShouldBindJSON error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.RoleID == 0 || req.MenuIDs == nil || len(req.MenuIDs) == 0 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// make sure role is exists
	roleExists := int64(0)
	if err := mysql.Instance().Model(new(structure.OperateRole)).Where("id = ? and status = 1", req.RoleID).Count(&roleExists).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query role [%d] error: [%s]-%s", req.RoleID, trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if roleExists == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the role is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// make sure all menu is exists
	menuExists := int64(0)
	if err := mysql.Instance().Model(new(structure.OperateMenu)).Where("id in ?", req.MenuIDs).Count(&menuExists).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query user [%+v] error: [%s]-%s", req.MenuIDs, trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if menuExists != int64(len(req.MenuIDs)) {
		_, response := define.ResponseMsg(define.CodeParamError, "the menu id list is not totally exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// get all menu that role have currentlly
	relationList := make([]uint, 0)
	if err := mysql.Instance().Model(new(structure.OperateRoleMenu)).Select("menu_id").Where("role_id = ?", req.RoleID).Find(&relationList).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query role [%d] relation error: [%s]-%s", req.RoleID, trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	/*

		get all relations
		if relations not in request.MenuIDs , then delete
		if request.MenuIDs not in relations , then add

	*/

	needDelete := make([]uint, 0)
	needAdd := make([]*structure.OperateRoleMenu, 0)

	for _, menu_id := range relationList {
		if !slice.Contain(req.MenuIDs, menu_id) {
			needDelete = append(needDelete, menu_id)
		}
	}
	for _, menu_id := range req.MenuIDs {
		if !slice.Contain(relationList, menu_id) {
			needAdd = append(needAdd, &structure.OperateRoleMenu{RoleID: req.RoleID, MenuID: menu_id})
		}
	}

	// delete
	if len(needDelete) > 0 {
		if err := mysql.Instance().Model(new(structure.OperateRoleMenu)).Where("role_id = ? and menu_id in ?", req.RoleID, needDelete).Delete(new(structure.OperateRoleMenu)).Error; err != nil {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("delete role's menu %+v relation error: [%s]-%s", needDelete, trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	// add
	if len(needAdd) > 0 {
		if err := mysql.Instance().Model(new(structure.OperateRoleMenu)).Create(needAdd).Error; err != nil {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("add role's menu %d relation error: [%s]-%s", len(needAdd), trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	relations := make([]structure.OperateRoleMenu, 0)
	if err := mysql.Instance().Model(new(structure.OperateRoleMenu)).Where("role_id = ?", req.RoleID).Find(&relations).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("find role's menu %d relation error: [%s]-%s", req.RoleID, trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, UpdateMenusRsp{List: relations})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
