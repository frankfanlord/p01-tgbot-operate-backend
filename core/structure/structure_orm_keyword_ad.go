package structure

type KeywordAd struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	Updated   int64  `gorm:"autoUpdateTime:milli" json:"updated"`
	Created   int64  `gorm:"autoCreateTime:milli" json:"created"`
	KeywordID uint64 `gorm:"column:keyword_id;type:int(15);not null;index:idx_keyword_ad,unique;comment:关键词ID" json:"keyword_id"`
	AdID      uint64 `gorm:"column:ad_id;type:int(15);not null;index:idx_keyword_ad,unique;comment:广告ID" json:"ad_id"`
	Status    uint8  `gorm:"column:status;type:tinyint(1);default:1;comment:状态(1-启用 2-禁用)" json:"status"`
}

func (ka *KeywordAd) TableName() string {
	return "keyword_ad"
}
