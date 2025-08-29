package structure

type SearchLog struct {
	ID      uint   `gorm:"column:id;not null;autoIncrement:false;primaryKey;comment:主键ID" json:"id"`
	Created int64  `gorm:"column:created;not null;index:idx_create_time;index:idx_word_create_time,priority:2;comment:时间戳(毫秒);primaryKey" json:"created"`
	Word    string `gorm:"column:word;type:varchar(255);not null;index:idx_word_create_time,priority:1;comment:搜索词" json:"word"`
}

func (sl *SearchLog) TableName() string {
	return "search_log"
}
