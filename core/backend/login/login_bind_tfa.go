package login

import (
	"crypto/hmac"
	"crypto/sha1"
	"encoding/base32"
	"encoding/binary"
	"errors"
	"fmt"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const PathBindTFA = "bindTFA"

type BindTFAReq struct {
	Account string `json:"account"`
	TFACode string `json:"tfa_code"`
}

type BindTFARsp struct{}

func BindTFA(ctx *gin.Context) {
	var req BindTFAReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("ShouldBindJSON error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Account == "" || req.TFACode == "" {
		_, response := define.Response(define.CodeParamError, nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 1.query
	user := new(structure.OperateUser)
	if err := mysql.Instance().Model(user).Where("username = ?", req.Account).First(user).Error; err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("query [%s] error: [%s]-%s", req.Account, trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if user.ID == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the account is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if user.TFAStatus == 1 {
		_, response := define.ResponseMsg(define.CodeParamError, "the account is no need bind TFA", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if user.TFAStatus == 4 {
		_, response := define.ResponseMsg(define.CodeParamError, "the account is already bind TFA", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	secret := base32.StdEncoding.EncodeToString([]byte(user.TFASalt))

	// 将时间戳转换为30秒的时间窗口
	timeStep := time.Now().Unix() / 30

	// 解码密钥
	key, err := base32.StdEncoding.DecodeString(secret)
	if err != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("DecodeString error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	// 转换时间步长为8字节
	timeBytes := make([]byte, 8)
	binary.BigEndian.PutUint64(timeBytes, uint64(timeStep))

	// 使用HMAC-SHA1计算哈希
	h := hmac.New(sha1.New, key)
	h.Write(timeBytes)
	hash := h.Sum(nil)

	// 动态截取
	offset := hash[len(hash)-1] & 0x0F
	code := binary.BigEndian.Uint32(hash[offset:offset+4]) & 0x7FFFFFFF

	// 生成6位数字验证码
	inCode := fmt.Sprintf("%06d", code%1000000)

	if inCode != req.TFACode {
		_, response := define.ResponseMsg(define.CodeParamError, "the 2FACode is wrong", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if tx := mysql.Instance().Model(new(structure.OperateUser)).Where("username = ?", req.Account).UpdateColumn("tfa_status", 4); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, nil)
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
