package dashboard

import (
	"context"
	"fmt"
	"jarvis/dao/db/redis"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"time"

	"github.com/gin-gonic/gin"
	ORedis "github.com/redis/go-redis/v9"
	"github.com/spf13/cast"
)

type TopicItem struct {
	Topic string  `json:"topic"` // 话题
	Score float64 `json:"score"` // 热度
	Count uint64  `json:"count"` // search times
}

type VisitItem struct {
	Time  int64  `json:"time"`
	Count uint64 `json:"count"`
}

type InfoRsp struct {
	TodayNewUse  uint64       `json:"today_new_use"`  // 今日新增人数使用量
	TodayNewUser uint64       `json:"today_new_user"` // 今日新增人数
	TodayUse     uint64       `json:"today_use"`      // 今日使用量
	TodayUser    uint64       `json:"today_user"`     // 今日使用人数
	TotalUse     uint64       `json:"total_use"`      // 总使用量
	TotalUser    uint64       `json:"total_user"`     // 总用户量
	HotTopic     []TopicItem  `json:"hot_topic"`      // 热门话题
	Visit        []*VisitItem `json:"visit"`          // 最近访问
}

const PathInfo = "info"

func Info(ctx *gin.Context) {
	rsp := &InfoRsp{}

	nowTime := time.Now()
	now := nowTime.Format("20060102")

	// today new user use
	if cmd := redis.Instance().Get(context.Background(), fmt.Sprintf("%s:TodayNewUserUse", now)); cmd.Err() != nil && cmd.Err() != ORedis.Nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Get error: [%s]-%s", trace, cmd.Err().Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	} else {
		rsp.TodayNewUse = cast.ToUint64(cmd.Val())
	}

	// today new user
	if cmd := redis.Instance().PFCount(context.Background(), fmt.Sprintf("%s:TodayNewUser", now)); cmd.Err() != nil && cmd.Err() != ORedis.Nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("PFCount error: [%s]-%s", trace, cmd.Err().Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	} else {
		rsp.TodayNewUser = uint64(cmd.Val())
	}

	// today use
	if cmd := redis.Instance().Get(context.Background(), fmt.Sprintf("%s:TodayUse", now)); cmd.Err() != nil && cmd.Err() != ORedis.Nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Get error: [%s]-%s", trace, cmd.Err().Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	} else {
		rsp.TodayUse = cast.ToUint64(cmd.Val())
	}

	// today user
	if cmd := redis.Instance().PFCount(context.Background(), fmt.Sprintf("%s:TodayUser", now)); cmd.Err() != nil && cmd.Err() != ORedis.Nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("PFCount error: [%s]-%s", trace, cmd.Err().Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	} else {
		rsp.TodayUser = uint64(cmd.Val())
	}

	// total use
	if cmd := redis.Instance().Get(context.Background(), "TotalUse"); cmd.Err() != nil && cmd.Err() != ORedis.Nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Get error: [%s]-%s", trace, cmd.Err().Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	} else {
		rsp.TotalUse = cast.ToUint64(cmd.Val())
	}

	// total user
	if cmd := redis.Instance().PFCount(context.Background(), "TotalUser"); cmd.Err() != nil && cmd.Err() != ORedis.Nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("PFCount error: [%s]-%s", trace, cmd.Err().Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	} else {
		rsp.TotalUser = uint64(cmd.Val())
	}

	// hot rank list
	if cmd := redis.Instance().ZRevRangeWithScores(context.Background(), "HotRankList", 0, 39); cmd.Err() != nil && cmd.Err() != ORedis.Nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("zrevrange error: [%s]-%s", trace, cmd.Err().Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	} else {
		if cmd.Val() != nil && len(cmd.Val()) > 0 {
			scoreMap := make(map[string]float64)
			fields := make([]string, 0)
			for _, z := range cmd.Val() {
				scoreMap[cast.ToString(z.Member)] = cast.ToFloat64(fmt.Sprintf("%.2f", z.Score))
				fields = append(fields, cast.ToString(z.Member))
			}

			if cmd := redis.Instance().HMGet(context.Background(), "SearchHash", fields...); cmd.Err() != nil && cmd.Err() != ORedis.Nil {
				trace, response := define.Response(define.CodeSvrInternalError, nil)
				logger.App().Errorf("HMGet error: [%s]-%s", trace, cmd.Err().Error())
				ctx.AbortWithStatusJSON(http.StatusOK, response)
				return
			} else {
				if cmd.Val() != nil && len(cmd.Val()) > 0 {
					list := make([]TopicItem, 0)
					for i, v := range cmd.Val() {
						key := fields[i]
						list = append(list, TopicItem{
							Topic: key,
							Score: scoreMap[key],
							Count: cast.ToUint64(v),
						})
					}
					rsp.HotTopic = list[:]
				}
			}

		}
	}

	// visit
	list := make([]*VisitItem, 0)
	keys := make([]string, 0)
	for i := 0; i < 12; i++ {
		current := nowTime.Add(time.Hour * time.Duration(-i))
		keys = append(keys, fmt.Sprintf("%s:Use", current.Format("2006010215")))
		list = append(list, &VisitItem{Time: current.Unix()})
	}

	if cmd := redis.Instance().MGet(context.Background(), keys...); cmd.Err() != nil && cmd.Err() != ORedis.Nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Get error: [%s]-%s", trace, cmd.Err().Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	} else {
		for i := range cmd.Val() {
			item := list[i]
			if cmd.Val()[i] != nil {
				item.Count = cast.ToUint64(cmd.Val()[i])
			}
		}
	}

	for i := len(list) - 1; i >= 0; i-- {
		rsp.Visit = append(rsp.Visit, list[i])
	}

	_, response := define.Response(define.CodeSuccess, rsp)
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
