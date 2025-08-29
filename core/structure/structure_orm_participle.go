package structure

type Participle struct {
	ID      uint   `gorm:"primarykey" json:"id"`
	Updated int64  `gorm:"autoUpdateTime:milli" json:"updated"`
	Created int64  `gorm:"autoCreateTime:milli" json:"created"`
	Word    string `gorm:"column:word;type:varchar(20);not null;uniqueIndex:idx_word;comment:词" json:"word"`
	Type    uint8  `gorm:"column:type;type:tinyint(1);default:1;comment:类型(1-分词 2-停词)" json:"type"`
	Status  uint8  `gorm:"column:status;type:tinyint(1);default:1;comment:状态(1-启用 2-禁用)" json:"status"`
}

func (p *Participle) TableName() string {
	return "participle"
}
