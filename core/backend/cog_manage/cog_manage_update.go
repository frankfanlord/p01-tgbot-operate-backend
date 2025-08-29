package cog_manage

import (
	"errors"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateReq struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Remark   string `json:"remark"`
	Category uint8  `json:"category"`
}

type UpdateRsp struct{ structure.COG }

const PathUpdate = "update"

func Update(ctx *gin.Context) {
	var req UpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.ID == 0 || req.Category > 27 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	old := new(structure.COG)
	if tx := mysql.Instance().Model(old).Where("id = ?", req.ID).First(old); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if old.ID == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the record of id is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	updates := map[string]any{}

	if req.Username != "" && old.Status == 1 {
		if strings.HasPrefix(req.Username, "@") {
			req.Username = strings.TrimPrefix(req.Username, "@")
		}
		updates["username"] = req.Username
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.Category != 0 {
		updates["category"] = req.Category
	}

	if len(updates) == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "nothing to update", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再修改
	if tx := mysql.Instance().Model(new(structure.COG)).Where("id = ?", req.ID).Updates(updates); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 3.查出来
	the := &structure.COG{}
	if tx := mysql.Instance().Model(the).Where("id = ?", req.ID).First(the); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, UpdateRsp{COG: *the})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
