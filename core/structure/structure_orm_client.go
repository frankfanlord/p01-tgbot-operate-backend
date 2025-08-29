package structure

type Client struct {
	ID        uint    `gorm:"primarykey" json:"id"`
	Updated   int64   `gorm:"autoUpdateTime:milli" json:"updated"`
	Created   int64   `gorm:"autoCreateTime:milli" json:"created"`
	Code      string  `gorm:"column:code;type:varchar(100);not null;unique;comment:客户编码" json:"code"`
	Name      string  `gorm:"column:name;type:varchar(100);not null;comment:客户名称" json:"name"`
	TGAccount string  `gorm:"column:tg_account;type:varchar(100);default:'';comment:TG账号" json:"tg_account"`
	Balance   float64 `gorm:"column:balance;type:decimal(10,2);default:0.00;comment:余额" json:"balance"`
	Spent     float64 `gorm:"column:spent;type:decimal(10,2);default:0.00;comment:已花费金额" json:"spent"`
	AdCount   uint64  `gorm:"column:ad_count;type:int(15);default:0;comment:投放广告数量" json:"ad_count"`
	Status    uint8   `gorm:"column:status;type:tinyint(1);default:1;comment:状态(1-启用 2-禁用)" json:"status"`
}

func (c *Client) TableName() string {
	return "client"
}
