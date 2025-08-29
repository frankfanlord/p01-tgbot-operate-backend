package structure

type AdClient struct {
	ID       uint   `gorm:"primarykey" json:"id"`
	Updated  int64  `gorm:"autoUpdateTime:milli" json:"updated"`
	Created  int64  `gorm:"autoCreateTime:milli" json:"created"`
	AdID     uint64 `gorm:"column:ad_id;type:int(15);not null;index:idx_keyword_ad,unique;comment:广告ID" json:"ad_id"`
	ClientID uint64 `gorm:"column:client_id;type:int(15);not null;index:idx_keyword_ad,unique;comment:客户ID" json:"client_id"`
	Status   uint8  `gorm:"column:status;type:tinyint(1);default:1;comment:状态(1-启用 2-禁用)" json:"status"`
}

func (ac *AdClient) TableName() string {
	return "ad_client"
}
