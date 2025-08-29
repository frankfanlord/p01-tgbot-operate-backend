package structure

// OperateRole represents a role structure in the operating system
type OperateRole struct {
	ID      uint   `gorm:"primarykey" json:"id"`
	Updated int64  `gorm:"autoUpdateTime:milli" json:"updated"`
	Created int64  `gorm:"autoCreateTime:milli" json:"created"`
	Name    string `gorm:"column:name;type:varchar(50);not null;unique;comment:角色名称" json:"name"`
	Code    string `gorm:"column:code;type:varchar(50);not null;comment:角色编码" json:"code"`
	Remark  string `gorm:"column:remark;type:varchar(255);default:'';comment:备注" json:"remark"`
	Creator uint   `gorm:"column:creator;type:int(11);not null;comment:创建角色用户ID" json:"creator"`
	Status  uint8  `gorm:"column:status;type:tinyint(1);default:1;comment:状态(1-启用 2-禁用)" json:"status"`
}

func (or *OperateRole) TableName() string {
	return "operate_role"
}
