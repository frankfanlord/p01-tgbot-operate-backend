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

type InsertReq struct {
	Title            string `json:"title"`     // 菜单标题
	ParentID         uint   `json:"parent_id"` // default 0
	Icon             string `json:"icon"`
	Type             uint8  `json:"type"` // 菜单类型(1-目录 2-按钮)
	Path             string `json:"path"` // 菜单路径
	Name             string `json:"name"`
	Remark           string `json:"remark"`
	Affix            bool   `json:"affix"`
	Cache            bool   `json:"cache"`
	Hidden           bool   `json:"hidden"`
	BreadcrumbEnable bool   `json:"breadcrumbEnable"`
	Component        string `json:"component"`
	Status           uint8  `json:"status"` // 状态(1-启用 2-禁用)
}

type InsertRsp struct{ structure.OperateMenu }

const PathInsert = "insert"

func Insert(ctx *gin.Context) {
	var req InsertReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Title == "" {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	exist := int64(0)
	if err := mysql.Instance().Model(new(structure.OperateMenu)).Where("title = ? and parent_id = ?", req.Title, req.ParentID).Count(&exist).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if exist > 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the title is exist on same level", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2. parent
	parentMenu := &structure.OperateMenu{}
	if req.ParentID > 0 {
		if err := mysql.Instance().Model(new(structure.OperateMenu)).Where("id = ?", req.ParentID).First(parentMenu).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("query parent id [%d] error: [%s]-%s", req.ParentID, trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
		if parentMenu.ID == 0 {
			_, response := define.ResponseMsg(define.CodeParamError, "the id of parent menu is not exist", nil)
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	if req.Name != "" {
		exist = int64(0)
		if err := mysql.Instance().Model(new(structure.OperateMenu)).Where("name = ?", req.Name).Count(&exist).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("query error: [%s]-%s", trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
		if exist > 0 {
			_, response := define.ResponseMsg(define.CodeParamError, "the name is exist", nil)
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	// 2. next id in same level
	curMaxSort := &struct {
		MaxSort int `json:"max_sort"`
	}{}
	if err := mysql.Instance().Model(new(structure.OperateMenu)).Select("MAX(sort) as max_sort").Where("parent_id = ?", req.ParentID).Scan(curMaxSort).Error; err != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query max_sort error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.ParentID == 0 {
		curMaxSort.MaxSort += 1
	} else {
		if curMaxSort.MaxSort == 0 {
			curMaxSort.MaxSort = parentMenu.Sort*10 + 1
		} else {
			curMaxSort.MaxSort += 1
		}
	}

	// 2.再插入
	newMenu := &structure.OperateMenu{
		Title:            req.Title,
		ParentID:         req.ParentID,
		Icon:             req.Icon,
		Type:             req.Type,
		Path:             req.Path,
		Name:             req.Name,
		Remark:           req.Remark,
		Sort:             curMaxSort.MaxSort,
		Affix:            req.Affix,
		Cache:            req.Cache,
		Hidden:           req.Hidden,
		BreadcrumbEnable: req.BreadcrumbEnable,
		Component:        req.Component,
		Status:           req.Status,
	}
	if tx := mysql.Instance().Model(newMenu).Create(newMenu); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, InsertRsp{OperateMenu: *newMenu})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
