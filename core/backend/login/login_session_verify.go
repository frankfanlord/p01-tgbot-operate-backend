package login

import (
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
)

const (
	CTXHeaderTokenKey   = "token"
	CTXHeaderSessionKey = "session"
	CTXUserKey          = "_USER_"
)

func SessionVerify(ctx *gin.Context) {
	token := ctx.GetHeader(CTXHeaderTokenKey)
	session := ctx.GetHeader(CTXHeaderSessionKey)

	if token == "" || session == "" {
		trace, response := define.Response(define.CodeUnAuthenticate, nil)
		logger.App().Warnf("request un authenticate [%s] : %s - %s - %s", ctx.Request.URL.Path, token, session, trace)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// verify token-session match and not expired
	sToken, isExpire := SMInstance().Load(session)
	if sToken == "" || sToken != token || isExpire {
		trace, response := define.Response(define.CodeUnAuthenticate, nil)
		logger.App().Warnf("request un authenticate [%s] : %s - %s - %s - %s - %t", ctx.Request.URL.Path, sToken, token, session, trace, isExpire)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	} else {
		user := new(structure.OperateUser)
		if err := mysql.Instance().Where(user).Where("token = ?", token).First(user).Error; err != nil {
			logger.App().Warnf("query [%s] token user error : %s", token, err.Error())
		} else {
			ctx.Set(CTXUserKey, user)
		}
	}

	// reflush this session expire-time
	SMInstance().Keep(session)
}
