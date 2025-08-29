package structure

import "jarvis/dao/db/mysql"

// Init 初始化
func Init() error {
	if err := mysql.Instance().Set("gorm:table_options", "ENGINE=InnoDB").AutoMigrate(
		new(Participle),
		new(TGSpiderAccount),
		new(COG),
		new(Keyword),
		new(Ad),
		new(Client),
		new(KeywordAd),
		new(AdClient),
		new(OperateMenu),
		new(OperateRole),
		new(OperateUser),
		new(OperateRoleMenu),
		new(OperateUserRole),
		new(LoginLog),
		new(OperateLog),
	); err != nil {
		return err
	}

	if err := mysql.Instance().Exec(`
CREATE TABLE IF NOT EXISTS search_log (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  created BIGINT NOT NULL COMMENT '时间戳(毫秒)',
  word VARCHAR(255) NOT NULL COMMENT '搜索词',
  PRIMARY KEY (id, created),
  KEY idx_create_time (created),
  KEY idx_word_create_time (word, created)
)
PARTITION BY RANGE (created) (
  PARTITION p20250630_20250709 VALUES LESS THAN (1752076800000)
);
	`).Error; err != nil {
		return err
	}

	if err := mysql.Instance().Exec(`
CREATE TABLE IF NOT EXISTS ad_log (
  id BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  created BIGINT NOT NULL COMMENT '时间戳(毫秒)',
  ad_id BIGINT NOT NULL COMMENT '广告ID',
  username VARCHAR(255) NOT NULL COMMENT '展示用户名',
  price DECIMAL(10,2) NOT NULL COMMENT '价格',

  PRIMARY KEY (id, created),
  KEY idx_create_time (created),
  KEY idx_ad_created (ad_id, created),
  KEY idx_username_created (username, created)
)
PARTITION BY RANGE (created) (
  PARTITION p20250630_20250709 VALUES LESS THAN (1752076800000)
);
	`).Error; err != nil {
		return err
	}

	return nil
}
