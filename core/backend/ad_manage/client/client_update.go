package client

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
	ID        int64   `json:"id"`
	Code      string  `json:"code"`
	Name      string  `json:"name"`
	TGAccount string  `json:"tg_account"`
	Balance   float64 `json:"balance"`
	Status    uint8   `json:"status"` // 状态(1-启用 2-禁用)
}

type UpdateRsp struct{ structure.Client }

const PathUpdate = "update"

func Update(ctx *gin.Context) {
	var req UpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.ID == 0 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	old := new(structure.Client)
	if err := mysql.Instance().Model(old).Where("id = ?", req.ID).First(old).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if old.ID == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the record of id is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// if want to change code
	if req.Code != "" {
		exist := int64(0)
		if err := mysql.Instance().Model(new(structure.Client)).Where("code = ? and id != ?", req.Code, req.ID).Count(&exist).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("query code [%s] error: [%s]-%s", req.Code, trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
		if exist > 0 {
			_, response := define.ResponseMsg(define.CodeParamError, "the code is exist", nil)
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	updates := map[string]any{}

	if req.Code != "" {
		updates["code"] = req.Code
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.TGAccount != "" {
		updates["tg_account"] = req.TGAccount
	}
	if req.Balance != 0.0 {
		updates["balance"] = gorm.Expr("balance + ?", req.Balance)
	}
	if req.Status != 0 {
		updates["status"] = req.Status
	}

	if len(updates) == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "nothing to update", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再修改
	if tx := mysql.Instance().Model(new(structure.Client)).Where("id = ?", req.ID).Updates(updates); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 3.查出来
	the := &structure.Client{}
	if tx := mysql.Instance().Model(the).Where("id = ?", req.ID).First(the); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, UpdateRsp{Client: *the})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
