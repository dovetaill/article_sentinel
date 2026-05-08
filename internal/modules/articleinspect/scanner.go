package articleinspect

import scanpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/scan"

type CandidateArticle = scanpkg.CandidateArticle

type KeywordRule = scanpkg.KeywordRule

type Hit = scanpkg.Hit

type Scanner = scanpkg.Scanner

type KeywordScanner = scanpkg.KeywordScanner

func NewKeywordScanner() *KeywordScanner {
	return scanpkg.NewKeywordScanner()
}
