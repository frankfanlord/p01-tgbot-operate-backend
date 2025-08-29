package ad

import (
	"errors"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"jarvis/middleware/mq/nats"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type UpdateReq struct {
	ID             int64   `json:"id"`
	Title          string  `json:"title"`           // 广告标题
	Type           uint8   `json:"type"`            // 1-关键词广告 2-置顶广告 3-搜索底部广告 4-搜索初页广告
	Link           string  `json:"link"`            // 广告链接
	ClientID       uint64  `json:"client_id"`       // 客户ID
	PricePerView   float64 `json:"price_per_view"`  // 每次展示价格
	MaxImpressions uint64  `json:"max_impressions"` // 最多展示次数
	StartTime      uint64  `json:"start_time"`      // 开始时间
	StopTime       uint64  `json:"stop_time"`       // 结束时间
	Status         uint8   `json:"status"`          // 状态(1-启用 2-禁用)
}

type UpdateRsp struct{ structure.Ad }

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
	old := new(structure.Ad)
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

	if req.Type > 0 && req.Type != old.Type && old.Type == 1 {
		exist := int64(0)
		if err := mysql.Instance().Model(new(structure.KeywordAd)).Where("ad_id = ?", req.ID).Count(&exist).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("query word [%d] error: [%s]-%s", req.ClientID, trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
		if exist > 0 {
			_, response := define.ResponseMsg(define.CodeParamError, "the ad is bingding keywords", nil)
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	if req.ClientID != 0 {
		exist := int64(0)
		if err := mysql.Instance().Model(new(structure.Client)).Where("id = ?", req.ClientID).Count(&exist).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("query word [%d] error: [%s]-%s", req.ClientID, trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
		if exist == 0 {
			_, response := define.ResponseMsg(define.CodeParamError, "the client is not exist", nil)
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	updates := map[string]any{}

	if req.Title != "" {
		updates["title"] = req.Title
	}
	if req.Link != "" {
		updates["link"] = req.Link
	}
	if req.ClientID != 0 && req.ClientID != old.ClientID {
		updates["client_id"] = req.ClientID
	}
	if req.PricePerView != 0.0 {
		updates["price_per_view"] = req.PricePerView
	}
	if req.MaxImpressions != 0.0 {
		updates["max_impressions"] = req.MaxImpressions
	}
	if req.StartTime != 0 {
		updates["start_time"] = req.StartTime
	}
	if req.StopTime != 0 {
		updates["stop_time"] = req.StopTime
	}
	if req.Type != 0 {
		updates["type"] = req.Type
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
	if tx := mysql.Instance().Model(new(structure.Ad)).Where("id = ?", req.ID).Updates(updates); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 3.查出来
	the := &structure.Ad{}
	if tx := mysql.Instance().Model(the).Where("id = ?", req.ID).First(the); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.ClientID != 0 && req.ClientID != old.ClientID {
		if err := mysql.Instance().Model(new(structure.AdClient)).Where("ad_id = ? and client_id = ?", old.ID, req.ClientID).UpdateColumn("client_id", req.ClientID).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.App().Errorf("update client_id [%s] ad_amount error: %s", req.ClientID, err.Error())
		}
	}

	_, response := define.Response(define.CodeSuccess, UpdateRsp{Ad: *the})
	ctx.AbortWithStatusJSON(http.StatusOK, response)

	go func() {
		go func() {
			if err := nats.Instance().Publish("Search.Cache", []byte("2")); err != nil {
				logger.App().Errorf("publish Search.UpdateAD err: %s", err.Error())
			}
		}()
	}()
}
