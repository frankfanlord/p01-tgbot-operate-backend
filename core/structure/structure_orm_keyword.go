package structure

type Keyword struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Updated  int64  `gorm:"autoUpdateTime:milli" json:"updated"`
	Created  int64  `gorm:"autoCreateTime:milli" json:"created"`
	Word     string `gorm:"column:word;type:varchar(100);not null;unique;comment:关键词" json:"word"`
	ParentID uint64 `gorm:"column:parent_id;type:int(15);default:0;comment:上级关键词ID" json:"parent_id"`
	Level    uint8  `gorm:"olumn:level;type:tinyint(1);default:1;comment:层级(从1开始)" json:"level"`
	Status   uint8  `gorm:"column:status;type:tinyint(1);default:1;comment:状态(1-启用 2-禁用)" json:"status"`
}

func (k *Keyword) TableName() string {
	return "keyword"
}
