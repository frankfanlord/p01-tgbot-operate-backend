package user

import (
	"errors"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/duke-git/lancet/v2/random"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateReq struct {
	ID        uint   `json:"id"` // ID
	Nickname  string `json:"nickname"`
	Password  string `json:"password"`
	Remark    string `json:"remark"`
	TFAStatus uint8  `json:"tfa_status"` // 2FA状态 1-未启用 2-已启用
	Status    uint8  `json:"status"`     // 状态(1-启用 2-禁用)
}

type UpdateRsp struct{ structure.OperateUser }

const PathUpdate = "update"

func Update(ctx *gin.Context) {
	var req UpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.ID == 0 || req.TFAStatus > 2 || req.Status > 2 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	old := new(structure.OperateUser)
	if tx := mysql.Instance().Model(new(structure.OperateUser)).Where("id = ?", req.ID).First(old); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if old.ID == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the user of id is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	updates := map[string]any{}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Password != "" {
		updates["password"] = cryptor.Md5String(req.Password)
	}
	if req.Remark != "" {
		updates["remark"] = req.Remark
	}
	if req.TFAStatus != 0 {
		if req.TFAStatus == 2 {
			if old.TFASalt == "" {
				updates["tfa_salt"] = random.RandString(20)
			}
			if old.TFAStatus > 2 {
				req.TFAStatus = old.TFAStatus
			}
		}

		updates["tfa_status"] = req.TFAStatus
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
	if tx := mysql.Instance().Model(new(structure.OperateUser)).Where("id = ?", req.ID).Updates(updates); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 3.查出来
	the := &structure.OperateUser{}
	if tx := mysql.Instance().Model(the).Where("id = ?", req.ID).First(the); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, UpdateRsp{OperateUser: *the})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
