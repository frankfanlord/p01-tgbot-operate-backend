package user

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
	Username string `json:"username"`
	Nickname string `json:"nickname"`
	Page     uint32 `json:"page"`      // 页码(from 1)
	PageSize uint32 `json:"page_size"` // 每页
}

type QueryItem struct {
	structure.OperateUser
	RoleID   uint   `json:"role_id"`
	RoleName string `json:"role_name"`
}

type QueryRsp struct {
	Page     uint32      `json:"page"`
	PageSize uint32      `json:"page_size"`
	Total    uint64      `json:"total"`
	List     []QueryItem `json:"list"`
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

	tx := mysql.Instance().Model(new(structure.OperateUser))

	if req.Username != "" {
		tx = tx.Where(fmt.Sprintf("username LIKE'%%%s%%'", req.Username))
	}
	if req.Nickname != "" {
		tx = tx.Where(fmt.Sprintf("nickname LIKE'%%%s%%'", req.Nickname))
	}
	if !admin && children != nil {
		tx = tx.Where("id in ?", children)
	}

	total := int64(0)
	if tmp := tx.Count(&total); tmp.Error != nil {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Query total error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	atx := mysql.Instance().Table(fmt.Sprintf("%s as a", new(structure.OperateUser).TableName())).Select("a.*,c.id as role_id,c.name as role_name").
		Joins(fmt.Sprintf("left join %s as b on b.user_id = a.id", new(structure.OperateUserRole).TableName())).
		Joins(fmt.Sprintf("left join %s as c on c.id = b.role_id", new(structure.OperateRole).TableName()))

	if req.Username != "" {
		atx = atx.Where(fmt.Sprintf("a.username LIKE'%%%s%%'", req.Username))
	}
	if req.Nickname != "" {
		atx = atx.Where(fmt.Sprintf("a.nickname LIKE'%%%s%%'", req.Nickname))
	}
	if !admin && children != nil {
		atx = atx.Where("a.id in ?", children)
	}

	list := make([]QueryItem, 0)
	if tx = atx.Offset((page - 1) * limit).Limit(limit).Order("a.id desc").Find(&list); tx.Error != nil {
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
