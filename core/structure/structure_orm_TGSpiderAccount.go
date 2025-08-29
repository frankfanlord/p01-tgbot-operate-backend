package structure

import (
	"jarvis/dao/db/mysql"
)

type TGSpiderAccount struct {
	ID      uint   `gorm:"primarykey" json:"id"`
	Updated int64  `gorm:"autoUpdateTime:milli" json:"updated"`
	Created int64  `gorm:"autoCreateTime:milli" json:"created"`
	Phone   string `gorm:"column:phone;type:varchar(20);not null;index:idx_phone_app_id_hash,unique;comment:手机号码" json:"phone"`
	Code    string `gorm:"column:code;type:varchar(10);comment:验证码" json:"code"`
	TFAPwd  string `gorm:"column:tfa_pwd;type:varchar(20);comment:2FA密码" json:"tfa_pwd"`
	AppID   uint64 `gorm:"column:app_id;type:bigint(10);not null;index:idx_phone_app_id_hash,unique;comment:appid" json:"app_id"`
	AppHash string `gorm:"column:app_hash;type:varchar(50);not null;index:idx_phone_app_id_hash,unique;comment:appHash" json:"app_hash"`
	Session string `gorm:"column:session;type:varchar(10000);comment:缓存" json:"session"`
	Status  uint8  `gorm:"column:status;type:tinyint(1);default:1;comment:状态(1-停用 2-启用)" json:"status"`
	Process uint8  `gorm:"column:process;type:tinyint(1);default:1;comment:进程(1-无 2-下发 3-确认下发 4-请求验证码 5-已登录 6-进行中 7-停止中)" json:"process"`
}

func (p *TGSpiderAccount) TableName() string {
	return "tg_spider_account"
}

// UpdateProcess 修改进度
func UpdateProcess(id uint, process uint8) error {
	if tx := mysql.Instance().Model(new(TGSpiderAccount)).Where("id = ?", id).UpdateColumn("process", process); tx.Error != nil {
		return tx.Error
	}
	return nil
}
