package middlewares

import (
	"bytes"
	"errors"
	"io"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"operate-backend/core/structure"
	"time"

	"github.com/bytedance/sonic"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	rw.body.Write(b)
	return rw.ResponseWriter.Write(b)
}

func HiJack(ctx *gin.Context) {
	startTime := time.Now()

	var reqBody []byte
	if ctx.Request.Body != nil {
		reqBody, _ = io.ReadAll(ctx.Request.Body)
		ctx.Request.Body = io.NopCloser(bytes.NewBuffer(reqBody))
	}

	respWriter := &responseWriter{
		ResponseWriter: ctx.Writer,
		body:           bytes.NewBufferString(""),
	}
	ctx.Writer = respWriter

	go saveRequest(startTime.Unix(), ctx.Request.Method, ctx.Request.URL.Path, ctx.ClientIP(), ctx.GetHeader("session"), ctx.GetHeader("token"), string(reqBody))

	logger.App().Infof("[%s][%s][%s][%s][%s] : %s",
		ctx.Request.Method, ctx.Request.URL.Path, ctx.ClientIP(), ctx.GetHeader("session"), ctx.GetHeader("token"), string(reqBody))

	// ==============================================
	ctx.Next()
	// ==============================================

	logger.App().Infof("[%s][%s][%s][%s][%s] : %s --- [%d][%d] : %s",
		ctx.Request.Method, ctx.Request.URL.Path, ctx.ClientIP(), ctx.GetHeader("session"), ctx.GetHeader("token"), string(reqBody),
		ctx.Writer.Status(), time.Since(startTime).Milliseconds(), respWriter.body.String())
}

func saveRequest(now int64, method, path, ip, session, token, reqData string) {
	switch path {
	case "/backend/login":
		{
			m := make(map[string]any)
			if err := sonic.Unmarshal([]byte(reqData), &m); err != nil {
				logger.App().Errorf("unmarshal req data to map error: %s", err.Error())
				return
			}
			username := m["account"]
			if username == "" {
				return
			}

			user := &structure.OperateUser{}
			if tx := mysql.Instance().Model(user).Where("username = ?", username).First(user); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				logger.App().Errorf("query [%s] error: [%s]-%s", username, tx.Error.Error())
				return
			}

			log := &structure.LoginLog{
				User:      user.Nickname,
				LoginTime: uint64(now),
				LoginIP:   ip,
			}
			if err := mysql.Instance().Model(log).Create(log).Error; err != nil {
				logger.App().Errorf("insert login log error: %s", err.Error())
			}
			return
		}
	default:
		{
			if token == "" {
				logger.App().Warnf("token is empty : %s - %s - %s", path, ip, reqData)
				return
			}

			user := &structure.OperateUser{}
			if tx := mysql.Instance().Model(user).Where("token = ?", token).First(user); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				logger.App().Errorf("query [%s] error: [%s]-%s", token, tx.Error.Error())
				return
			}

			log := &structure.OperateLog{
				User: user.Nickname,
				Desc: descMap[path],
				Path: path,
				Data: reqData,
			}
			if err := mysql.Instance().Model(log).Create(log).Error; err != nil {
				logger.App().Errorf("insert login log error: %s", err.Error())
			}
		}
	}
}

var descMap = map[string]string{
	"/backend/ik/insert":                 "分停词插入",
	"/backend/ik/update":                 "分停词修改",
	"/backend/ik/query":                  "分停词查询",
	"/backend/ik/delete":                 "分停词删除",
	"/backend/tg_spider_account/insert":  "爬虫账户插入",
	"/backend/tg_spider_account/update":  "爬虫账户修改",
	"/backend/tg_spider_account/query":   "爬虫账户查询",
	"/backend/tg_spider_account/delete":  "爬虫账户删除",
	"/backend/admin/user/insert":         "用户插入",
	"/backend/admin/user/update":         "用户修改",
	"/backend/admin/user/query":          "用户查询",
	"/backend/admin/user/delete":         "用户删除",
	"/backend/admin/user/addRole":        "用户增加角色",
	"/backend/admin/user/deleteRole":     "用户删除角色",
	"/backend/admin/menu/insert":         "菜单插入",
	"/backend/admin/menu/update":         "菜单修改",
	"/backend/admin/menu/query":          "菜单查询",
	"/backend/admin/menu/delete":         "菜单删除",
	"/backend/admin/role/insert":         "角色插入",
	"/backend/admin/role/update":         "角色修改",
	"/backend/admin/role/query":          "角色查询",
	"/backend/admin/role/delete":         "角色删除",
	"/backend/admin/role/queryTotal":     "角色查询全部角色",
	"/backend/admin/role/updateMenus":    "角色修改菜单权限",
	"/backend/admin/role/queryMenus":     "角色查询全部菜单权限",
	"/backend/login":                     "登录",
	"/backend/bindTFA":                   "绑定二次验证",
	"/backend/updateMe":                  "修改我的信息",
	"/backend/me":                        "获取我的信息",
	"/backend/meMenus":                   "获取我的菜单",
	"/backend/dashboard/info":            "仪表盘信息",
	"/backend/cog/insert":                "群组频道插入",
	"/backend/cog/update":                "群组频道修改",
	"/backend/cog/query":                 "群组频道查询",
	"/backend/cog/delete":                "群组频道删除",
	"/backend/ad_manage/keyword/insert":  "关键词插入",
	"/backend/ad_manage/keyword/update":  "关键词修改",
	"/backend/ad_manage/keyword/query":   "关键词查询",
	"/backend/ad_manage/keyword/delete":  "关键词修改",
	"/backend/ad_manage/keyword/addAd":   "关键词绑定广告",
	"/backend/ad_manage/keyword/queryAd": "关键词查询广告",
	"/backend/ad_manage/ad/insert":       "广告插入",
	"/backend/ad_manage/ad/update":       "广告修改",
	"/backend/ad_manage/ad/query":        "广告查询",
	"/backend/ad_manage/ad/delete":       "广告删除",
	"/backend/ad_manage/ad/total":        "全部广告",
	"/backend/ad_manage/client/insert":   "客户插入",
	"/backend/ad_manage/client/update":   "客户修改",
	"/backend/ad_manage/client/query":    "客户查询",
	"/backend/ad_manage/client/delete":   "客户删除",
	"/backend/ad_manage/client/total":    "客户全部"}
