package menu

import (
	"operate-backend/core/backend/login"
	"operate-backend/core/backend/middlewares"
	"operate-backend/core/structure"

	"github.com/gin-gonic/gin"
)

const ComponentName = "menu"

// Init 初始化
func Init(grouper *gin.RouterGroup) error {
	handler := grouper.Group(ComponentName).Use(login.SessionVerify, middlewares.HiJack)

	handler.POST(PathInsert, Insert)
	handler.POST(PathUpdate, Update)
	handler.POST(PathQuery, Query)
	handler.POST(PathDelete, Delete)

	// // 初始化菜单数据
	// list := []*structure.OperateMenu{
	// 	{ID: 1, Title: "仪表盘", ParentID: 0, Sort: 1},
	// 	{ID: 2, Title: "广告管理", ParentID: 0, Sort: 2},
	// 	{ID: 3, Title: "关键词", ParentID: 2, Sort: 21},
	// 	{ID: 4, Title: "广告信息", ParentID: 2, Sort: 22},
	// 	{ID: 5, Title: "客户信息", ParentID: 2, Sort: 23},
	// 	{ID: 6, Title: "频道管理", ParentID: 0, Sort: 3},
	// 	{ID: 7, Title: "爬虫管理", ParentID: 0, Sort: 4},
	// 	{ID: 8, Title: "权限管理", ParentID: 0, Sort: 5},
	// 	{ID: 9, Title: "用户管理", ParentID: 8, Sort: 51},
	// 	{ID: 10, Title: "菜单管理", ParentID: 8, Sort: 52},
	// 	{ID: 11, Title: "角色管理", ParentID: 8, Sort: 53},
	// 	{ID: 12, Title: "日志管理", ParentID: 0, Sort: 6},
	// 	{ID: 13, Title: "登录日志", ParentID: 12, Sort: 61},
	// 	{ID: 14, Title: "操作日志", ParentID: 12, Sort: 62},
	// }

	// var err error
	// for _, item := range list {
	// 	exist := int64(0)
	// 	if err := mysql.Instance().Model(new(structure.OperateMenu)).Where("title = ? and parent_id = ?", item.Title, item.ParentID).Count(&exist).Error; err != nil {
	// 		break
	// 	}
	// 	if exist == 0 {
	// 		if err = mysql.Instance().Create(item).Error; err != nil {
	// 			break
	// 		}
	// 	}
	// }
	//
	// return err
	return nil
}

// Start 启动
func Start() error { return nil }

// Shutdown 关闭
func Shutdown() error { return nil }

// LoadCache 加载缓存
func LoadCache() error { return nil }

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
