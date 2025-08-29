package keyword

import (
	"errors"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"jarvis/middleware/mq/nats"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/duke-git/lancet/slice"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type AddAdReq struct {
	KeywordID uint   `json:"keyword_id"`
	AdIDs     []uint `json:"ad_ids"`
}

type AddAdRsp struct{}

const PathAddAd = "addAd"

func AddAd(ctx *gin.Context) {
	var req AddAdReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("ShouldBindJSON error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.KeywordID == 0 || req.AdIDs == nil || len(req.AdIDs) == 0 {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// make sure keyword is exists
	keywordExists := int64(0)
	if err := mysql.Instance().Model(new(structure.Keyword)).Where("id = ?", req.KeywordID).Count(&keywordExists).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query keyword [%d] error: [%s]-%s", req.KeywordID, trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if keywordExists == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the keyword is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// make sure all menu is exists
	adExists := int64(0)
	if err := mysql.Instance().Model(new(structure.Ad)).Where("id in ? and type = 1", req.AdIDs).Count(&adExists).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query ad [%+v] error: [%s]-%s", req.AdIDs, trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if adExists != int64(len(req.AdIDs)) {
		_, response := define.ResponseMsg(define.CodeParamError, "the ad id list is not totally exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// get all menu that keyword have currentlly
	relationList := make([]uint, 0)
	if err := mysql.Instance().Model(new(structure.KeywordAd)).Select("ad_id").Where("keyword_id = ?", req.KeywordID).Find(&relationList).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query keyword [%d] relation error: [%s]-%s", req.KeywordID, trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	/*
		get all relations
		if relations not in request.AdIDs , then delete
		if request.AdIDs not in relations , then add
	*/

	needDelete := make([]uint, 0)
	needAdd := make([]*structure.KeywordAd, 0)

	for _, adID := range relationList {
		if !slice.Contain(req.AdIDs, adID) {
			needDelete = append(needDelete, adID)
		}
	}
	for _, adID := range req.AdIDs {
		if !slice.Contain(relationList, adID) {
			needAdd = append(needAdd, &structure.KeywordAd{KeywordID: uint64(req.KeywordID), AdID: uint64(adID)})
		}
	}

	// delete
	if len(needDelete) > 0 {
		if err := mysql.Instance().Model(new(structure.KeywordAd)).Where("keyword_id = ? and ad_id in ?", req.KeywordID, needDelete).Delete(new(structure.KeywordAd)).Error; err != nil {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("delete role's menu %+v relation error: [%s]-%s", needDelete, trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	// add
	if len(needAdd) > 0 {
		if err := mysql.Instance().Model(new(structure.KeywordAd)).Create(needAdd).Error; err != nil {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("add role's menu %d relation error: [%s]-%s", len(needAdd), trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	relations := make([]structure.KeywordAd, 0)
	if err := mysql.Instance().Model(new(structure.KeywordAd)).Where("keyword_id = ?", req.KeywordID).Find(&relations).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("find keyword's ad %d relation error: [%s]-%s", req.KeywordID, trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, AddAdRsp{})
	ctx.AbortWithStatusJSON(http.StatusOK, response)

	go func() {
		go func() {
			if err := nats.Instance().Publish("Search.Cache", []byte("3")); err != nil {
				logger.App().Errorf("publish Search.UpdateAD err: %s", err.Error())
			}
		}()
	}()
}
