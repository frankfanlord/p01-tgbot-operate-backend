package analysis_ik

import (
	"fmt"
	"github.com/duke-git/lancet/v2/cryptor"
	"github.com/gin-gonic/gin"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/structure"
	"strings"
	"time"
)

func ExtDict(ctx *gin.Context) {
	list := make([]string, 0)
	if tx := mysql.Instance().Model(new(structure.Participle)).Select("word").Where("type = 1 AND status = 1").Find(&list); tx.Error != nil {
		logger.App().Errorf("query all error : %s", tx.Error.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	tmp := &struct {
		Max int64 `json:"max"`
	}{}
	if tx := mysql.Instance().Model(new(structure.Participle)).Select("max(distinct updated) as max").Where("type = 1 AND status = 1").First(tmp); tx.Error != nil {
		logger.App().Errorf("query all error : %s", tx.Error.Error())
		ctx.JSON(http.StatusInternalServerError, nil)
		return
	}

	ctx.Header("Content-Type", "text/plain; charset=utf-8")
	ctx.Header("Last-Modified", time.UnixMilli(tmp.Max).Format(time.RFC1123))
	ctx.Header("ETag", cryptor.Md5String(fmt.Sprintf("%d", tmp.Max)))
	ctx.Data(http.StatusOK, "text/plain; charset=utf-8", []byte(strings.Join(list, "\n")))
}
