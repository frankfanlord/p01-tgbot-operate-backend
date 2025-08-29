package login

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

type UpdateMeReq struct {
	Type        uint8  `json:"type"` // 0 or 1
	Nickname    string `json:"nickname"`
	Remark      string `json:"remark"`
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

type UpdateMeRsp struct{ structure.OperateUser }

const PathUpdateMe = "updateMe"

func UpdateMe(ctx *gin.Context) {
	var req UpdateMeReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Type > 1 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Type == 0 && req.Nickname == "" && req.Remark == "" {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Type == 1 && (req.OldPassword == "" || req.NewPassword == "") {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

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

	if req.Type == 1 {
		if req.OldPassword != "" && req.OldPassword != old.Password {
			_, response := define.ResponseMsg(define.CodeParamError, "old password wrong", nil)
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	updates := map[string]any{}
	if req.Type == 0 {
		if req.Nickname != "" {
			updates["nickname"] = req.Nickname
		}
		if req.Remark != "" {
			updates["remark"] = req.Remark
		}
	}
	if req.Type == 1 {
		if req.NewPassword != "" {
			updates["password"] = req.NewPassword
		}
	}

	if len(updates) == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "nothing to update", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再修改
	if tx := mysql.Instance().Model(new(structure.OperateUser)).Where("token = ?", token).Updates(updates); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 3.查出来
	the := &structure.OperateUser{}
	if tx := mysql.Instance().Model(the).Where("token = ?", token).First(the); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, UpdateMeRsp{OperateUser: *the})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
