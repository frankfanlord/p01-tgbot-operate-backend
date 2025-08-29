package structure

// OperateRole represents a role structure in the operating system
type OperateUser struct {
	ID        uint   `gorm:"primarykey" json:"id"`
	Updated   int64  `gorm:"autoUpdateTime:milli" json:"updated"`
	Created   int64  `gorm:"autoCreateTime:milli" json:"created"`
	Username  string `gorm:"column:username;type:varchar(50);not null;unique;comment:用户名" json:"username"`
	Nickname  string `gorm:"column:nickname;type:varchar(50);not null;comment:昵称" json:"nickname"`
	Token     string `gorm:"column:token;type:varchar(50);not null;uniquelcomment:token" json:"-"`
	UserType  uint8  `gorm:"column:user_type;type:tinyint(1);default:1;comment:用户类型(1-普通用户 2-管理员)" json:"user_type"`
	LoginIP   string `gorm:"column:login_ip;type:varchar(50);default:'';comment:登录IP" json:"login_ip"`
	LoginTime int64  `gorm:"column:login_time;type:int(11);default:0;comment:登录时间" json:"login_time"`
	Password  string `gorm:"column:password;type:varchar(255);not null;comment:密码" json:"-"`
	Remark    string `gorm:"column:remark;type:varchar(255);default:'';comment:备注" json:"remark"`
	TFASalt   string `gorm:"column:tfa_salt;type:varchar(20);default:'';comment:2FA盐值" json:"-"`
	TFAStatus uint8  `gorm:"column:tfa_status;type:tinyint(1);default:1;comment:2FA状态(1-未启用 2-已启用 3-启用未绑定 4-启用已绑定)" json:"tfa_status"`
	ParentID  uint   `gorm:"column:parent_id;type:int(11);not null;comment:父级ID" json:"parent_id"`
	Status    uint8  `gorm:"column:status;type:tinyint(1);default:1;comment:状态(1-启用 2-禁用)" json:"status"`
}

func (ou *OperateUser) TableName() string {
	return "operate_user"
}
