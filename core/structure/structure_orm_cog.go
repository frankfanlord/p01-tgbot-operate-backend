package structure

type COG struct {
	ID         uint   `gorm:"primarykey" json:"id"`
	Updated    int64  `gorm:"autoUpdateTime:milli" json:"updated"`
	Created    int64  `gorm:"autoCreateTime:milli" json:"created"`
	Username   string `gorm:"column:username;type:varchar(50);not null;unique;comment:唯一标识" json:"username"`
	TGID       uint64 `gorm:"column:tg_id;type:bigint unsigned;default:0;comment:tgID" json:"tg_id"`
	AccessID   int64  `gorm:"column:access_id;type:bigint;default:0;comment:tg访问ID" json:"access_id"`
	Title      string `gorm:"column:title;type:varchar(500);default:'';comment:名称" json:"title"`
	About      string `gorm:"column:about;type:varchar(5000);default:'';comment:介绍" json:"about"`
	Type       uint8  `gorm:"column:type;type:tinyint(1);default:1;comment:状态(1-频道 2-群组)" json:"type"`
	Members    uint64 `gorm:"column:members;type:int(10);default:0;comment:成员数量" json:"members"`
	LinkChatID uint64 `gorm:"column:link_chat_id;type:bigint unsigned;default:0;comment:关联讨论组" json:"link_chat_id"`
	Remark     string `gorm:"column:remark;type:varchar(255);default:'';comment:备注" json:"remark"`
	Category   uint8  `gorm:"column:category;type:tinyint(2);default:1;comment:类别" json:"category"`
	Status     uint8  `gorm:"column:status;type:tinyint(1);default:1;comment:状态(1-新增 2-确认中 3-已确认)" json:"status"`
}

func (cog *COG) TableName() string {
	return "cog"
}
