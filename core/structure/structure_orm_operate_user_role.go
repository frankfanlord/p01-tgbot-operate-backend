package structure

// OperateUserRole represents a user-role association structure in the operating system
type OperateUserRole struct {
	ID      uint  `gorm:"primarykey" json:"id"`
	Updated int64 `gorm:"autoUpdateTime:milli" json:"updated"`
	Created int64 `gorm:"autoCreateTime:milli" json:"created"`
	UserID  uint  `gorm:"column:user_id;type:int(11);index:idx_user_role,unique;comment:用户ID" json:"user_id"`
	RoleID  uint  `gorm:"column:role_id;type:int(11);index:idx_user_role,unique;comment:角色ID" json:"role_id"`
}

func (our *OperateUserRole) TableName() string {
	return "operate_user_role"
}
