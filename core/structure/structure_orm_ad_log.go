package structure

type ADLog struct {
	ID       uint    `gorm:"column:id" json:"id"`
	Created  int64   `gorm:"column:created" json:"created"`
	AdID     uint64  `gorm:"column:ad_id" json:"word"`
	Username string  `gorm:"column:username" json:"username"`
	Price    float64 `gorm:"column:price" json:"price"`
}

func (ad *ADLog) TableName() string {
	return "ad_log"
}
