package articleinspect

import (
	resultspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/results"
	"gorm.io/gorm"
)

type ResultRepository = resultspkg.ResultRepository

func NewResultRepository(db *gorm.DB) *ResultRepository {
	return resultspkg.NewResultRepository(db)
}
