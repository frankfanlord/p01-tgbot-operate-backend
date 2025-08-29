package tg_spider_account

import (
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"jarvis/middleware/mq/nats"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"
	"strings"
)

type UpdateReq struct {
	ID      uint   `json:"id"`
	Phone   string `json:"phone"`    // 停用才能改
	Code    string `json:"code"`     // 启用才能改
	TFAPwd  string `json:"tfa_pwd"`  // 停用才能改
	AppID   uint64 `json:"app_id"`   // 停用才能改
	AppHash string `json:"app_hash"` // 停用才能改
	Status  uint8  `json:"status"`   // 任何时候都能改 1-停用 2-启用
}

type UpdateRsp struct{ structure.TGSpiderAccount }

const PathUpdate = "update"

func Update(ctx *gin.Context) {
	var req UpdateReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.ID == 0 || req.Status < 0 || req.Status > 2 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.先查询有没有
	old := new(structure.TGSpiderAccount)
	if tx := mysql.Instance().Model(old).Where("id = ?", req.ID).First(old); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			_, response := define.ResponseMsg(define.CodeParamError, "the record of id is not exist", nil)
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	updates := make(map[string]any)

	if req.Phone != "" && req.Phone != old.Phone && old.Status == 1 {
		updates["phone"] = req.Phone
	}
	if req.Code != "" && req.Code != old.Code && old.Status == 2 {
		updates["code"] = req.Code
	}
	if req.TFAPwd != "" && req.TFAPwd != old.TFAPwd && old.Status == 1 {
		updates["tfa_pwd"] = req.TFAPwd
	}
	if req.AppID != 0 && req.AppID != old.AppID && old.Status == 1 {
		updates["app_id"] = req.AppID
	}
	if req.AppHash != "" && req.AppHash != old.AppHash && old.Status == 1 {
		updates["app_hash"] = req.AppHash
	}
	if req.Status != 0 && req.Status != old.Status {
		updates["status"] = req.Status
	}

	if len(updates) == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "nothing to update", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 2.再修改
	if tx := mysql.Instance().Model(new(structure.TGSpiderAccount)).Where("id = ?", req.ID).Updates(updates); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s-%+v", trace, tx.Error.Error(), updates)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 3.查出来
	sAccount := &structure.TGSpiderAccount{}
	if tx := mysql.Instance().Model(sAccount).Where("id = ?", req.ID).First(sAccount); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, sAccount)
	ctx.AbortWithStatusJSON(http.StatusOK, response)

	// 符合 停用->启用 需要发送启动命令
	if old.Status == 1 && req.Status == 2 {
		go func() {
			_startChannel <- StartInfo{
				ID:      sAccount.ID,
				Phone:   sAccount.Phone,
				Code:    sAccount.Code,
				TFAPwd:  sAccount.TFAPwd,
				AppID:   sAccount.AppID,
				AppHash: sAccount.AppHash,
				Session: sAccount.Session,
			}
		}()
	}

	// 符合 启用->停用 需要发送关闭命令
	if old.Status == 2 && req.Status == 1 {
		go func() { _stopChannel <- sAccount.ID }()
	}

	// 下发验证码
	if req.Code != "" && req.Code != old.Code && old.Status == 2 {
		go func() {
			if err := nats.Instance().Publish(
				TGSpiderSSSubjectPrefix+strings.Join([]string{fmt.Sprintf("%d", sAccount.ID), SSActCode}, "."),
				[]byte(sAccount.Code),
			); err != nil {
				logger.App().Errorf("publish code to err: %s-%s", TGSpiderSSSubjectPrefix+SSActCode, err.Error())
			}
		}()
	}
}
