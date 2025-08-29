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
	"net/url"
	"operate-backend/core/define"
	"operate-backend/core/structure"
	"time"

	"github.com/duke-git/lancet/v2/random"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

const PathLogin = "login"

type LoginReq struct {
	Account  string `json:"account"`
	Password string `json:"password"`
	TFACode  string `json:"tfa_code"`
}

type LoginRsp struct {
	Status  uint8  `json:"status"` // 0-success 1-need bind 2-need code
	TOTPUrl string `json:"totp_url"`
	Token   string `json:"token"`
	Session string `json:"session"`
}

func Login(ctx *gin.Context) {
	var req LoginReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("ShouldBindJSON error: [%s]-%s", trace, err.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	if req.Account == "" || req.Password == "" {
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

	if user.Password != req.Password {
		_, response := define.ResponseMsg(define.CodeParamError, "the account is not match password", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	rsp := &LoginRsp{}

	// 2.tfa
	switch user.TFAStatus {
	case 1:
		{
			rsp.Token = user.Token
			rsp.Session = random.RandString(32)
			SMInstance().Store(rsp.Token, rsp.Session)
		}
	case 2, 3:
		{
			// need bind TFA
			rsp.Status = 1

			secret := base32.StdEncoding.EncodeToString([]byte(user.TFASalt))

			account := fmt.Sprintf("%s:%s", user.Nickname, user.Username)
			params := url.Values{}
			params.Set("secret", secret)
			params.Set("issuer", user.Username)

			rsp.TOTPUrl = fmt.Sprintf("otpauth://totp/%s?%s", url.QueryEscape(account), params.Encode())
		}
	case 4:
		{
			// todo : verify TFA code
			if req.TFACode == "" {
				rsp.Status = 2
				break
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

			rsp.Token = user.Token
			rsp.Session = random.RandString(32)
			SMInstance().Store(rsp.Token, rsp.Session)
		}
	}

	_, response := define.Response(define.CodeSuccess, rsp)
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
