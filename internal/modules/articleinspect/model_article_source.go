package articleinspect

import (
	"database/sql"
	"time"
)

// ChuangqiOrg 和 Article / ArticleInfo 一样，属于上游真实业务表，不是本项目自建巡检表。
type ChuangqiOrg struct {
	ID       uint64    `gorm:"column:id;primaryKey" json:"id"`
	Name     string    `gorm:"column:name;size:128;not null" json:"name"`
	CateID   uint64    `gorm:"column:cateid;not null;default:0" json:"cate_id"`
	Enabled  bool      `gorm:"column:enabled;not null;default:true" json:"enabled"`
	Sort     int64     `gorm:"column:sort;not null;default:0" json:"sort"`
	CreateAt time.Time `gorm:"column:create_at;not null;autoCreateTime" json:"created_at"`
	UpdateAt time.Time `gorm:"column:update_at;not null;autoUpdateTime" json:"updated_at"`
}

func (ChuangqiOrg) TableName() string {
	return "xt_chuangqi_org"
}

// Article / ArticleInfo 都是上游真实文稿表，巡检逻辑会直接读取它们，不由本项目迁移创建。
type Article struct {
	ID            uint64         `gorm:"column:id;primaryKey" json:"id"`
	OrgID         uint64         `gorm:"column:orgid;not null;index" json:"orgid"`
	Title         string         `gorm:"column:title;size:240;not null;default:''" json:"title"`
	ShortTitle    sql.NullString `gorm:"column:short_title;size:100" json:"short_title"`
	RichTitle     string         `gorm:"column:rich_title;size:1000;not null;default:''" json:"rich_title"`
	Keyword       string         `gorm:"column:keyword;size:240;not null;default:''" json:"keyword"`
	Desc          string         `gorm:"column:desc;size:250;not null;default:''" json:"desc"`
	Thumbnail     string         `gorm:"column:thumbnail;size:600;not null;default:''" json:"thumbnail"`
	State         int8           `gorm:"column:state;not null;default:0" json:"state"`
	PublishAtUnix int64          `gorm:"column:publish_at_time;not null;default:0" json:"publish_at_time"`
	UpdateAtUnix  int64          `gorm:"column:update_at;not null;default:0" json:"update_at"`
}

func (Article) TableName() string {
	return "xt_article"
}

// ArticleInfo 用主键直接对应文稿 ID，这也是 worker 读取正文时的关键约定。
type ArticleInfo struct {
	ID        uint64 `gorm:"column:id;primaryKey" json:"id"`
	OrgID     uint64 `gorm:"column:orgid;not null;default:0" json:"orgid"`
	Body      string `gorm:"column:body;type:longtext" json:"body"`
	Media     string `gorm:"column:media;type:text" json:"media"`
	Relate    string `gorm:"column:relate;type:text" json:"relate"`
	ShareInfo string `gorm:"column:share_info;type:text" json:"share_info"`
}

func (ArticleInfo) TableName() string {
	return "xt_article_info"
}
