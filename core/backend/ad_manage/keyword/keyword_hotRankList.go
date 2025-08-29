package keyword

import (
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gorm.io/gorm"
)

type HotRankListReq struct {
	// TopN        uint16   `json:"top_n"`
	LevelFilter []uint64 `json:"level_filter"`
	Start       int64    `json:"start"`
	End         int64    `json:"end"`
	Page        uint64   `json:"page"`
	PageSize    uint64   `json:"page_size"`
}

type HotRankListItem struct {
	Word         string `json:"word"`
	Count        uint64 `json:"count"`
	Level        uint8  `json:"level"`
	InKeyword    bool   `json:"in_keyword"`
	KeywordLevel uint8  `json:"keyword_level"`
}

type HotRankListRsp struct {
	Page     uint32            `json:"page"`
	PageSize uint32            `json:"page_size"`
	Total    uint64            `json:"total"`
	List     []HotRankListItem `json:"list"`
}

const PathHotRankList = "hotRankList"

func HotRankList(ctx *gin.Context) {
	var req HotRankListReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Start == 0 || req.End == 0 || req.End <= req.Start || req.LevelFilter == nil || len(req.LevelFilter) == 0 || len(req.LevelFilter) > 5 {
		_, response := define.ResponseMsg(define.CodeParamError, "the level filter is illegal", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	page := int(req.Page)
	if req.Page == 0 {
		page = 1
	}
	limit := int(req.PageSize)
	if req.PageSize == 0 {
		limit = 20
	}

	total := int64(0)
	if err := mysql.Instance().Model(new(structure.SearchLog)).Where("created BETWEEN ? AND ?", req.Start, req.End).Distinct("word").Count(&total).Error; err != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 查询分页数据
	slList := make([]*HotRankListItem, 0)
	err := mysql.Instance().Model(new(structure.SearchLog)).
		Select("word, COUNT(*) AS count").
		Where("created BETWEEN ? AND ?", req.Start, req.End).
		Group("word").
		Order("count DESC").
		Offset((page - 1) * limit).Limit(limit).
		Scan(&slList).Error
	if err != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// get hot rank
	countMap := make(map[string]uint64)
	fields := make([]string, 0)
	for _, z := range slList {
		fields = append(fields, cast.ToString(z.Word))
		countMap[z.Word] = z.Count
	}

	// get word in table
	oldMap := make(map[string]structure.Keyword)
	list := make([]structure.Keyword, 0)
	if err := mysql.Instance().Model(new(structure.Keyword)).Where("word in ?", fields).Find(&list).Error; err != nil && err != gorm.ErrRecordNotFound {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	} else {
		for _, item := range list {
			oldMap[item.Word] = item
		}
	}

	// make it

	levelJun := func(count uint64, levelFilter []uint64) uint8 {
		level := uint8(len(levelFilter))
		for i := len(levelFilter) - 1; i >= 0; i-- {
			if count >= levelFilter[i] {
				level = uint8(i + 1)
			} else {
				break
			}
		}
		return level
	}

	hotRankList := make([]HotRankListItem, 0)

	for _, field := range fields {

		old, exist := oldMap[field]

		hotRankList = append(hotRankList, HotRankListItem{
			Word:         field,
			Count:        countMap[field],
			Level:        levelJun(countMap[field], req.LevelFilter),
			InKeyword:    exist,
			KeywordLevel: old.Level,
		})
	}

	_, response := define.Response(define.CodeSuccess, HotRankListRsp{
		Page:     uint32(page),
		PageSize: uint32(limit),
		Total:    uint64(total),
		List:     hotRankList[:],
	})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
