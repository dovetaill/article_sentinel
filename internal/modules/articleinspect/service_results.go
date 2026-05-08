package articleinspect

import (
	resultspkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/results"
	"gorm.io/gorm"
)

var ErrInvalidResultQuery = resultspkg.ErrInvalidResultQuery

type ResultService = resultspkg.ResultService

func NewResultService(db *gorm.DB) *ResultService {
	return resultspkg.NewResultService(db)
}
