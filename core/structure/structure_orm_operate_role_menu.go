package structure

// OperateRoleMenu represents a role-menu association structure in the operating system
type OperateRoleMenu struct {
	ID      uint  `gorm:"primarykey" json:"id"`
	Updated int64 `gorm:"autoUpdateTime:milli" json:"updated"`
	Created int64 `gorm:"autoCreateTime:milli" json:"created"`
	RoleID  uint  `gorm:"column:role_id;type:int(11);index:idx_role_menu,unique;comment:角色ID" json:"role_id"`
	MenuID  uint  `gorm:"column:menu_id;type:int(11);index:idx_role_menu,unique;comment:菜单ID" json:"menu_id"`
}

func (orm *OperateRoleMenu) TableName() string {
	return "operate_role_menu"
}
