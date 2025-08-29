package menu

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

type UpdateReq struct {
	ID               uint   `json:"id"` // ID
	Title            string `json:"title"`
	Icon             string `json:"icon"`
	Type             uint8  `json:"type"` // 菜单类型(1-目录 2-按钮)
	Path             string `json:"path"`
	Name             string `json:"name"`
	Remark           string `json:"remark"`
	Sort             uint   `json:"sort"`
	Affix            bool   `json:"affix"`
	Cache            bool   `json:"cache"`
	Hidden           bool   `json:"hidden"`
	BreadcrumbEnable bool   `json:"breadcrumbEnable"`
	Component        string `json:"component"`
	Status           uint8  `json:"status"` // 状态(1-启用 2-禁用)
}

type UpdateRsp struct{ structure.OperateMenu }

const PathUpdate = "update"

func Update(ctx *gin.Context) {
	var req UpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.ID == 0 || req.Type > 2 || req.Status > 2 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	the := new(structure.OperateMenu)
	if err := mysql.Instance().Model(new(structure.OperateMenu)).Where("id = ?", req.ID).First(the).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query [%d] error: [%s]-%s", trace, req.ID, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if the.ID == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the menu of id is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Title != "" {
		exist := int64(0)
		if err := mysql.Instance().Model(new(structure.OperateMenu)).Where("id != ? and title = ? and parent_id = ?", req.ID, req.Title, the.ParentID).Count(&exist).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("query [%d-%s] error: [%s]-%s", trace, req.Title, the.ParentID, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
		if exist == 1 {
			_, response := define.ResponseMsg(define.CodeParamError, "the title of same level is exist", nil)
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}
	if req.Name != "" {
		exist := int64(0)
		if err := mysql.Instance().Model(new(structure.OperateMenu)).Where("id != ? and name = ?", req.ID, req.Name).Count(&exist).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("query [%d-%s] error: [%s]-%s", trace, req.Title, the.ParentID, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
		if exist > 0 {
			_, response := define.ResponseMsg(define.CodeParamError, "the same is exist", nil)
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	updates := map[string]any{}
	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Icon != "" {
		updates["icon"] = req.Icon
	}
	if req.Path != "" {
		updates["path"] = req.Path
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.Type != 0 {
		updates["type"] = req.Type
	}
	if req.Sort != 0 {
		updates["sort"] = req.Sort
	}
	if req.Status != 0 {
		updates["status"] = req.Status
	}
	updates["affix"] = req.Affix
	updates["cache"] = req.Cache
	updates["hidden"] = req.Hidden
	updates["breadcrumb_enable"] = req.BreadcrumbEnable
	if req.Component != "" {
		updates["component"] = req.Component
	}

	if len(updates) == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "nothing to update", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再修改
	if tx := mysql.Instance().Model(new(structure.OperateMenu)).Where("id = ?", req.ID).Updates(updates); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 3.查出来
	the = &structure.OperateMenu{}
	if tx := mysql.Instance().Model(the).Where("id = ?", req.ID).First(the); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, UpdateRsp{OperateMenu: *the})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
