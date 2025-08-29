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

type InsertReq struct {
	Title          string  `json:"title"`           // 广告标题
	Link           string  `json:"link"`            // 广告链接
	Type           uint8   `json:"type"`            // 1-关键词广告 2-置顶广告 3-搜索內连大广告 4-搜索內连小广告 5-搜索内容广告
	ClientID       uint64  `json:"client_id"`       // 客户ID
	PricePerView   float64 `json:"price_per_view"`  // 每次展示价格
	MaxImpressions uint64  `json:"max_impressions"` // 最多展示次数
	StartTime      uint64  `json:"start_time"`      // 开始时间
	StopTime       uint64  `json:"stop_time"`       // 结束时间
	Status         uint8   `json:"status"`          // 状态(1-启用 2-禁用)
}

type InsertRsp struct{ structure.Ad }

const PathInsert = "insert"

func Insert(ctx *gin.Context) {
	var req InsertReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Title == "" || req.Link == "" || req.Status < 1 || req.Status > 2 || req.Type < 1 || req.Type > 5 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.ClientID != 0 {
		exist := int64(0)
		if err := mysql.Instance().Model(new(structure.Client)).Where("id = ?", req.ClientID).Count(&exist).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("query client_id [%s] error: [%s]-%s", req.ClientID, trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
		if exist == 0 {
			_, response := define.ResponseMsg(define.CodeParamError, "the client id is not exist", nil)
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	// 2.再插入
	the := &structure.Ad{
		Title:          req.Title,
		Link:           req.Link,
		ClientID:       req.ClientID,
		PricePerView:   req.PricePerView,
		MaxImpressions: req.MaxImpressions,
		StartTime:      req.StartTime,
		StopTime:       req.StopTime,
		Type:           req.Type,
		Status:         req.Status,
	}
	if tx := mysql.Instance().Model(the).Create(the); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Insert error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// update client
	if req.ClientID != 0 {
		relation := &structure.AdClient{AdID: uint64(the.ID), ClientID: uint64(req.ClientID)}
		if err := mysql.Instance().Model(relation).Create(relation).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.App().Errorf("create ad_client [%d-%d] error: %s", the.ID, req.ClientID, err.Error())
		}
		if err := mysql.Instance().Model(new(structure.Client)).Where("id = ?", req.ClientID).UpdateColumn("ad_count", gorm.Expr("ad_count + ?", 1)).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			logger.App().Errorf("update client_id [%s] ad_amount error: %s", req.ClientID, err.Error())
		}
	}

	_, response := define.Response(define.CodeSuccess, InsertRsp{Ad: *the})
	ctx.AbortWithStatusJSON(http.StatusOK, response)

	go func() {
		if err := nats.Instance().Publish("Search.Cache", []byte("2")); err != nil {
			logger.App().Errorf("publish Search.UpdateAD err: %s", err.Error())
		}
	}()
}
