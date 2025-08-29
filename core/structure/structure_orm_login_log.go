package structure

type LoginLog struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	Updated   int64  `gorm:"autoUpdateTime:milli" json:"updated"`
	Created   int64  `gorm:"autoCreateTime:milli" json:"created"`
	User      string `gorm:"column:user;type:varchar(100);not null;comment:昵称" json:"user"`
	LoginTime uint64 `gorm:"column:login_time;type:int(15);not null;comment:登录时间" json:"login_time"`
	LoginIP   string `gorm:"olumn:login_ip;varchar(30);not null;comment:登录IP" json:"login_ip"`
}

func (ll *LoginLog) TableName() string {
	return "login_log"
}
