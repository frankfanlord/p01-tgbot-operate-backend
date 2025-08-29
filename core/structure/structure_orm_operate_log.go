package structure

type OperateLog struct {
	ID      uint   `gorm:"primarykey" json:"id"`
	Updated int64  `gorm:"autoUpdateTime:milli" json:"updated"`
	Created int64  `gorm:"autoCreateTime:milli" json:"created"`
	User    string `gorm:"column:user;type:varchar(100);not null;comment:昵称" json:"user"`
	Desc    string `gorm:"column:desc;type:varchar(100);not null;comment:昵称" json:"desc"`
	Path    string `gorm:"column:path;type:varchar(100);not null;comment:昵称" json:"path"`
	Data    string `gorm:"column:data;type:varchar(5000);not null;comment:昵称" json:"data"`
}

func (ll *OperateLog) TableName() string {
	return "operate_log"
}
