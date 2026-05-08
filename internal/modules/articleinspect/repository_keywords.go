package articleinspect

import (
	rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"
	"gorm.io/gorm"
)

type KeywordRepository = rulespkg.KeywordRepository

type KeywordListFilter = rulespkg.KeywordListFilter

type KeywordRecord = rulespkg.KeywordRecord

func NewKeywordRepository(db *gorm.DB) *KeywordRepository {
	return rulespkg.NewKeywordRepository(db)
}
