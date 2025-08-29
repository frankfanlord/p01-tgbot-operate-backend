package login

import (
	"errors"
	"fmt"
	"jarvis/dao/db/mysql"
	"jarvis/logger"
	"net/http"
	"operate-backend/core/define"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type MeMenusReq struct{}

type MenuItem struct {
	structure.OperateMenu
	Children []*MenuItem `json:"children,omitempty"` // 子菜单
}

func (mi *MenuItem) AddChild(item *MenuItem) {
	if item.ParentID == mi.ID {
		mi.Children = append(mi.Children, item)
		return
	}

	for _, child := range mi.Children {
		child.AddChild(item)
	}
}

type MeMenusRsp struct {
	List []*MenuItem `json:"list"`
}

const PathMeMenus = "meMenus"

func MeMenus(ctx *gin.Context) {
	admin := false
	v, e := ctx.Get(CTXUserKey)
	if e {
		user := v.(*structure.OperateUser)
		admin = user.UserType == 2
	}

	token := ctx.GetHeader(CTXHeaderTokenKey)

	// 1.先查询有没有
	old := new(structure.OperateUser)
	if tx := mysql.Instance().Model(new(structure.OperateUser)).Where("token = ?", token).First(old); tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		trace, response := define.Response(define.CodeSvrInternalError, nil)
		logger.App().Errorf("Update error: [%s]-%s", trace, tx.Error.Error())
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}
	if old.ID == 0 {
		_, response := define.ResponseMsg(define.CodeParamError, "the token of user is not exist", nil)
		ctx.AbortWithStatusJSON(http.StatusOK, response)
		return
	}

	list := make([]structure.OperateMenu, 0)
	if !admin {
		if err := mysql.Instance().Table(fmt.Sprintf("%s as a", new(structure.OperateUser).TableName())).Select("e.*").
			Joins(fmt.Sprintf("left join %s as b on b.user_id = a.id", new(structure.OperateUserRole).TableName())).
			Joins(fmt.Sprintf("left join %s as c on c.id = b.role_id", new(structure.OperateRole).TableName())).
			Joins(fmt.Sprintf("left join %s as d on d.role_id = c.id", new(structure.OperateRoleMenu).TableName())).
			Joins(fmt.Sprintf("left join %s as e on e.id = d.menu_id", new(structure.OperateMenu).TableName())).
			Where("a.token = ? AND e.id IS NOT NULL", token).Order("e.sort,e.id").Find(&list).Error; err != nil {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("Query error: [%s]-%s", trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	} else {
		if err := mysql.Instance().Model(new(structure.OperateMenu)).Order("sort,id").Find(&list).Error; err != nil {
			trace, response := define.Response(define.CodeSvrInternalError, nil)
			logger.App().Errorf("Query total error: [%s]-%s", trace, err.Error())
			ctx.AbortWithStatusJSON(http.StatusOK, response)
			return
		}
	}

	root := &MenuItem{Children: make([]*MenuItem, 0)}

	for _, item := range list {
		root.AddChild(&MenuItem{OperateMenu: item, Children: make([]*MenuItem, 0)})
	}

	_, response := define.Response(define.CodeSuccess, root.Children[:])
	ctx.AbortWithStatusJSON(http.StatusOK, response)
}
