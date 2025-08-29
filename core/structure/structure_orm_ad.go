package structure

type Ad struct {
	ID             uint    `gorm:"primarykey" json:"id"`
	Updated        int64   `gorm:"autoUpdateTime:milli" json:"updated"`
	Created        int64   `gorm:"autoCreateTime:milli" json:"created"`
	Title          string  `gorm:"column:title;type:varchar(200);not null;comment:广告标题" json:"title"`
	Link           string  `gorm:"column:link;type:varchar(500);not null;comment:广告链接" json:"link"`
	ClientID       uint64  `gorm:"column:client_id;type:int(15);default:0;comment:客户ID" json:"client_id"`
	PricePerView   float64 `gorm:"column:price_per_view;type:decimal(10,2);default:0.00;comment:每次展示价格" json:"price_per_view"`
	MaxImpressions uint64  `gorm:"column:max_impressions;type:int(15);default:0;comment:最多展示次数" json:"max_impressions"`
	Impressions    uint64  `gorm:"column:impressions;type:int(15);default:0;comment:已展示次数" json:"impressions"`
	StartTime      uint64  `gorm:"column:start_time;type:int(15);default:0;comment:开始时间" json:"start_time"`
	StopTime       uint64  `gorm:"column:stop_time;type:int(15);default:0;comment:结束时间" json:"stop_time"`
	Type           uint8   `gorm:"column:type;tinyint(1);default:1;comment:广告类型(1-关键词广告 2-置顶广告 3-搜索內连大广告 4-搜索內连小广告 5-搜索内容广告)" json:"type"`
	Status         uint8   `gorm:"column:status;type:tinyint(1);default:1;comment:状态(1-启用 2-禁用)" json:"status"`
}

func (ad *Ad) TableName() string {
	return "ad"
}
