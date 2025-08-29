package role

import (
	"fmt"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/backend/login"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
)

type QueryReq struct {
	Name     string `json:"name"`
	Status   int    `json:"status"`
	Page     uint32 `json:"page"`      // 页码(from 1)
	PageSize uint32 `json:"page_size"` // 每页
}

type QueryRsp struct {
	Page     uint32                  `json:"page"`
	PageSize uint32                  `json:"page_size"`
	Total    uint64                  `json:"total"`
	List     []structure.OperateRole `json:"list"`
}

const PathQuery = "query"

func Query(ctx *gin.Context) {
	pid, admin := 0, false
	v, e := ctx.Get(login.CTXUserKey)
	if e {
		user := v.(*structure.OperateUser)
		pid = int(user.ID)
		admin = user.UserType == 2
	}

	var req QueryReq
	if err := ctx.ShouldBindJSON(&req); err != nil {
		trace, response := define.Response(define.CodeParamError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
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

	children, gsErr := getAllSubUserIDs(pid)
	if gsErr != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query all sub user ids error: [%s]-%s", trace, gsErr.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	tx := mysql.Instance().Model(new(structure.OperateRole))

	if req.Name != "" {
		tx = tx.Where(fmt.Sprintf("name LIKE'%%%s%%'", req.Name))
	}
	if req.Status != 0 {
		tx = tx.Where("status = ?", req.Status)
	}
	if !admin && children != nil {
		tx = tx.Where("creator in ?", children)
	}

	total := int64(0)
	if tmp := tx.Count(&total); tmp.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query total error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	list := make([]structure.OperateRole, 0)
	if tx = tx.Offset((page - 1) * limit).Limit(limit).Order("id desc").Find(&list); tx.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	_, response := define.Response(define.CodeSuccess, QueryRsp{
		Page:     uint32(page),
		PageSize: uint32(limit),
		Total:    uint64(total),
		List:     list[:],
	})
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}

func getAllSubUserIDs(userID int) ([]int, error) {
	var ids []int
	query := `
	WITH RECURSIVE user_tree AS (
		SELECT id FROM operate_user WHERE id = ?
		UNION ALL
		SELECT ou.id FROM operate_user ou
		INNER JOIN user_tree ut ON ou.parent_id = ut.id
	)
	SELECT id FROM user_tree
	`
	if err := mysql.Instance().Raw(query, userID).Scan(&ids).Error; err != nil {
		return nil, err
	}
	return ids, nil
}
