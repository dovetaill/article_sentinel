package articleinspect

import rulespkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/rules"

var (
	ErrKeywordNotFound     = rulespkg.ErrKeywordNotFound
	ErrInvalidKeywordInput = rulespkg.ErrInvalidKeywordInput
)

type KeywordService = rulespkg.KeywordService

func NewKeywordService(repo *KeywordRepository) *KeywordService {
	return rulespkg.NewKeywordService(repo)
}
