package articleinspect

import sharedpkg "github.com/dovetaill/article-sentinel/internal/modules/articleinspect/shared"

func normalizePage(page, pageSize int) (int, int) {
	return sharedpkg.NormalizePage(page, pageSize)
}
