package articleinspect

import "time"

type OrgDTO struct {
	ID      uint64 `json:"id"`
	Name    string `json:"name"`
	CateID  uint64 `json:"cate_id"`
	Enabled bool   `json:"enabled"`
	Sort    int64  `json:"sort"`
}

type OrgListResult struct {
	Items []OrgDTO `json:"items"`
}

type CategoryListInput struct {
	OrgID    uint64
	Page     int
	PageSize int
	Enabled  *bool
	Query    string
}

type CreateCategoryInput struct {
	OrgID   uint64
	Name    string
	Enabled bool
	Sort    int64
}

type UpdateCategoryInput struct {
	ID uint64
	CreateCategoryInput
}

type PatchCategoryStatusInput struct {
	OrgID      uint64
	CategoryID uint64
	Enabled    bool
}

type CategoryDTO struct {
	ID          uint64    `json:"id"`
	OrgID       uint64    `json:"orgid"`
	Name        string    `json:"name"`
	Enabled     bool      `json:"enabled"`
	Sort        int64     `json:"sort"`
	CreatorID   uint64    `json:"creator_id"`
	CreatorName string    `json:"creator_name"`
	UpdaterID   uint64    `json:"updater_id"`
	UpdaterName string    `json:"updater_name"`
	CreateAt    time.Time `json:"created_at"`
	UpdateAt    time.Time `json:"updated_at"`
}

type CategoryListResult struct {
	Page     int           `json:"page"`
	PageSize int           `json:"page_size"`
	Total    int64         `json:"total"`
	Items    []CategoryDTO `json:"items"`
}
