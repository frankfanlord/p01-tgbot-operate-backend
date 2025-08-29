package structure

// OperateMenu represents a menu structure in the operating system
type OperateMenu struct {
	ID               uint   `gorm:"primarykey" json:"id"`
	Updated          int64  `gorm:"autoUpdateTime:milli" json:"updated"`
	Created          int64  `gorm:"autoCreateTime:milli" json:"created"`
	Title            string `gorm:"column:title;type:varchar(50);not null;comment:菜单标题" json:"title"`
	ParentID         uint   `gorm:"column:parent_id;type:int(11);default:0;comment:父菜单ID" json:"parent_id"`
	Icon             string `gorm:"column:icon;type:varchar(50);default:'';comment:菜单图标" json:"icon"`
	Type             uint8  `gorm:"column:type;type:tinyint(1);default:1;comment:菜单类型(1-目录 2-按钮)" json:"type"`
	Path             string `gorm:"column:path;type:varchar(255);default:'';comment:菜单路径" json:"path"`
	Name             string `gorm:"column:name;type:varchar(255);default:'';comment:菜单名称" json:"name"`
	Remark           string `gorm:"column:remark;type:varchar(255);default:'';comment:备注" json:"remark"`
	Sort             int    `gorm:"column:sort;type:int(11);default:0;comment:排序" json:"sort"`
	Affix            bool   `gorm:"column:affix;type:tinyint(1);default:false;comment:Tab是否固定" json:"affix"`
	Cache            bool   `gorm:"column:cache;type:tinyint(1);default:false;comment:是否缓存" json:"cache"`
	Hidden           bool   `gorm:"column:hidden;type:tinyint(1);default:false;comment:是否隐藏" json:"hidden"`
	BreadcrumbEnable bool   `gorm:"column:breadcrumb_enable;type:tinyint(1);default:false;comment:是否显示面包屑" json:"breadcrumbEnable"`
	Component        string `gorm:"column:component;type:varchar(255);default:'';comment:路由组件" json:"component"`
	Status           uint8  `gorm:"column:status;type:tinyint(1);default:1;comment:状态(1-启用 2-禁用)" json:"status"`
}

func (om *OperateMenu) TableName() string {
	return "operate_menu"
}
